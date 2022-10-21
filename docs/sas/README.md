# Specifying SAS Initiators

## SAS Initiator Discovery

The node driver will attempt to discover the address of any available SAS initiators. This may not work for all brands/models of SAS HBA, so if you need or prefer to specify these values manually, you can do so in the file `/etc/kubernetes/sas-addresses`. You must create or update this file on each Kubernetes node where you have SAS connections to a storage array.

Example of finding the SAS host address on your node:
```
# lsscsi -t -H
[0]    ata_piix      ata:
[1]    ata_piix      ata:
[2]    mpt3sas       sas:0x500605b00b4ec831
```
Note: The `lsscsi` command may only find the 1st SAS host address. In this case we needed to add one to that value to get the 2nd SAS host address.

Example contents of the `sas-addresses` file, **note that the '0x' prefix is omitted**:
```
500605b00b4ec831
500605b00b4ec832
```

## Specify your initiator as a topology value in your storage class:
Use the CSI Topology feature to ensure your PVCs are only scheduled on Nodes that are connected to your SAS array. If this field is not present in your storage class, PVCs for this storage class
may be scheduled on Nodes that are not connected to the array and they will fail.

Also see the storage class example file in this directory
```
allowedTopologies:
  - matchLabelExpressions:
    - key: com.seagate-exos-x-csi/sas-address-0
      values:
      - 500605b00b4ec831 # The first sas initiator on your node
```