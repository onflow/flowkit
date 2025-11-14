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

	t.Run("Resolve imports with canonical aliases", func(t *testing.T) {
		// Create contracts - FUSD1 and FUSD2 are alias deployments of FUSD
		// In practice, only the canonical contract (FUSD) would be deployed, and
		// FUSD1/FUSD2 would be aliases pointing to different addresses where FUSD is deployed
		contracts := []*Contract{
			NewContract("FUSD", "./contracts/FUSD.cdc", nil, flow.HexToAddress("0x1"), "", nil),
			NewContract("FUSD1", "", nil, flow.HexToAddress("0x2"), "", nil),  // Alias deployment
			NewContract("FUSD2", "", nil, flow.HexToAddress("0x3"), "", nil),  // Another alias deployment
			NewContract("FT", "./contracts/FT.cdc", nil, flow.HexToAddress("0x4"), "", nil),       // Regular contract
		}
		
		// Canonical mapping to simulate FUSD1 and FUSD2 having FUSD as canonical
		canonicalMapping := map[string]string{
			"FUSD1": "FUSD",
			"FUSD2": "FUSD",
		}
		
		replacer := &ImportReplacer{
			contracts:        contracts,
			aliases:          nil,
			canonicalMapping: canonicalMapping,
		}
		
		t.Run("basic alias replacement", func(t *testing.T) {
			code := []byte(`
				import "FUSD"
				import "FUSD1"
				import "FUSD2"
				import "FT"
				
				access(all) contract Test {}
			`)
			
			program, err := NewProgram(code, nil, "./Test.cdc")
			require.NoError(t, err)
			
			replaced, err := replacer.Replace(program)
			require.NoError(t, err)
			
			expected := []byte(`
				import FUSD from 0x0000000000000001
				import FUSD as FUSD1 from 0x0000000000000002
				import FUSD as FUSD2 from 0x0000000000000003
				import FT from 0x0000000000000004
				
				access(all) contract Test {}
			`)
			
			assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
		})
	})

	t.Run("ConvertAddressImports with aliases", func(t *testing.T) {
		code := []byte(`
			import FUSD from 0x0000000000000001
			import FUSD as FUSD1 from 0x0000000000000002
			import FUSD as FUSD2 from 0x0000000000000003
			import FT from 0x0000000000000004
			
			access(all) contract Test {}
		`)
		
		program, err := NewProgram(code, nil, "./Test.cdc")
		require.NoError(t, err)
		
		// ConvertAddressImports should convert both regular and alias imports back to identifier imports
		// For alias imports (import X as Y from 0x...), it should use the alias name (Y)
		expected := []byte(`
			import "FUSD"
			import "FUSD1"
			import "FUSD2"
			import "FT"
			
			access(all) contract Test {}
		`)
		
		assert.Equal(t, cleanCode(expected), cleanCode(program.CodeWithUnprocessedImports()))
	})

	t.Run("Import replacer with no canonical mapping", func(t *testing.T) {
		// Test that contracts work normally without canonical mapping
		contracts := []*Contract{
			NewContract("Token", "./contracts/Token.cdc", nil, flow.HexToAddress("0x1"), "", nil),
		}
		
		replacer := NewImportReplacer(contracts, nil)
		
		code := []byte(`
			import "Token"
			
			access(all) contract Test {}
		`)
		
		program, err := NewProgram(code, nil, "./Test.cdc")
		require.NoError(t, err)
		
		replaced, err := replacer.Replace(program)
		require.NoError(t, err)
		
		expected := []byte(`
			import Token from 0x0000000000000001
			
			access(all) contract Test {}
		`)
		
		assert.Equal(t, cleanCode(expected), cleanCode(replaced.Code()))
	})

}
