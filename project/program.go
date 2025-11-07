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
	"fmt"
	"regexp"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/parser"
)

type Program struct {
	code                       []byte
	args                       []cadence.Value
	location                   string
	astProgram                 *ast.Program
	codeWithUnprocessedImports []byte
}

func NewProgram(code []byte, args []cadence.Value, location string) (*Program, error) {
	astProgram, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return nil, err
	}

	p := &Program{
		code:                       code,
		args:                       args,
		location:                   location,
		astProgram:                 astProgram,
		codeWithUnprocessedImports: code, // has converted import syntax e.g. 'import "Foo"'
	}
	
	// Convert address imports to identifier imports for codeWithUnprocessedImports
	p.ConvertAddressImports()
	
	return p, nil
}

func (p *Program) AddressImportDeclarations() []*ast.ImportDeclaration {
	addressImports := make([]*ast.ImportDeclaration, 0)

	for _, importDeclaration := range p.astProgram.ImportDeclarations() {
		if len(importDeclaration.Imports) > 0 && len(importDeclaration.Location.String()) > 0 {
			addressImports = append(addressImports, importDeclaration)
		}
	}

	return addressImports
}

func (p *Program) HasAddressImports() bool {
	return len(p.AddressImportDeclarations()) > 0
}

// Imports builds an array of all the import locations
// It currently supports getting import locations as identifiers or as strings. Strings locations
// can represent a file or an account name, whereas identifiers represent contract names.
func (p *Program) imports() []string {
	imports := make([]string, 0)

	for _, importDeclaration := range p.astProgram.ImportDeclarations() {
		// we parse all string locations, that are all imports that look like "import X from "Y"" or "import "X""
		_, isStringImport := importDeclaration.Location.(common.StringLocation)
		if isStringImport {
			imports = append(imports, importDeclaration.Location.String())
		}
	}

	return imports
}

func (p *Program) HasImports() bool {
	return len(p.imports()) > 0
}

func (p *Program) replaceImport(from string, to string, canonicalName ...string) *Program {
	code := string(p.Code())

	// Extract the import name from the 'from' parameter
	importName := from
	if regexp.MustCompile(`\.cdc$`).MatchString(from) {
		// If it's a path, extract the contract name
		matches := regexp.MustCompile(`([^/]+)\.cdc$`).FindStringSubmatch(from)
		if len(matches) > 1 {
			importName = matches[1]
		}
	}

	// Handle path imports (e.g., import X from "./X.cdc")
	pathRegex := regexp.MustCompile(fmt.Sprintf(`import\s+(\w+)\s+from\s+"%s"`, regexp.QuoteMeta(from)))
	// Handle identifier imports (e.g., import "X")
	identifierRegex := regexp.MustCompile(fmt.Sprintf(`import\s+"%s"`, regexp.QuoteMeta(from)))

	// Determine if we need alias syntax
	canonical := ""
	if len(canonicalName) > 0 {
		canonical = canonicalName[0]
	}
	
	if canonical != "" && canonical != importName {
		// Use alias syntax: import CanonicalName as AliasName from 0xAddress
		pathReplacement := fmt.Sprintf(`import %s as $1 from 0x%s`, canonical, to)
		identifierReplacement := fmt.Sprintf(`import %s as %s from 0x%s`, canonical, importName, to)
		
		code = pathRegex.ReplaceAllString(code, pathReplacement)
		code = identifierRegex.ReplaceAllString(code, identifierReplacement)
	} else {
		// Use regular syntax: import Name from 0xAddress
		replacement := fmt.Sprintf(`import $1 from 0x%s`, to)
		code = pathRegex.ReplaceAllString(code, replacement)
		
		identifierReplacement := fmt.Sprintf(`import %s from 0x%s`, importName, to)
		code = identifierRegex.ReplaceAllString(code, identifierReplacement)
	}

	p.code = []byte(code)
	p.reload()
	return p
}

func (p *Program) Location() string {
	return p.location
}

func (p *Program) Code() []byte {
	return p.code
}

func (p *Program) CodeWithUnprocessedImports() []byte {
	return p.codeWithUnprocessedImports
}

func (p *Program) Name() (string, error) {
	if len(p.astProgram.CompositeDeclarations()) > 1 || len(p.astProgram.InterfaceDeclarations()) > 1 ||
		len(p.astProgram.CompositeDeclarations())+len(p.astProgram.InterfaceDeclarations()) > 1 {
		return "", fmt.Errorf("the code must declare exactly one contract or contract interface")
	}

	for _, compositeDeclaration := range p.astProgram.CompositeDeclarations() {
		if compositeDeclaration.CompositeKind == common.CompositeKindContract {
			return compositeDeclaration.Identifier.Identifier, nil
		}
	}

	for _, interfaceDeclaration := range p.astProgram.InterfaceDeclarations() {
		if interfaceDeclaration.CompositeKind == common.CompositeKindContract {
			return interfaceDeclaration.Identifier.Identifier, nil
		}
	}

	return "", fmt.Errorf("unable to determine contract name")
}

func (p *Program) ConvertAddressImports() {
	code := string(p.code)
	// Handle regular imports: import X from 0xAddress
	addressImportRegex := regexp.MustCompile(`import\s+(\w+)\s+from\s+0x[0-9a-fA-F]+`)
	modifiedCode := addressImportRegex.ReplaceAllString(code, `import "$1"`)
	
	// Handle alias imports: import X as Y from 0xAddress -> import "Y"
	aliasImportRegex := regexp.MustCompile(`import\s+\w+\s+as\s+(\w+)\s+from\s+0x[0-9a-fA-F]+`)
	modifiedCode = aliasImportRegex.ReplaceAllString(modifiedCode, `import "$1"`)

	p.codeWithUnprocessedImports = []byte(modifiedCode)
}

func (p *Program) reload() {
	astProgram, err := parser.ParseProgram(nil, p.code, parser.Config{})
	if err != nil {
		return
	}

	p.astProgram = astProgram
	p.ConvertAddressImports()
}
