package beater

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	fabricbeatConfig "github.com/balazsprehoda/fabricbeat/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	providerMSP "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

// Fabricbeat configuration.
type Fabricbeat struct {
	done          chan struct{}
	config        fabricbeatConfig.Config
	client        beat.Client
	fSetup        *FabricSetup
	lastBlockNums map[*ledger.Client]uint64
}

// New creates an instance of fabricbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := fabricbeatConfig.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Fabricbeat{
		done:          make(chan struct{}),
		config:        c,
		lastBlockNums: make(map[*ledger.Client]uint64),
	}

	fSetup := &FabricSetup{
		OrgName:       bt.config.Organization,
		ConfigFile:    bt.config.ConnectionProfile,
		Peer:          bt.config.Peer,
		AdminCertPath: bt.config.AdminCertPath,
		AdminKeyPath:  bt.config.AdminKeyPath,
	}

	// Initialization of the Fabric SDK from the previously set properties
	err1 := fSetup.Initialize()
	if err1 != nil {
		logp.Error(err1)
		return nil, err1
	}

	bt.fSetup = fSetup

	// Get all channels the peer is part of
	channelsResponse, err2 := fSetup.resClient.QueryChannels(resmgmt.WithTargetEndpoints(fSetup.Peer))
	if err2 != nil {
		return nil, err2
	}

	// Initialize the ledger client for each channel
	for _, channel := range channelsResponse.Channels {
		logp.Info(channel.ChannelId)
		channelContext := fSetup.sdk.ChannelContext(channel.ChannelId, fabsdk.WithIdentity(fSetup.adminIdentity))
		if channelContext == nil {
			logp.Warn("Channel context creation failed, ChannelContext() returned nil for channel " + channel.ChannelId)
		}
		ledgerClient, err4 := ledger.New(channelContext)
		if err4 != nil {
			return nil, err4
		}
		fSetup.ledgerClients = append(fSetup.ledgerClients, ledgerClient)
		logp.Info("Ledger client initialized for channel " + channel.ChannelId)
	}
	logp.Info("Channel clients initialized")

	// Get installed chaincodes of the peer
	chaincodeResponse, err3 := fSetup.resClient.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints(fSetup.Peer))
	if err3 != nil {
		return nil, err3
	}
	for _, chaincode := range chaincodeResponse.Chaincodes {
		logp.Info(chaincode.Name)
	}

	return bt, nil
}

// Run starts fabricbeat.
func (bt *Fabricbeat) Run(b *beat.Beat) error {
	logp.Info("fabricbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	counter := 1

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		logp.Info("Start event loop")
		for _, channelClient := range bt.fSetup.ledgerClients {
			infoResponse, err1 := channelClient.QueryInfo()
			if err1 != nil {
				logp.Warn("QueryInfo returned error: " + err1.Error())
			}
			queriedBlockNum := infoResponse.BCI.Height
			logp.Info("Block height: %d", queriedBlockNum)
			for bt.lastBlockNums[channelClient] < queriedBlockNum {
				blockResponse, blockError := channelClient.QueryBlock(bt.lastBlockNums[channelClient])
				if blockError != nil {
					logp.Warn("QueryBlock returned error: " + blockError.Error())
				}
				logp.Info("Block queried: " + blockResponse.String())
				bt.lastBlockNums[channelClient]++
			}
		}

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"counter": counter,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
		counter++
	}
}

// Stop stops fabricbeat.
func (bt *Fabricbeat) Stop() {
	bt.client.Close()
	close(bt.done)
	defer bt.fSetup.CloseSDK()
}

// FabricSetup implementation
type FabricSetup struct {
	ConfigFile    string
	initialized   bool
	OrgName       string
	Peer          string
	mspClient     *msp.Client
	AdminCertPath string
	AdminKeyPath  string
	adminIdentity providerMSP.SigningIdentity
	resClient     *resmgmt.Client
	ledgerClients []*ledger.Client
	sdk           *fabsdk.FabricSDK
}

// Initialize reads the configuration file and sets up FabricSetup
func (setup *FabricSetup) Initialize() error {

	logp.Info("Initializing SDK")
	// Add parameters for the initialization
	if setup.initialized {
		return errors.New("SDK already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err0 := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err0 != nil {
		logp.Warn("SDK initialization failed!")
		return errors.WithMessage(err0, "failed to create SDK")
	}
	setup.sdk = sdk
	logp.Info("SDK created")

	mspContext := setup.sdk.Context(fabsdk.WithOrg(setup.OrgName))
	if mspContext == nil {
		logp.Warn("setup.sdk.Context() returned nil")
	}

	cert, err := ioutil.ReadFile(setup.AdminCertPath)
	if err != nil {
		return err
	}
	pk, err := ioutil.ReadFile(setup.AdminKeyPath)
	if err != nil {
		return err
	}

	var err1 error
	setup.mspClient, err1 = msp.New(mspContext)
	if err1 != nil {
		return err1
	}

	id, err2 := setup.mspClient.CreateSigningIdentity(providerMSP.WithCert(cert), providerMSP.WithPrivateKey(pk))
	if err2 != nil {
		return err2
	}

	setup.adminIdentity = id
	resContext := setup.sdk.Context(fabsdk.WithIdentity(id))
	var err7 error
	setup.resClient, err7 = resmgmt.New(resContext)
	if err7 != nil {
		return errors.WithMessage(err7, "failed to create new resmgmt client")
	}
	logp.Info("Resmgmt client created")

	logp.Info("Initialization Successful")
	setup.initialized = true
	return nil
}

// Closes SDK
func (setup *FabricSetup) CloseSDK() {
	setup.sdk.Close()
}
