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

	"github.com/stretchr/testify/assert"
)

func Test_ConfigDependencies(t *testing.T) {
	b := []byte(`{
		"HelloWorld": "testnet://877931736ee77cff.HelloWorld"
		}`)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, dependencies, 1)

	dependencyOne := dependencies.ByName("HelloWorld")
	assert.NotNil(t, dependencyOne)

	assert.NotNil(t, dependencyOne)
}

func Test_TransformDependenciesToJSON(t *testing.T) {
	b := []byte(`{
		"HelloWorld": "testnet://877931736ee77cff.HelloWorld"
	}`)

	bOut := []byte(`{
		"HelloWorld": {
			"source": "testnet://877931736ee77cff.HelloWorld",
			"hash": "",
			"aliases": {}
		}
	}`)

	var jsonContracts jsonContracts
	errContracts := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, errContracts)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)
	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	j := transformDependenciesToJSON(dependencies, contracts)
	x, _ := json.Marshal(j)

	assert.Equal(t, cleanSpecialChars(bOut), cleanSpecialChars(x))
}

func Test_ConfigDependenciesWithCanonical(t *testing.T) {
	b := []byte(`{
		"NumberFormatter": {
			"source": "testnet://8a4dce54554b225d.NumberFormatter",
			"hash": "dc7043832da46dbcc8242a53fa95b37f020bc374df42586a62703b2651979fb9",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7",
				"testnet": "8a4dce54554b225d"
			}
		},
		"NumberFormatterAlias": {
			"source": "testnet://8a4dce54554b225d.NumberFormatter",
			"hash": "dc7043832da46dbcc8242a53fa95b37f020bc374df42586a62703b2651979fb9",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7",
				"testnet": "8a4dce54554b225d"
			},
			"canonical": "NumberFormatter"
		}
	}`)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	assert.Len(t, dependencies, 2)

	canonicalDep := dependencies.ByName("NumberFormatter")
	assert.NotNil(t, canonicalDep)
	assert.Equal(t, "", canonicalDep.Canonical)

	aliasDep := dependencies.ByName("NumberFormatterAlias")
	assert.NotNil(t, aliasDep)
	assert.Equal(t, "NumberFormatter", aliasDep.Canonical)
}

func Test_TransformDependenciesWithCanonicalToJSON(t *testing.T) {
	b := []byte(`{
		"NumberFormatter": {
			"source": "testnet://8a4dce54554b225d.NumberFormatter",
			"hash": "dc7043832da46dbcc8242a53fa95b37f020bc374df42586a62703b2651979fb9",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7",
				"testnet": "8a4dce54554b225d"
			}
		},
		"NumberFormatterAlias": {
			"source": "testnet://8a4dce54554b225d.NumberFormatter",
			"hash": "dc7043832da46dbcc8242a53fa95b37f020bc374df42586a62703b2651979fb9",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7",
				"testnet": "8a4dce54554b225d"
			},
			"canonical": "NumberFormatter"
		}
	}`)

	var jsonContracts jsonContracts
	errContracts := json.Unmarshal(b, &jsonContracts)
	assert.NoError(t, errContracts)

	var jsonDependencies jsonDependencies
	err := json.Unmarshal(b, &jsonDependencies)
	assert.NoError(t, err)

	contracts, err := jsonContracts.transformToConfig()
	assert.NoError(t, err)
	dependencies, err := jsonDependencies.transformToConfig()
	assert.NoError(t, err)

	j := transformDependenciesToJSON(dependencies, contracts)

	// Marshal and check
	x, _ := json.Marshal(j)

	// Parse back and check canonical field
	var result map[string]jsonDependency
	err = json.Unmarshal(x, &result)
	assert.NoError(t, err)

	assert.Equal(t, "", result["NumberFormatter"].Extended.Canonical)
	assert.Equal(t, "NumberFormatter", result["NumberFormatterAlias"].Extended.Canonical)
}
