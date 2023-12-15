package dependencymanager

import (
	"fmt"
	"sync"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	"github.com/onflow/flow-go/model/flow"

	"github.com/onflow/flow-cli/flowkit/gateway"

	"github.com/onflow/flow-cli/flowkit/project"

	"github.com/onflow/flow-cli/flowkit/config"
	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
)

type DependencyInstaller struct {
	Gateways map[string]gateway.Gateway
	Logger   output.Logger
	State    *flowkit.State
	Mutex    sync.Mutex
}

func NewDepdencyInstaller(logger output.Logger, state *flowkit.State) *DependencyInstaller {
	emulatorGateway, err := gateway.NewGrpcGateway(config.EmulatorNetwork)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating emulator gateway: %v", err))
	}

	testnetGateway, err := gateway.NewGrpcGateway(config.TestnetNetwork)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating testnet gateway: %v", err))
	}

	mainnetGateway, err := gateway.NewGrpcGateway(config.MainnetNetwork)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating mainnet gateway: %v", err))
	}

	gateways := map[string]gateway.Gateway{
		config.EmulatorNetwork.Name: emulatorGateway,
		config.TestnetNetwork.Name:  testnetGateway,
		config.MainnetNetwork.Name:  mainnetGateway,
	}

	return &DependencyInstaller{
		Gateways: gateways,
		Logger:   logger,
		State:    state,
	}
}

func (ci *DependencyInstaller) install() error {
	for _, dependency := range *ci.State.Dependencies() {
		if err := ci.processDependency(dependency); err != nil {
			ci.Logger.Error(fmt.Sprintf("Error processing dependency: %v", err))
			return err
		}
	}
	return nil
}

func (ci *DependencyInstaller) add(depRemoteSource, customName string) error {
	depNetwork, depAddress, depContractName, err := config.ParseRemoteSourceString(depRemoteSource)
	if err != nil {
		return fmt.Errorf("error parsing remote source: %w", err)
	}

	var name string

	if customName != "" {
		name = customName
	} else {
		name = depContractName
	}

	dep := config.Dependency{
		Name: name,
		RemoteSource: config.RemoteSource{
			NetworkName:  depNetwork,
			Address:      flowsdk.HexToAddress(depAddress),
			ContractName: depContractName,
		},
	}

	if err := ci.processDependency(dep); err != nil {
		return fmt.Errorf("error processing dependency: %w", err)
	}

	return nil
}

func (ci *DependencyInstaller) processDependency(dependency config.Dependency) error {
	depAddress := flowsdk.HexToAddress(dependency.RemoteSource.Address.String())
	return ci.fetchDependencies(dependency.RemoteSource.NetworkName, depAddress, dependency.Name, dependency.RemoteSource.ContractName)
}

func (ci *DependencyInstaller) fetchDependencies(networkName string, address flowsdk.Address, assignedName, contractName string) error {
	account, err := ci.Gateways[networkName].GetAccount(address)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}
	if account == nil {
		return fmt.Errorf("account is nil for address: %s", address)
	}

	if account.Contracts == nil {
		return fmt.Errorf("contracts are nil for account: %s", address)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(account.Contracts))

	for _, contract := range account.Contracts {

		program, err := project.NewProgram(contract, nil, "")
		if err != nil {
			return fmt.Errorf("failed to parse program: %v", err)
		}

		parsedContractName, err := program.Name()
		if err != nil {
			return fmt.Errorf("failed to parse contract name: %v", err)
		}

		if parsedContractName == contractName {
			program.ConvertImports()

			if err := ci.handleFoundContract(networkName, address.String(), assignedName, parsedContractName, string(program.DevelopmentCode())); err != nil {
				return fmt.Errorf("failed to handle found contract: %v", err)
			}

			if program.HasAddressImports() {
				imports := program.AddressImportDeclarations()
				for _, imp := range imports {
					wg.Add(1)
					go func(importAddress flowsdk.Address, contractName string) {
						defer wg.Done()
						err := ci.fetchDependencies("testnet", importAddress, contractName, contractName)
						if err != nil {
							errCh <- err
						}
					}(flowsdk.HexToAddress(imp.Location.String()), imp.Identifiers[0].String())
				}
			}
		}
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (ci *DependencyInstaller) handleFoundContract(networkName, contractAddr, assignedName, contractName, contractData string) error {
	ci.Mutex.Lock()
	defer ci.Mutex.Unlock()

	if !contractFileExists(contractAddr, contractName) {
		if err := createContractFile(contractAddr, contractName, contractData); err != nil {
			return fmt.Errorf("failed to create contract file: %v", err)
		}
	}

	err := ci.updateState(networkName, contractAddr, assignedName, contractName)
	if err != nil {
		ci.Logger.Error(fmt.Sprintf("Error updating state: %v", err))
		return err
	}

	return nil
}

const (
	NetworkEmulator = "emulator"
	NetworkTestnet  = "testnet"
	NetworkMainnet  = "mainnet"
)

var networkToChainID = map[string]flow.ChainID{
	NetworkEmulator: flow.Emulator,
	NetworkTestnet:  flow.Testnet,
	NetworkMainnet:  flow.Mainnet,
}

func isCoreContract(networkName, contractName, contractAddress string) bool {
	sc := systemcontracts.SystemContractsForChain(networkToChainID[networkName])
	coreContracts := sc.All()

	for _, coreContract := range coreContracts {
		if coreContract.Name == contractName && coreContract.Address.String() == contractAddress {
			return true
		}
	}

	return false
}

func getCoreContractByName(networkName, contractName string) *systemcontracts.SystemContract {
	sc := systemcontracts.SystemContractsForChain(networkToChainID[networkName])

	for i, coreContract := range sc.All() {
		if coreContract.Name == contractName {
			return &sc.All()[i]
		}
	}

	return nil
}

func (ci *DependencyInstaller) updateState(networkName, contractAddress, assignedName, contractName string) error {
	dep := config.Dependency{
		Name: assignedName,
		RemoteSource: config.RemoteSource{
			NetworkName:  networkName,
			Address:      flowsdk.HexToAddress(contractAddress),
			ContractName: contractName,
		},
	}

	var aliases []config.Alias

	// If core contract found by name and address matches, then use all core contract aliases across networks
	if isCoreContract(networkName, contractName, contractAddress) {
		for _, networkStr := range []string{NetworkEmulator, NetworkTestnet, NetworkMainnet} {
			coreContract := getCoreContractByName(networkStr, contractName)
			if coreContract != nil {
				aliases = append(aliases, config.Alias{
					Network: networkStr,
					Address: flowsdk.HexToAddress(coreContract.Address.String()),
				})
			}
		}
	}

	// If no core contract match, then use the address in remoteSource as alias
	if len(aliases) == 0 {
		aliases = append(aliases, config.Alias{
			Network: dep.RemoteSource.NetworkName,
			Address: dep.RemoteSource.Address,
		})
	}

	ci.State.Dependencies().AddOrUpdate(dep)
	ci.State.Contracts().AddDependencyAsContract(dep, aliases)
	err := ci.State.SaveDefault()
	if err != nil {
		return err
	}

	return nil
}
