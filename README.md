# Seagate CSI dynamic provisioner for Kubernetes

The Seagate Exos X CSI Driver supports the following storage arrays

- Seagate Exos X and AssuredSAN (4006/5005/4005/3005)
- Dell PowerVault ME4 and ME5 Series

iSCSI, SAS, and FC host interfaces are supported for both block and filesystem mount types

[![Go Report Card](https://goreportcard.com/badge/github.com/Seagate/seagate-exos-x-csi)](https://goreportcard.com/report/github.com/Seagate/seagate-exos-x-csi)

## Introduction

The Seagate Exos X CSI Driver helps storage admins efficiently manage
their storage within container platforms that support the CSI
standard.  Dealing with persistent storage on Kubernetes can be
particularly cumbersome, especially when dealing with on-premises
installations, or when the cloud-provider persistent storage solutions
are not applicable.  The Seagate CSI Driver is a direct result of
customer demand to bring the ease of use of Seagate Exos X to DevOps
practices, and demonstrates Seagate’s continued commitment to the
Kubernetes ecosystem

More information about Seagate Data Storage Systems can be found
[online](https://www.seagate.com/products/storage/data-storage-systems/)

## This project

This project implements the **Container Storage Interface** in order to facilitate dynamic provisioning of persistent volumes on a Kubernetes cluster.

This CSI driver is an open-source project under the Apache 2.0 [license](./LICENSE).

## Key Features
- Manage persistent volumes on Exos X enclosures
- Control multiple Exos X systems within a single Kubernetes cluster
- Manage Exos X snapshots and clones, including restoring from snapshots
- Clone, extend and manage persistent volumes created outside of the Exos CSI Driver
- Collect usage and performance metrics for CSI driver usage and expose them via an open-source systems monitoring and alerting toolkit, such as Prometheus

## Installation

### Install iSCSI tools and multipath driver on your nodes

`iscsid` and `multipathd` must be installed on every node. Check the
installation method appropriate for your Linux distribution.  The
example below shows steps for Ubuntu Server but the process will be
very similar for other GNU/Linux distributions.

#### Ubuntu Installation procedure
- Remove any containers that were running an earlier version of the Seagate Exos X CSI Driver.
- Install required packages:

    ```
    sudo apt update && sudo apt install open-iscsi scsitools multipath-tools -y
    ```
- Determine if any packages are required for your filesystem (ext3/ext4/xfs) choice and view current support:

    ```
    cat /proc/filesystems
    ```
- Update /etc/multipath.conf. Check docs/iscsi/multipath.conf as a reference. In particular ensure your configuration includes these settings:
    ```
    find_multipaths "greedy"
    user_friendly_names		"no"
    ```

- Restart `multipathd`:

    ```
    service multipath-tools restart
    ```

### Deploy the provisioner to your kubernetes cluster

These examples assume you have already installed the [helm]() command.

The easiest method for installing the driver is to use Helm to install
the helm package from
[Github](https://github.com/seagate/seagate-exos-x-csi/releases).  On
the Releases page, right-click on the Helm Package and select "Copy
Link Address".  Choose a namespace in which to run
the driver (in this example, _seagate_), and a name for the
application (_exos-x-csi_) and then paste the link the onto the end of
this command.  For example: 
```
helm install --create-namespace -n seagate exos-x-csi <url-of-helm-package>
```

Alternately, you can download and unpack the [helm
package](https://github.com/Seagate/seagate-exos-x-csi/releases/download/v1.6.3/seagate-exos-x-csi-1.6.3.tgz)
and extract it:
```
wget https://github.com/Seagate/seagate-exos-x-csi/releases/download/v1.6.3/seagate-exos-x-csi-1.6.3.tgz
tar xpzf seagate-exos-x-csi-1.6.3.tgz
helm install --create-namespace -n seagate exos-x-csi seagate-exos-x-csi
```
or clone the Github repository and install from the helm/csi-charts folder:

```
git clone https://github.com/Seagate/seagate-exos-x-csi
cd seagate-exos-x-csi
helm install exos-x-csi -n seagate --create-namespace \
  helm/csi-charts -f helm/csi-charts/values.yaml
```

#### To deploy the provisioner to OpenShift cluster, run the following commands prior to using Helm:
```
oc create -f scc/exos-x-csi-access-scc.yaml --as system:admin
oc adm policy add-scc-to-user exos-x-csi-access -z default -n NAMESPACE
oc adm policy add-scc-to-user exos-x-csi-access -z csi-provisioner -n NAMESPACE
```

#### Configure your release

- Update `helm/csi-charts/values.yaml` to match your configuration settings.
- Update `example/secret-example1.yaml` with your storage controller credentials. Use `example/secret-example2-CHAP.yaml` if you wish to specify CHAP credentials as well. 
- Update `example/storageclass-example1.yaml` with your storage controller values. Use `example/storageclass-example2-CHAP.yaml` if you are using CHAP authentication
- Update `example/testpod-example1.yaml` with any of you new values.

## Documentation

You can find more documentation in the [docs](./docs) directory.
Check docs/Seagate_Exos_X_CSI_driver_functionality.ipynb for usage examples and configuration files.

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
