package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/elastic/beats/libbeat/logp"
)

// This struct is for parsing block filter query response from Elasticsearch
type BlockIndexFilterResponse struct {
	BlockIndexFilterHitsObject struct {
		BlockIndexFilterHit []struct {
			BlockIndexData struct {
				BlockHash string `json:"block_hash"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// This struct is for getting the block number from the block number response.
type BlockNumber struct {
	BlockNumber uint64 `json:"blockNumber"`
}

// This struct is for parsing the block number response from Elasticsearch.
type BlockNumberResponse struct {
	Index       string      `json:"_index"`
	Type        string      `json:"_type"`
	Id          string      `json:"_id"`
	BlockNumber BlockNumber `json:"_source"`
}

// Returns the hash of the specified block from the specified index.
func GetBlockHash(elasticURL, blockIndexName, organization, peerName string, blockNumber uint64) (string, error) {
	// Get the index from which we want to get the last known block
	resp, err := http.Get(fmt.Sprintf(elasticURL+"/_cat/indices/fabricbeat-*%s*%s*", blockIndexName, organization))
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// [2] is the name of the index
	blockIndex := strings.Fields(string(body))[2]

	// Retrieve the last known block from Elasticsearch
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/%s/_search", elasticURL, blockIndex)
	requestBody := fmt.Sprintf(`{
		"size": 1,
		"query": {
		  "bool": {
			"filter": [
			  {
				"term": {
				  "block_number": {
					"value": "%d"
				  }
				}
			  },
			  {
				"term": {
				  "peer": {
					"value": "%s"
				  }
				}
			  }
			]
		  }
		},
		"sort": [
		  {
			"value": {
			  "order": "desc"
			}
		  }
		]
	}`, blockNumber, peerName)
	logp.Debug("URL for last block query: ", url)
	request, err := http.NewRequest("GET", url, bytes.NewBufferString(requestBody))
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "application/json")
	resp, err = httpClient.Do(request)
	if err != nil {
		return "", err
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("Failed to get last block from Elasticsearch: " + string(body))
	}
	var lastBlockResponseFromElastic BlockIndexFilterResponse
	err = json.Unmarshal(body, &lastBlockResponseFromElastic)
	if err != nil {
		return "", err
	}
	fmt.Println(string(body))
	if lastBlockResponseFromElastic.BlockIndexFilterHitsObject.BlockIndexFilterHit == nil {
		return "", errors.New("Could not properly unmarshal the response body to BlockIndexFilterResponse: BlockIndexFilterResponse.BlockIndexFilterHitsObject.BlockIndexFilterHit is nil")
	}

	blockHashFromElastic := lastBlockResponseFromElastic.BlockIndexFilterHitsObject.BlockIndexFilterHit[0].BlockIndexData.BlockHash
	return blockHashFromElastic, nil
}

// Sends a GET Http request to the sepcified URL, parses the response and returns the block number.
func GetBlockNumber(url string) (uint64, error) {
	var lastBlockNumber BlockNumber
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		return 0, errors.New(fmt.Sprintf("Failed getting the last block number! Http response status code: %d", resp.StatusCode))
	}
	if resp.StatusCode == 404 {
		// It is the very first start of the agent, there is no last block yet.
		lastBlockNumber.BlockNumber = 0
		logp.Info("Last known block number not found, setting to 0")
	} else {
		// Get the block number info from the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		var lastBlockNumberResponse BlockNumberResponse
		err = json.Unmarshal(body, &lastBlockNumberResponse)
		if err != nil {
			return 0, err
		}
		lastBlockNumber = lastBlockNumberResponse.BlockNumber
	}
	return lastBlockNumber.BlockNumber, nil
}

func SendBlockNumber(url string, lastBlockNumber BlockNumber) error {
	jsonBlockNumber, err := json.Marshal(lastBlockNumber)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBlockNumber))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New("Sending last block number to Elasticsearch failed!")
	}
	return nil
}
