package fabricbeatsetup

import (
	"io/ioutil"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	providerMSP "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"

	fabricConfig "github.com/blockchain-analyzer/agent/fabricbeat/config"
)

// Fabric, Elasticsearch and Kibana specific setup
type FabricbeatSetup struct {
	ConfigFile           string
	initialized          bool
	OrgName              string
	Peer                 string
	ElasticURL           string
	KibanaURL            string
	MspClient            *msp.Client
	AdminCertPath        string
	AdminKeyPath         string
	AdminIdentity        providerMSP.SigningIdentity
	ResClient            *resmgmt.Client
	LedgerClients        []*ledger.Client
	Channels             map[*ledger.Client]string
	SDK                  *fabsdk.FabricSDK
	BlockIndexName       string
	TransactionIndexName string
	KeyIndexName         string
	DashboardDirectory   string
	TemplateDirectory    string
	Chaincodes           []fabricConfig.Chaincode
}

// Initialize reads the configuration file and sets up FabricSetup
func (setup *FabricbeatSetup) Initialize() error {

	logp.Info("Initializing SDK")
	// Add parameters for the initialization
	if setup.initialized {
		return errors.New("SDK already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err != nil {
		logp.Warn("SDK initialization failed!")
		return errors.WithMessage(err, "failed to create SDK")
	}
	setup.SDK = sdk
	logp.Info("SDK created")

	mspContext := setup.SDK.Context(fabsdk.WithOrg(setup.OrgName))
	if mspContext == nil {
		logp.Warn("setup.sdk.Context() returned nil")
	}

	// Retrieving admin identity (key + cert)
	cert, err := ioutil.ReadFile(setup.AdminCertPath)
	if err != nil {
		return err
	}
	pk, err := ioutil.ReadFile(setup.AdminKeyPath)
	if err != nil {
		return err
	}

	// Setting up msp client
	setup.MspClient, err = msp.New(mspContext)
	if err != nil {
		return err
	}

	// Creating signing identity
	id, err := setup.MspClient.CreateSigningIdentity(providerMSP.WithCert(cert), providerMSP.WithPrivateKey(pk))
	if err != nil {
		return err
	}

	setup.AdminIdentity = id
	resContext := setup.SDK.Context(fabsdk.WithIdentity(id))
	setup.ResClient, err = resmgmt.New(resContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create new resmgmt client")
	}
	logp.Info("Resmgmt client created")

	// Get all channels the peer is part of
	channelsResponse, err := setup.ResClient.QueryChannels(resmgmt.WithTargetEndpoints(setup.Peer))
	if err != nil {
		return err
	}

	// Initialize the ledger client for each channel
	setup.Channels = make(map[*ledger.Client]string)
	for _, channel := range channelsResponse.Channels {
		channelContext := setup.SDK.ChannelContext(channel.ChannelId, fabsdk.WithIdentity(setup.AdminIdentity))
		if channelContext == nil {
			logp.Warn("Channel context creation failed, ChannelContext() returned nil for channel " + channel.ChannelId)
		}
		ledgerClient, err := ledger.New(channelContext)
		if err != nil {
			return err
		}
		setup.LedgerClients = append(setup.LedgerClients, ledgerClient)
		setup.Channels[ledgerClient] = channel.ChannelId
		logp.Info("Ledger client initialized for channel " + channel.ChannelId)
	}
	logp.Info("Channel clients initialized")

	// Get installed chaincodes of the peer
	chaincodeResponse, err := setup.ResClient.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints(setup.Peer))
	if err != nil {
		return err
	}
	for _, chaincode := range chaincodeResponse.Chaincodes {
		logp.Info("Installed chaincode name: %s", chaincode.Name)
	}

	logp.Info("Initialization Successful")
	setup.initialized = true
	return nil
}

// Closes SDK
func (setup *FabricbeatSetup) CloseSDK() {
	setup.SDK.Close()
}
