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
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flowkit/v2/config"
)

func keys() []crypto.PrivateKey {
	var keys []crypto.PrivateKey
	key, _ := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47")
	keys = append(keys, key)
	return keys
}

func Test_SimpleJSONConfig(t *testing.T) {
	b := []byte(`{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		},
		"contracts": {},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"deployments": {}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)

	assert.NoError(t, err)
	assert.Len(t, conf.Accounts, 1)
	assert.Equal(t, "emulator-account", conf.Accounts[0].Name)
	assert.Equal(t, "0x11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Key.PrivateKey.String())
	network, err := conf.Networks.ByName("emulator")
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:3569", network.Host)
}

func Test_NonExistingContractForDeployment(t *testing.T) {
	b := []byte(`{
		"contracts": {},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"deployments": {
			"emulator": {
				"emulator-account": ["FungibleToken"]
			}
		}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = conf.Validate()
	assert.Equal(t, "deployment contains nonexisting contract FungibleToken", err.Error())
}

func Test_NonExistingAccountForDeployment(t *testing.T) {
	b := []byte(`{
		"contracts": {
			"FungibleToken": "./test.cdc"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"deployments": {
			"emulator": {
				"test-1": ["FungibleToken"]
			}
		}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = conf.Validate()
	assert.Equal(t, "deployment contains nonexisting account test-1", err.Error())
}

func Test_NonExistingNetworkForDeployment(t *testing.T) {
	b := []byte(`{
		"contracts": {
			"FungibleToken": "./test.cdc"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"networks": {},
		"deployments": {
			"foo": {
				"test-1": ["FungibleToken"]
			}
		}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = conf.Validate()
	assert.Equal(t, "deployment contains nonexisting network foo", err.Error())
}

func Test_NonExistingAccountForEmulator(t *testing.T) {
	b := []byte(`{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		}
	}`)

	parser := NewParser()
	conf, err := parser.Deserialize(b)
	assert.NoError(t, err)

	err = conf.Validate()
	assert.Equal(t, "emulator default contains nonexisting service account emulator-account", err.Error())
}

// If config has default emulator values, it will not show up in flow.json
func Test_SerializeConfigToJsonEmulatorDefault(t *testing.T) {
	configJson := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		}
	}`)
	cfg := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           3569,
			ServiceAccount: "emulator-account",
		}},
		Contracts:   config.Contracts{},
		Deployments: config.Deployments{},
		Accounts: config.Accounts{{
			Name:    "emulator-account",
			Address: flow.ServiceAddress(flow.Emulator),
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[0],
			},
		}},
		Networks: config.Networks{{
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}
	parser := NewParser()
	conf, _ := parser.Serialize(&cfg)
	assert.JSONEq(t, string(configJson), string(conf))
}

func Test_SerializeConfigToJsonEmulatorNotDefault(t *testing.T) {
	configJson := []byte(`{
		"emulators": {
			"default": {
				"port": 6000,
				"serviceAccount": "emulator-account"
			}
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "dd72967fd2bd75234ae9037dd4694c1f00baad63a10c35172bf65fbb8ad74b47"
			}
		},
		"networks": {
			"emulator": "127.0.0.1.3569"
		}
	}`)
	config := config.Config{
		Emulators: config.Emulators{{
			Name:           "default",
			Port:           6000,
			ServiceAccount: "emulator-account",
		}},
		Accounts: config.Accounts{{
			Name:    "emulator-account",
			Address: flow.ServiceAddress(flow.Emulator),
			Key: config.AccountKey{
				Type:       config.KeyTypeHex,
				Index:      0,
				SigAlgo:    crypto.ECDSA_P256,
				HashAlgo:   crypto.SHA3_256,
				PrivateKey: keys()[0],
			},
		}},
		Networks: config.Networks{{
			Name: "emulator",
			Host: "127.0.0.1.3569",
		}},
	}
	parser := NewParser()
	conf, _ := parser.Serialize(&config)
	assert.JSONEq(t, string(configJson), string(conf))

}

// Test_RoundTripPreservesKeyOrder verifies that loading a flow.json and then
// re-serializing it preserves the user's chosen ordering of keys within each
// section, including nested deployments. This was the motivating bug for the
// orderedMap data structure.
func Test_RoundTripPreservesKeyOrder(t *testing.T) {
	input := []byte(`{
	"contracts": {
		"ZetaContract": "./z.cdc",
		"AlphaContract": "./a.cdc",
		"MuContract": "./m.cdc"
	},
	"dependencies": {
		"Burner": "testnet://9a0766d93b6608b7.Burner",
		"AAA": "testnet://9a0766d93b6608b7.AAA"
	},
	"networks": {
		"mainnet": "access.mainnet.nodes.onflow.org:9000",
		"testnet": "access.testnet.nodes.onflow.org:9000",
		"emulator": "127.0.0.1:3569"
	},
	"accounts": {
		"zoo-account": {
			"address": "f8d6e0586b0a20c7",
			"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
		},
		"alpha-account": {
			"address": "f8d6e0586b0a20c7",
			"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
		}
	},
	"deployments": {
		"testnet": {
			"zoo-account": ["ZetaContract"]
		},
		"emulator": {
			"alpha-account": ["MuContract", "AlphaContract"]
		}
	}
}`)

	parser := NewParser()
	conf, err := parser.Deserialize(input)
	assert.NoError(t, err)

	out, err := parser.Serialize(conf)
	assert.NoError(t, err)

	// Re-parse into the JSON layer directly to compare structural ordering
	// without depending on whitespace.
	var original, roundTripped jsonConfig
	assert.NoError(t, json.Unmarshal(input, &original))
	assert.NoError(t, json.Unmarshal(out, &roundTripped))

	assert.Equal(t,
		sectionKeys(original.Contracts.orderedMap),
		sectionKeys(roundTripped.Contracts.orderedMap),
		"contracts order changed")
	assert.Equal(t,
		sectionKeys(original.Dependencies.orderedMap),
		sectionKeys(roundTripped.Dependencies.orderedMap),
		"dependencies order changed")
	assert.Equal(t,
		sectionKeys(original.Networks.orderedMap),
		sectionKeys(roundTripped.Networks.orderedMap),
		"networks order changed")
	assert.Equal(t,
		sectionKeys(original.Accounts.orderedMap),
		sectionKeys(roundTripped.Accounts.orderedMap),
		"accounts order changed")
	assert.Equal(t,
		sectionKeys(original.Deployments.orderedMap),
		sectionKeys(roundTripped.Deployments.orderedMap),
		"deployments order changed")

	// Verify inner deployment account ordering is also preserved (this is the
	// nested orderedMap[[]deployment] case).
	originalEmu, _ := original.Deployments.Get("emulator")
	roundTrippedEmu, _ := roundTripped.Deployments.Get("emulator")
	assert.Equal(t,
		sectionKeys(originalEmu.orderedMap),
		sectionKeys(roundTrippedEmu.orderedMap),
		"deployments.emulator account order changed")
}

// Test_NewKeysAppendAtEnd verifies that when a new entry is added programmatically
// (e.g. a new dependency is fetched), it is appended at the end rather than
// inserted alphabetically.
func Test_NewKeysAppendAtEnd(t *testing.T) {
	input := []byte(`{
	"contracts": {
		"ZetaContract": "./z.cdc",
		"AlphaContract": "./a.cdc"
	},
	"networks": {
		"emulator": "127.0.0.1:3569"
	},
	"accounts": {
		"emulator-account": {
			"address": "f8d6e0586b0a20c7",
			"key": "11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
		}
	}
}`)

	parser := NewParser()
	conf, err := parser.Deserialize(input)
	assert.NoError(t, err)

	conf.Contracts.AddOrUpdate(config.Contract{
		Name:     "MuContract",
		Location: "./m.cdc",
	})

	out, err := parser.Serialize(conf)
	assert.NoError(t, err)

	var result jsonConfig
	assert.NoError(t, json.Unmarshal(out, &result))

	assert.Equal(t,
		[]string{"ZetaContract", "AlphaContract", "MuContract"},
		sectionKeys(result.Contracts.orderedMap),
		"new contract should be appended at the end, not inserted alphabetically")
}

func sectionKeys[V any](m orderedMap[V]) []string {
	keys := make([]string, 0, m.Len())
	for k := range m.All {
		keys = append(keys, k)
	}
	return keys
}
