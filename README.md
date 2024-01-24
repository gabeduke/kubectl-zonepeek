Sure, here's a basic README for your `main.go` script:

# kubectl-zonepeek

`kubectl-zonepeek` is a kubectl plugin written in Go that gathers zone information from nodes and associated Persistent Volumes in a Kubernetes cluster.

## Features

- Retrieves the zone information of nodes and associated Persistent Volumes.
- Accepts a label selector to identify the pods.

## Prerequisites

- Go 1.16 or higher
- Access to a Kubernetes cluster and the `kubectl` command-line tool

## Installation

To install `kubectl-zonepeek`, clone the repository and build the binary:

```bash
go install github.com/gabeduke/kubectl-zonepeek@latest
```

## Usage

To use `kubectl-zonepeek`, you need to pass a label selector to identify the pods:

```bash
kubectl zonepeek --label <label-selector>
```

You can specify the output format (table, json, text) with the `--output` flag:

```bash
kubectl zonepeek --label <label-selector> --output json
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.