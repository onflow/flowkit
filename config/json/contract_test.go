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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigContractsSimple(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": "./cadence/kittyItems/contracts/KittyItemsMarket.cdc"
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	contract, _ := contracts.ByName("KittyItems")
	assert.NotNil(t, contract)

	marketContract, _ := contracts.ByName("KittyItemsMarket")
	assert.NotNil(t, marketContract)

	assert.Equal(t, filepath.FromSlash("./cadence/kittyItems/contracts/KittyItems.cdc"), contract.Location)
	assert.Equal(t, filepath.FromSlash("./cadence/kittyItems/contracts/KittyItemsMarket.cdc"), marketContract.Location)
}

func Test_ConfigContractsComplex(t *testing.T) {
	b := []byte(`{
    "KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
    "KittyItemsMarket": {
		"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
		"aliases": {
			"testnet": "f8d6e0586b0a20c7"
		}
    }
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, contracts, 2)

	kittyItems, _ := contracts.ByName("KittyItems")
	assert.NotNil(t, kittyItems)

	kittyItemsMarket, _ := contracts.ByName("KittyItemsMarket")
	assert.NotNil(t, kittyItemsMarket)

	assert.Equal(t, filepath.FromSlash("./cadence/kittyItems/contracts/KittyItems.cdc"), kittyItems.Location)
	assert.Equal(t, filepath.FromSlash("./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc"), kittyItemsMarket.Location)

	assert.True(t, kittyItemsMarket.Aliases.ByNetwork("emulator") == nil)

	assert.True(t, kittyItemsMarket.IsAliased())
	assert.True(t, kittyItemsMarket.Aliases.ByNetwork("emulator") == nil)
	assert.Equal(t, "f8d6e0586b0a20c7", kittyItemsMarket.Aliases.ByNetwork("testnet").Address.String())

	assert.False(t, kittyItems.IsAliased())
	assert.True(t, kittyItems.Aliases.ByNetwork("emulator") == nil)
}

func Test_ConfigContractsAliases(t *testing.T) {
	b := []byte(`{
		"NonFungibleToken": "../hungry-kitties/cadence/contracts/NonFungibleToken.cdc",
		"Kibble": {
			"source": "../hungry-kitties/cadence/contracts/Kibble.cdc",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7",
				"testnet": "ead892083b3e2c6c"
			}
		},
		"FungibleToken": {
			"source": "../hungry-kitties/cadence/contracts/FungibleToken.cdc",
			"aliases": {
				"emulator": "e5a8b7f23e8b548f"
			}
		}
  }`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	fungibleToken, _ := contracts.ByName("FungibleToken")
	assert.NotNil(t, fungibleToken)
	assert.True(t, fungibleToken.IsAliased())
	assert.Equal(t, "e5a8b7f23e8b548f", fungibleToken.Aliases.ByNetwork("emulator").Address.String())
	assert.Equal(t, filepath.FromSlash("../hungry-kitties/cadence/contracts/FungibleToken.cdc"), fungibleToken.Location)

	kibble, _ := contracts.ByName("Kibble")
	assert.NotNil(t, kibble)
	assert.True(t, kibble.IsAliased())
	assert.Equal(t, filepath.FromSlash("../hungry-kitties/cadence/contracts/Kibble.cdc"), kibble.Location)
	assert.Equal(t, "ead892083b3e2c6c", kibble.Aliases.ByNetwork("testnet").Address.String())
	assert.Equal(t, "f8d6e0586b0a20c7", kibble.Aliases.ByNetwork("emulator").Address.String())

	nft, _ := contracts.ByName("NonFungibleToken")
	assert.NotNil(t, nft)
	assert.False(t, nft.IsAliased())
	assert.Equal(t, nft.Location, filepath.FromSlash("../hungry-kitties/cadence/contracts/NonFungibleToken.cdc"))
}

func Test_TransformContractToJSON(t *testing.T) {
	b := []byte(`{
		"KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
		"KittyItemsMarket": {
			"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			"aliases": {
				"testnet":"e5a8b7f23e8b548f"
			}
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	j := transformContractsToJSON(contracts)
	x, _ := json.Marshal(j)

	assert.JSONEq(t, string(b), string(x))
}

func Test_TransformComplexContractToJSON(t *testing.T) {
	b := []byte(`{
		"KittyItems": "./cadence/kittyItems/contracts/KittyItems.cdc",
		"KittyItemsMarket": {
			"source": "./cadence/kittyItemsMarket/contracts/KittyItemsMarket.cdc",
			"aliases": {
				"testnet":"e5a8b7f23e8b548f"
			}
		},
		"Kibble": {
			"source": "./cadence/kittyItems/contracts/KittyItems.cdc",
			"aliases": {
				"testnet": "e5a8b7f23e8b548f",
				"emulator": "f8d6e0586b0a20c7"
			}
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	j := transformContractsToJSON(contracts)
	x, _ := json.Marshal(j)

	assert.JSONEq(t, string(b), string(x))
}

func Test_ConfigContractsWithCanonical(t *testing.T) {
	b := []byte(`{
		"FUSD": {
			"source": "./contracts/FUSD.cdc",
			"aliases": {
				"testnet": "9a0766d93b6608b7",
				"mainnet": "3c5959b568896393"
			}
		},
		"FUSD1": {
			"source": "./contracts/FUSD.cdc",
			"aliases": {
				"testnet": "e223d8a629e49c68",
				"mainnet": "8d0e87b65159ae63"
			},
			"canonical": "FUSD"
		},
		"FUSD2": {
			"source": "./contracts/FUSD.cdc",
			"aliases": {
				"testnet": "0f9df91c9121c460",
				"mainnet": "754a90d51a1c8e1b"
			},
			"canonical": "FUSD"
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, contracts, 3)

	fusd, _ := contracts.ByName("FUSD")
	assert.NotNil(t, fusd)
	assert.Equal(t, "", fusd.Canonical)
	assert.False(t, fusd.IsAlias())

	fusd1, _ := contracts.ByName("FUSD1")
	assert.NotNil(t, fusd1)
	assert.Equal(t, "FUSD", fusd1.Canonical)
	assert.True(t, fusd1.IsAlias())
	assert.Equal(t, "FUSD", fusd1.CanonicalName())

	fusd2, _ := contracts.ByName("FUSD2")
	assert.NotNil(t, fusd2)
	assert.Equal(t, "FUSD", fusd2.Canonical)
	assert.True(t, fusd2.IsAlias())
	assert.Equal(t, "FUSD", fusd2.CanonicalName())
}

func Test_TransformContractsWithCanonicalToJSON(t *testing.T) {
	b := []byte(`{
		"FUSD": {
			"source": "./contracts/FUSD.cdc",
			"aliases": {
				"testnet": "9a0766d93b6608b7",
				"mainnet": "3c5959b568896393"
			}
		},
		"FUSD1": {
			"source": "./contracts/FUSD.cdc",
			"aliases": {
				"testnet": "e223d8a629e49c68",
				"mainnet": "8d0e87b65159ae63"
			},
			"canonical": "FUSD"
		},
		"FUSD2": {
			"source": "./contracts/FUSD.cdc",
			"aliases": {
				"testnet": "0f9df91c9121c460",
				"mainnet": "754a90d51a1c8e1b"
			},
			"canonical": "FUSD"
		}
	}`)

	var jsonContracts jsonContracts
	err := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)

	j := transformContractsToJSON(contracts)

	x, _ := json.Marshal(j)

	var result map[string]jsonContract
	err = json.Unmarshal(x, &result)
	assert.NoError(t, err)

	assert.Equal(t, "", result["FUSD"].Advanced.Canonical)
	assert.Equal(t, "FUSD", result["FUSD1"].Advanced.Canonical)
	assert.Equal(t, "FUSD", result["FUSD2"].Advanced.Canonical)
}
