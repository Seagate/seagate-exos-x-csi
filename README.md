# Seagate CSI dynamic provisioner for Kubernetes

Seagate Exos X CSI driver supports Seagate storage systems with 4xx5/5xx5 controllers (including OEM versions)

[![Go Report Card](https://goreportcard.com/badge/github.com/Seagate/seagate-exos-x-csi)](https://goreportcard.com/report/github.com/Seagate/seagate-exos-x-csi)

## Introduction

Seagate Exos X CSI Driver helps users of storage systems with 4xx5/5xx5 controllers from Seagate and OEM vendors efficiently manage their storage within container platforms that support the CSI standard.
Dealing with persistent storage on Kubernetes can be particularly cumbersome, especially when dealing with on-premises installations, or when the cloud-provider persistent storage solutions are not applicable.
The Seagate CSI Driver is a direct result of customer demand to bring the ease of use of Seagate Exos X to DevOps practices, and demonstrates Seagate’s continued commitment to the Kubernetes ecosystem

More information about Seagate Data Storage Systems can be found [online](https://www.seagate.com/products/storage/data-storage-systems/)

## This project

This project implements the **Container Storage Interface** in order to facilitate dynamic provisioning of persistent volumes on a Kubernetes cluster.

This CSI driver is an open-source project under the Apache 2.0 [license](./LICENSE).

## Key Features
- Manage persistent volumes backed by iSCSI protocols on Exos X enclosures
- Control multiple Exos X systems within a single Kubernetes cluster
- Manage Exos X snapshots and clones, including restoring from snapshots
- Clone, extend and manage persistent volumes created outside of the Exos CSI Driver
- Collect usage and performance metrics for CSI driver usage and expose them via an open-source systems monitoring and alerting toolkit, such as Prometheus

## Installation

### Install ISCSI tools and Multipath driver on your node(s)

`iscsid` and `multipathd` must be installed on every node. Check the installation method appropriate for your Linux distribution.
#### Ubuntu installation procedure
- Remove any containers that were running a prior CSI Driver version.
- Install required packages:
```
    sudo apt update && sudo apt install open-iscsi scsitools multipath-tools -y
```
- Determine if any packages are required for your filesystem (ext3/ext4/xfs) choice and view current support:
```
cat /proc/filesystems
```
- Update /etc/multipath.conf with the following lines:
```
    defaults {
      polling_interval 2
      find_multipaths "yes"
      retain_attached_hw_handler "no"
      disable_changed_wwids "yes"
    }
    devices {
            device {
            vendor "HP"
            product "MSA 2040 SAN"
            path_grouping_policy "group_by_prio"
            getuid_callout "/lib/udev/scsi_id --whitelisted --device=/dev/%n"
            prio "alua"
            path_selector "round-robin 0"
            path_checker "tur"
            hardware_handler "0"
            failback "immediate"
            rr_weight "uniform"
            rr_min_io_rq 1
            no_path_retry 18
            }
    }
```
- Restart MultipathD
```
    service multipath-tools restart
```

### Deploy the provisioner to your kubernetes cluster

The preferred installation approach is to use the provided `Helm Charts` under the helm folder.
```
  helm install seagate-csi -n seagate --create-namespace \
    helm/csi-charts -f helm/csi-charts/values.yaml
```

#### To deploy the provisioner to OpenShift cluster, run the following commands prior to using Helm:
```
    oc create -f scc/exos-x-csi-access-scc.yaml --as system:admin
    oc adm policy add-scc-to-user exos-x-csi-access -z default -n NAMESPACE
```

#### Configure your release

- Update `helm/csi-charts/values.yaml` to match your configuration settings.
- Update `example/secret-example1.yaml` with your storage controller credentials.
- Update `example/storageclass-example1.yaml` with your storage controller values.
- Update `example/testpod-example1.yaml` with any of you new values.


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
