package fabricsetup

import (
	"io/ioutil"

	"log"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	providerMSP "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"

)

type Chaincode struct {
	Name       string   //`chaincode:"name"`
	Linkingkey string   //`chaincode:"linkingKey"`
	Values     []string //`chaincode:"values"`
}

// Fabric, Elasticsearch and Kibana specific setup
type FabricSetup struct {
	ConfigFile           string
	initialized          bool
	OrgName              string
	Peer                 string
	MspClient            *msp.Client
	AdminCertPath        string
	AdminKeyPath         string
	AdminIdentity        providerMSP.SigningIdentity
	ResClient            *resmgmt.Client
	LedgerClients        []*ledger.Client
	Channels             map[*ledger.Client]string
	SDK                  *fabsdk.FabricSDK
	Chaincodes           []Chaincode
	ElasticURL           string
	KibanaURL            string
	DashboardDirectory   string
	TemplateDirectory    string
	BlockIndexName       string
	TransactionIndexName string
	KeyIndexName         string
}

// Initialize reads the configuration file and sets up FabricSetup
func (setup *FabricSetup) Initialize() error {

	log.Print("Initializing SDK")
	// Add parameters for the initialization
	if setup.initialized {
		return errors.New("SDK already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err != nil {
		log.Print("SDK initialization failed!")
		return errors.WithMessage(err, "failed to create SDK")
	}
	setup.SDK = sdk
	log.Print("SDK created")

	mspContext := setup.SDK.Context(fabsdk.WithOrg(setup.OrgName))
	if mspContext == nil {
		log.Print("setup.sdk.Context() returned nil")
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
	log.Print("Resmgmt client created")

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
			log.Print("Channel context creation failed, ChannelContext() returned nil for channel " + channel.ChannelId)
		}
		ledgerClient, err := ledger.New(channelContext)
		if err != nil {
			return err
		}
		setup.LedgerClients = append(setup.LedgerClients, ledgerClient)
		setup.Channels[ledgerClient] = channel.ChannelId
		log.Print("Ledger client initialized for channel " + channel.ChannelId)
	}
	log.Print("Channel clients initialized")

	// Get installed chaincodes of the peer
	chaincodeResponse, err := setup.ResClient.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints(setup.Peer))
	if err != nil {
		return err
	}
	for _, chaincode := range chaincodeResponse.Chaincodes {
		log.Print("Installed chaincode name: " + chaincode.Name)
	}

	log.Print("Initialization Successful")
	setup.initialized = true
	return nil
}

// Closes SDK
func (setup *FabricSetup) CloseSDK() {
	setup.SDK.Close()
}
