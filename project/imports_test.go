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

package project

import (
	"regexp"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanCode(code []byte) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(string(code), " ")
}

func TestResolver(t *testing.T) {

	t.Run("Resolve imports", func(t *testing.T) {
		contracts := []*Contract{
			NewContract("Kibble", "./tests/Kibble.cdc", nil, flow.HexToAddress("0x1"), "", nil),
			NewContract("FT", "./tests/FT.cdc", nil, flow.HexToAddress("0x2"), "", nil),
		}

		aliases := map[string]string{
			"./tests/NFT.cdc": flow.HexToAddress("0x4").String(),
		}

		paths := []string{
			"./tests/foo.cdc",
			"./scripts/bar/foo.cdc",
			"./scripts/bar/foo.cdc",
			"./tests/foo.cdc",
		}

		scripts := [][]byte{
			[]byte(`
			import Kibble from "./Kibble.cdc"
			import FT from "./FT.cdc"
			access(all) fun main() {}
    `), []byte(`
			import Kibble from "../../tests/Kibble.cdc"
			import FT from "../../tests/FT.cdc"
			access(all) fun main() {}
    `), []byte(`
			import Kibble from "../../tests/Kibble.cdc"
			import NFT from "../../tests/NFT.cdc"
			access(all) fun main() {}
    `), []byte(`
			import Kibble from "./Kibble.cdc"
			import Crypto
			import Foo from 0x0000000000000001
			access(all) fun main() {}
	`),
		}

		resolved := [][]byte{
			[]byte(`
			import Kibble from 0x0000000000000001
			import FT from 0x0000000000000002
			access(all) fun main() {}
    `), []byte(`
			import Kibble from 0x0000000000000001
			import FT from 0x0000000000000002
			access(all) fun main() {}
    `), []byte(`
			import Kibble from 0x0000000000000001
			import NFT from 0x0000000000000004
			access(all) fun main() {}
    `), []byte(`
			import Kibble from 0x0000000000000001
			import Crypto
			import Foo from 0x0000000000000001
			access(all) fun main() {}
	`),
		}

		replacer := NewImportReplacer(contracts, aliases)
		for i, script := range scripts {
			program, err := NewProgram(script, nil, paths[i])
			require.NoError(t, err)

			program, err = replacer.Replace(program)
			assert.NoError(t, err)
			assert.Equal(t, cleanCode(program.Code()), cleanCode(resolved[i]))
		}
	})

	t.Run("Resolve new schema", func(t *testing.T) {
		contracts := []*Contract{
			NewContract("Bar", "./Bar.cdc", nil, flow.HexToAddress("0x2"), "", nil),
			NewContract("Foo", "./Foo.cdc", nil, flow.HexToAddress("0x1"), "", nil),
			NewContract("Zoo", "./Zoo.cdc", nil, flow.HexToAddress("0x2"), "", nil),
		}

		replacer := NewImportReplacer(contracts, nil)

		code := []byte(`
			import Foo from "./Foo.cdc"
			import "Bar"

			access(all) contract Zoo {}
		`)
		program, err := NewProgram(code, nil, "./Zoo.cdc")
		require.NoError(t, err)

		replaced, err := replacer.Replace(program)
		require.NoError(t, err)

		expected := []byte(`
			import Foo from 0x0000000000000001
			import Bar from 0x0000000000000002

			access(all) contract Zoo {}
		`)

		assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
	})

	t.Run("Resolve import aliases", func(t *testing.T) {
		fusdCode := []byte(`access(all) contract FUSD {}`)

		t.Run("Basic import alias - same source, different keys", func(t *testing.T) {
			contracts := []*Contract{
				NewContract("FUSD1", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x1"), "", nil),
				NewContract("FUSD2", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x2"), "", nil),
			}

			replacer := NewImportReplacer(contracts, nil)

			code := []byte(`
				import "FUSD1"
				import "FUSD2"

				access(all) fun main(): UFix64 {
					return FUSD1.totalSupply + FUSD2.totalSupply
				}
			`)

			program, err := NewProgram(code, nil, "./scripts/test.cdc")
			require.NoError(t, err)

			replaced, err := replacer.Replace(program)
			require.NoError(t, err)

			expected := []byte(`
				import FUSD as FUSD1 from 0x0000000000000001
				import FUSD as FUSD2 from 0x0000000000000002

				access(all) fun main(): UFix64 {
					return FUSD1.totalSupply + FUSD2.totalSupply
				}
			`)

			assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
		})

		t.Run("Smart name matching - key matches source name", func(t *testing.T) {
			contracts := []*Contract{
				NewContract("FUSD", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x1"), "", nil),
				NewContract("FUSD2", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x2"), "", nil),
			}

			replacer := NewImportReplacer(contracts, nil)

			code := []byte(`
				import "FUSD"
				import "FUSD2"

				access(all) fun main() {}
			`)

			program, err := NewProgram(code, nil, "./scripts/test.cdc")
			require.NoError(t, err)

			replaced, err := replacer.Replace(program)
			require.NoError(t, err)

			expected := []byte(`
				import FUSD from 0x0000000000000001
				import FUSD as FUSD2 from 0x0000000000000002

				access(all) fun main() {}
			`)

			assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
		})

		t.Run("Path normalization - different path formats", func(t *testing.T) {
			contracts := []*Contract{
				NewContract("FUSD1", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x1"), "", nil),
				NewContract("FUSD2", "contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x2"), "", nil),
			}

			replacer := NewImportReplacer(contracts, nil)

			code := []byte(`
				import "FUSD1"
				import "FUSD2"

				access(all) fun main() {}
			`)

			program, err := NewProgram(code, nil, "./scripts/test.cdc")
			require.NoError(t, err)

			replaced, err := replacer.Replace(program)
			require.NoError(t, err)

			// Both should be recognized as sharing the same source file
			expected := []byte(`
				import FUSD as FUSD1 from 0x0000000000000001
				import FUSD as FUSD2 from 0x0000000000000002

				access(all) fun main() {}
			`)

			assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
		})

		t.Run("Mixed scenario - aliases and regular imports", func(t *testing.T) {
			kibbleCode := []byte(`access(all) contract Kibble {}`)

			contracts := []*Contract{
				NewContract("FUSD1", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x1"), "", nil),
				NewContract("FUSD2", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x2"), "", nil),
				NewContract("Kibble", "./contracts/Kibble.cdc", kibbleCode, flow.HexToAddress("0x3"), "", nil),
			}

			replacer := NewImportReplacer(contracts, nil)

			code := []byte(`
				import "FUSD1"
				import "FUSD2"
				import "Kibble"

				access(all) fun main() {}
			`)

			program, err := NewProgram(code, nil, "./scripts/test.cdc")
			require.NoError(t, err)

			replaced, err := replacer.Replace(program)
			require.NoError(t, err)

			expected := []byte(`
				import FUSD as FUSD1 from 0x0000000000000001
				import FUSD as FUSD2 from 0x0000000000000002
				import Kibble from 0x0000000000000003

				access(all) fun main() {}
			`)

			assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
		})

		t.Run("No alias needed - single contract from file", func(t *testing.T) {
			contracts := []*Contract{
				NewContract("FUSD", "./contracts/FUSD.cdc", fusdCode, flow.HexToAddress("0x1"), "", nil),
			}

			replacer := NewImportReplacer(contracts, nil)

			code := []byte(`
				import "FUSD"

				access(all) fun main() {}
			`)

			program, err := NewProgram(code, nil, "./scripts/test.cdc")
			require.NoError(t, err)

			replaced, err := replacer.Replace(program)
			require.NoError(t, err)

			// Should use regular import syntax
			expected := []byte(`
				import FUSD from 0x0000000000000001

				access(all) fun main() {}
			`)

			assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
		})
	})

}
