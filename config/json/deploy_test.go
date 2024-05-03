/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package json

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanSpecialChars(code []byte) string {
	space := regexp.MustCompile(`\s+`)
	return strings.ReplaceAll(space.ReplaceAllString(string(code), " "), " ", "")
}

func Test_ConfigDeploymentsSimple(t *testing.T) {
	b := []byte(`{
		"testnet": {
			"account-1": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"]
		},
		"emulator": {
			"account-2": ["KittyItems", "KittyItemsMarket"],
			"account-3": ["FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"]
		}
	}`)

	var jsonDeployments jsonDeployments
	err := json.Unmarshal(b, &jsonDeployments)
	assert.NoError(t, err)

	deployments, err := jsonDeployments.transformToConfig()
	assert.NoError(t, err)

	const account1Name = "account-1"
	const account2Name = "account-2"
	const account3Name = "account-3"

	assert.Len(t, deployments.ByNetwork("testnet"), 1)
	assert.Len(t, deployments.ByNetwork("emulator"), 2)

	account1Deployment := deployments.ByAccountAndNetwork(account1Name, "testnet")
	account2Deployment := deployments.ByAccountAndNetwork(account2Name, "emulator")
	account3Deployment := deployments.ByAccountAndNetwork(account3Name, "emulator")

	require.NotNil(t, account1Deployment)
	require.NotNil(t, account2Deployment)
	require.NotNil(t, account3Deployment)

	assert.Equal(t, account1Name, account1Deployment.Account)
	assert.Equal(t, account2Name, account2Deployment.Account)
	assert.Equal(t, account3Name, account3Deployment.Account)

	assert.Len(t, account1Deployment.Contracts, 4)

	for i, name := range []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems"} {
		assert.Equal(t, account1Deployment.Contracts[i].Name, name)
	}

	for i, name := range []string{"KittyItems", "KittyItemsMarket"} {
		assert.Equal(t, account2Deployment.Contracts[i].Name, name)
	}

	for i, name := range []string{"FungibleToken", "NonFungibleToken", "Kibble", "KittyItems", "KittyItemsMarket"} {
		assert.Equal(t, account3Deployment.Contracts[i].Name, name)
	}

}

func Test_TransformDeployToJSON(t *testing.T) {
	b := []byte(`{
		"emulator":{
			"account-3":["KittyItems", {
					"name": "Kibble",
					"args": [
						{ "type": "String", "value": "Hello World" },
						{ "type": "Int8", "value": "10" }
					]
			}],
			"account-4":["FungibleToken","NonFungibleToken","Kibble","KittyItems","KittyItemsMarket"]
		},
		"testnet":{
			"account-2":["FungibleToken","NonFungibleToken","Kibble","KittyItems"]
		}
	}`)

	var original jsonDeployments
	err := json.Unmarshal(b, &original)
	assert.NoError(t, err)

	deployments, err := original.transformToConfig()
	assert.NoError(t, err)

	j := transformDeploymentsToJSON(deployments)
	x, _ := json.Marshal(j)

	// Unmarshal the config again to compare against the original
	var result jsonDeployments
	err = json.Unmarshal(x, &result)
	assert.NoError(t, err)

	// Check that result is same as original after transformation
	assert.Equal(t, original, result)
}

func Test_DeploymentAdvanced(t *testing.T) {
	b := []byte(`{
		"emulator": {
			"alice": [
				{
					"name": "Kibble",
					"args": [
						{ "type": "String", "value": "Hello World" },
						{ "type": "Int8", "value": "10" },
						{ "type": "Bool", "value": false }
					]
				},
				"KittyItemsMarket"
			]
		}
	}`)

	var jsonDeployments jsonDeployments
	err := json.Unmarshal(b, &jsonDeployments)
	assert.NoError(t, err)

	deployments, err := jsonDeployments.transformToConfig()
	assert.NoError(t, err)

	alice := deployments.ByAccountAndNetwork("alice", "emulator")
	assert.NotNil(t, alice)
	assert.Len(t, alice.Contracts, 2)
	assert.Equal(t, "Kibble", alice.Contracts[0].Name)
	assert.Len(t, alice.Contracts[0].Args, 3)
	assert.Equal(t, `"Hello World"`, alice.Contracts[0].Args[0].String())
	assert.Equal(t, "10", alice.Contracts[0].Args[1].String())
	assert.Equal(t, "Bool", alice.Contracts[0].Args[2].Type().ID())
	assert.False(t, alice.Contracts[0].Args[2].ToGoValue().(bool))
	assert.Equal(t, "KittyItemsMarket", alice.Contracts[1].Name)
	assert.Len(t, alice.Contracts[1].Args, 0)
}
