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

package config

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestAliases_Add(t *testing.T) {
	aliases := Aliases{}
	aliases.Add("testnet", flow.HexToAddress("0xabcdef"))

	alias := aliases.ByNetwork("testnet")
	assert.NotNil(t, alias)
}

func TestAliases_Add_Duplicate(t *testing.T) {
	aliases := Aliases{}
	aliases.Add("testnet", flow.HexToAddress("0xabcdef"))
	aliases.Add("testnet", flow.HexToAddress("0x123456"))

	assert.Len(t, aliases, 1)
}

func TestContracts_AddOrUpdate_Add(t *testing.T) {
	contracts := Contracts{}
	contracts.AddOrUpdate(Contract{Name: "mycontract", Location: "path/to/contract.cdc"})

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("mycontract")
	assert.NoError(t, err)
	assert.Equal(t, "path/to/contract.cdc", contract.Location)
}

func TestContracts_AddOrUpdate_Update(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract", Location: "path/to/contract.cdc"},
	}
	contracts.AddOrUpdate(Contract{Name: "mycontract", Location: "new/path/to/contract.cdc"})

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("mycontract")
	assert.NoError(t, err)
	assert.Equal(t, "new/path/to/contract.cdc", contract.Location)
}

func TestContracts_Remove(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract", Location: "path/to/contract.cdc"},
	}
	err := contracts.Remove("mycontract")
	assert.NoError(t, err)
	assert.Len(t, contracts, 0)
}

func TestContracts_Remove_NotFound(t *testing.T) {
	contracts := Contracts{
		Contract{Name: "mycontract1", Location: "path/to/contract.cdc"},
		Contract{Name: "mycontract2", Location: "path/to/contract.cdc"},
		Contract{Name: "mycontract3", Location: "path/to/contract.cdc"},
	}
	err := contracts.Remove("mycontract2")
	assert.NoError(t, err)
	assert.Len(t, contracts, 2)
	_, err = contracts.ByName("mycontract1")
	assert.NoError(t, err)
	_, err = contracts.ByName("mycontract3")
	assert.NoError(t, err)
	_, err = contracts.ByName("mycontract2")
	assert.EqualError(t, err, "contract mycontract2 does not exist")
}

func TestContracts_AddDependencyAsContract(t *testing.T) {
	contracts := Contracts{}
	contracts.AddDependencyAsContract(Dependency{
		Name: "testcontract",
		Source: Source{
			NetworkName:  "testnet",
			Address:      flow.HexToAddress("0x0000000000abcdef"),
			ContractName: "TestContract",
		},
	}, "testnet")

	assert.Len(t, contracts, 1)

	contract, err := contracts.ByName("testcontract")
	assert.NoError(t, err)
	assert.Equal(t, "imports/0000000000abcdef/TestContract.cdc", contract.Location)
	assert.Len(t, contract.Aliases, 1)
}

func TestContract_IsAlias(t *testing.T) {
	tests := []struct {
		name     string
		contract Contract
		expected bool
	}{
		{
			name:     "contract with canonical is an alias",
			contract: Contract{Name: "FUSD1", Canonical: "FUSD"},
			expected: true,
		},
		{
			name:     "contract without canonical is not an alias",
			contract: Contract{Name: "FUSD"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.contract.IsAlias())
		})
	}
}

func TestContract_CanonicalName(t *testing.T) {
	tests := []struct {
		name     string
		contract Contract
		expected string
	}{
		{
			name:     "alias returns canonical name",
			contract: Contract{Name: "FUSD1", Canonical: "FUSD"},
			expected: "FUSD",
		},
		{
			name:     "non-alias returns its own name",
			contract: Contract{Name: "FUSD"},
			expected: "FUSD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.contract.CanonicalName())
		})
	}
}

func TestContracts_ValidateCanonical(t *testing.T) {
	tests := []struct {
		name      string
		contracts Contracts
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid canonical reference",
			contracts: Contracts{
				{Name: "FUSD", Location: "FUSD.cdc"},
				{Name: "FUSD1", Location: "FUSD.cdc", Canonical: "FUSD"},
			},
			wantErr: false,
		},
		{
			name: "self-referential canonical",
			contracts: Contracts{
				{Name: "FUSD", Location: "FUSD.cdc", Canonical: "FUSD"},
			},
			wantErr: true,
			errMsg:  "contract FUSD cannot have itself as canonical",
		},
		{
			name: "multiple aliases to same canonical",
			contracts: Contracts{
				{Name: "FUSD", Location: "FUSD.cdc"},
				{Name: "FUSD1", Location: "FUSD.cdc", Canonical: "FUSD"},
				{Name: "FUSD2", Location: "FUSD.cdc", Canonical: "FUSD"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contracts.ValidateCanonical()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContracts_GetAliases(t *testing.T) {
	contracts := Contracts{
		{Name: "FUSD", Location: "FUSD.cdc"},
		{Name: "FUSD1", Location: "FUSD.cdc", Canonical: "FUSD"},
		{Name: "FUSD2", Location: "FUSD.cdc", Canonical: "FUSD"},
		{Name: "FT", Location: "FT.cdc"},
		{Name: "FT1", Location: "FT.cdc", Canonical: "FT"},
	}

	fusdAliases := contracts.GetAliases("FUSD")
	assert.Len(t, fusdAliases, 2)
	assert.Equal(t, "FUSD1", fusdAliases[0].Name)
	assert.Equal(t, "FUSD2", fusdAliases[1].Name)

	ftAliases := contracts.GetAliases("FT")
	assert.Len(t, ftAliases, 1)
	assert.Equal(t, "FT1", ftAliases[0].Name)

	noAliases := contracts.GetAliases("NonExistent")
	assert.Len(t, noAliases, 0)
}
