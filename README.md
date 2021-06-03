# Seagate EXOS X CSI dynamic provisioner for Kubernetes

A dynamic persistent volume (PV) provisioner for Seagate EXOS X based storage systems.

[![Go Report Card](https://goreportcard.com/badge/github.com/Seagate/seagate-exos-x-csi)](https://goreportcard.com/report/github.com/Seagate/seagate-exos-x-csi)

## Introduction

Dealing with persistent storage on Kubernetes can be particularly cumbersome, especially when dealing with on-premises installations, or when the cloud-provider persistent storage solutions are not applicable.

Entry-level SAN appliances usually propose a low-cost, still powerful, solution to enable redundant persistent storage, with the flexibility of attaching it to any host on your network.

Seagate continues to maintain the line-up with subsequent series :
- [Seagate AssuredSAN](https://www.seagate.com/fr/fr/support/dothill-san/assuredsan-pro-5000-series/) 3000/4000/5000/6000 series

## This project

`exosx-csi` implements the **Container Storage Interface** in order to facilitate dynamic provisioning of persistent volumes on a Kubernetes cluster.

All Exos X based equipements share a common API.

This CSI driver is an open-source project under the Apache 2.0 [license](./LICENSE).

## Features
- dynamic provisioning
- resize
- snapshot
- prometheus metrics

## Installation

### Uninstall ISCSI tools on your node(s)

`iscsid` and `multipathd` are now shipped as sidecars on each nodes, it is therefore strongly suggested to uninstall any `open-iscsi` and `multipath-tools` package.

The decision of shipping `iscsid` and `multipathd` as sidecars comes from the desire to simplify the developpement process, as well as improving monitoring. It's essential that versions of those softwares match the candidates versions on your hosts, more about this in the [FAQ](./docs/troubleshooting.md#multipathd-segfault-or-a-volume-got-corrupted). This setup is currently being challenged.

### Deploy the provisioner to your kubernetes cluster

The preferred installation approach is to use the provided `Helm Charts` under the helm folder.

#### Configure your release

- Update `helm/exosx-csi/values.yaml` to match your configuration settings.
- Update `example/secret-example1.yaml` with your storage controller credentials.
- Update `example/storageclass-example1.yaml` with your storage controller values.
- Update `example/testpod-example1.yaml` with any of you new values.

#### Run the Installation Script

```sh
cd example
./testpod-start.sh example1
```

This script will install the local helm charts, create secrets, create the storage class, and then create a test pod. To clean up after a run.

```sh
cd example
./testpod-stop.sh
```

## Documentation

You can find more documentation in the [docs](./docs) directory.

## Command-line arguments

You can have a list of all available command line flags using the `-help` switch.

### Logging

Logging can be modified using the `-v` flag :

- `-v 0` : Standard logs to follow what's going on (default if not specified)
- `-v 9` : Debug logs (quite awful to see)

For advanced logging configuration, see [klog](https://github.com/kubernetes/klog).

### Development

You can start the drivers over TCP so your remote dev cluster can connect to them.

```
go run ./cmd/<driver> -bind=tcp://0.0.0.0:10000
```

## Testing

You can run sanity checks by using the `sanity` helper script in the `test/` directory:

```
./test/sanity
```
