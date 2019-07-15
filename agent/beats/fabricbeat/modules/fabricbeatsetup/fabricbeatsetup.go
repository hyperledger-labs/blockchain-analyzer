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
	SDK                  *fabsdk.FabricSDK
	BlockIndexName       string
	TransactionIndexName string
	KeyIndexName         string
	DashboardDirectory   string
	TemplateDirectory    string
}

// Initialize reads the configuration file and sets up FabricSetup
func (setup *FabricbeatSetup) Initialize() error {

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
	setup.SDK = sdk
	logp.Info("SDK created")

	mspContext := setup.SDK.Context(fabsdk.WithOrg(setup.OrgName))
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
	setup.MspClient, err1 = msp.New(mspContext)
	if err1 != nil {
		return err1
	}

	id, err2 := setup.MspClient.CreateSigningIdentity(providerMSP.WithCert(cert), providerMSP.WithPrivateKey(pk))
	if err2 != nil {
		return err2
	}

	setup.AdminIdentity = id
	resContext := setup.SDK.Context(fabsdk.WithIdentity(id))
	var err7 error
	setup.ResClient, err7 = resmgmt.New(resContext)
	if err7 != nil {
		return errors.WithMessage(err7, "failed to create new resmgmt client")
	}
	logp.Info("Resmgmt client created")

	logp.Info("Initialization Successful")
	setup.initialized = true
	return nil
}

// Closes SDK
func (setup *FabricbeatSetup) CloseSDK() {
	setup.SDK.Close()
}
