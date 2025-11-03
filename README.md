## Flowkit Package Design

Flowkit is a core package used by the CLI commands. It features APIs for interacting with the Flow network
in the context of flow.json configuration values. Flowkit is defined by the [interface here](services.go).

Flowkit contains multiple subpackages, the most important ones are:
- **config**: parsing and storing of flow.json values, as well as validation,
- **gateway**: implementation of Flow AN methods, uses emulator as well as Go SDK to communicate with ANs,
- **project**: stateful operations on top of flow.json, which allows resolving imports in contracts used in deployments

## Documentation

For detailed usage examples and tutorials on using Flowkit in your Go applications, see the [Flowkit documentation](https://developers.flow.com/build/tools/clients/flow-go-sdk/flowkit).
