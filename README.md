# Flowkit

Flowkit is a Go library for building applications that interact with the Flow blockchain. It provides high-level APIs for managing Flow projects, including working with `flow.json` configurations, deploying contracts, executing scripts, and building transactions.

## Features

- **Project Management** - Load and manage Flow project configurations (`flow.json`)
- **Contract Deployment** - Deploy Cadence contracts to different networks
- **Import Resolution** - Automatically resolve contract imports to blockchain addresses
- **Multi-Network Support** - Work seamlessly with emulator, testnet, and mainnet
- **Account Management** - Manage Flow accounts and signing keys
- **Script Execution** - Execute Cadence scripts and build transactions

## Installation

```bash
go get github.com/onflow/flowkit/v2
```

## Quick Example

```go
import "github.com/onflow/flowkit/v2"

// Load your Flow project
state, err := flowkit.Load([]string{"flow.json"}, afero.Afero{Fs: afero.NewOsFs()})

// Get contracts for a network
contracts, err := state.DeploymentContractsByNetwork(config.TestnetNetwork)

// Resolve imports in your Cadence code
importReplacer := project.NewImportReplacer(contracts, state.AliasesForNetwork(network))
resolvedProgram, err := importReplacer.Replace(program)
```

## Documentation

For comprehensive guides and examples, visit the [Flowkit documentation](https://developers.flow.com/build/tools/clients/flow-go-sdk/flowkit).

## Package Structure

Flowkit contains multiple subpackages:

- **config** - Parsing and storing of `flow.json` values, as well as validation
- **gateway** - Implementation of Flow Access Node methods, uses emulator and Go SDK to communicate with ANs
- **project** - Stateful operations on top of `flow.json`, including import resolution for contract deployments
- **accounts** - Account and key management
- **transactions** - Transaction building and signing

The main Flowkit interface is defined in [services.go](services.go).
