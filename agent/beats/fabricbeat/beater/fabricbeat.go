// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	fabricbeatConfig "github.com/balazsprehoda/fabricbeat/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

// Fabricbeat configuration.
type Fabricbeat struct {
	done   chan struct{}
	config fabricbeatConfig.Config
	client beat.Client
}

// New creates an instance of fabricbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := fabricbeatConfig.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Fabricbeat{
		done:   make(chan struct{}),
		config: c,
	}

	fSetup := FabricSetup{
		OrgName:    bt.config.Organization,
		ConfigFile: bt.config.ConnectionProfile,
		ChannelID:  bt.config.Channel,
	}

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		logp.Error(err)
		return nil, err
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
}

// FabricSetup implementation
type FabricSetup struct {
	ConfigFile  string
	ChannelID   string
	initialized bool
	OrgName     string
	sdk         *fabsdk.FabricSDK
}

// Initialize reads the configuration file and sets up the client, chain and event hub
func (setup *FabricSetup) Initialize() error {

	logp.Info("Initializing SDK")
	// Add parameters for the initialization
	if setup.initialized {
		return errors.New("sdk already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err0 := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err0 != nil {
		logp.Warn("SDK initialization failed!")
		return errors.WithMessage(err0, "failed to create SDK")
	}
	setup.sdk = sdk
	logp.Info("SDK created")

	// The MSP client allow us to retrieve user information from their identity, like its signing identity
	mspClient, err1 := mspclient.New(sdk.Context(), mspclient.WithOrg(setup.OrgName))
	if err1 != nil {
		return errors.WithMessage(err1, "failed to create MSP client")
	}

	enrollmentSecret, err2 := mspClient.Register(&msp.RegistrationRequest{Name: "fabricbeatUser"})
	if err2 != nil {
		logp.Info("Register returned error %s\n", err2)
		return err2
	}
	logp.Info("Successfully registered user")

	err3 := mspClient.Enroll("fabricbeatUser", msp.WithSecret(enrollmentSecret))
	if err3 != nil {
		logp.Warn("Failed to enroll user\n")
		logp.Error(err3)
		return err3
	}
	logp.Info("Successfully enrolled user")

	adminIdentity, err4 := mspClient.GetSigningIdentity("fabricbeatUser")
	if err4 != nil {
		logp.Warn("User not found %s\n", err4)
		logp.Error(err4)
		return err3
	}

	if adminIdentity.Identifier().ID != "fabricbeatUser" {
		logp.Warn("Enrolled user name doesn't match")
		return nil
	}

	logp.Info("Initialization Successful")
	setup.initialized = true
	return nil
}

func (setup *FabricSetup) CloseSDK() {
	setup.sdk.Close()
}
