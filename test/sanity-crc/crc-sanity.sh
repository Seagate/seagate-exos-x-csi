#!/bin/bash

# Launch cli sanity in a crc pod. Must be logged in with oc before running.

secretsTemplate="../secrets.template.yml"
secrets="secrets.yml"
volumeTemplate="../volume.template.yml"
volume="volume.yml"

sanity=/home/dwhite/github.com/csi-test/cmd/csi-sanity/csi-sanity

set -e

#make sure oc command is setup
oc > /dev/null 

function setup {
    cd $(dirname $0)
    set -a; . ../.env; set +a

    echo ""

    envsubst < ${secretsTemplate} > ${secrets}
    echo "===== ${secrets} ====="
    cat ${secrets}
    echo "===== END ====="

    echo ""

    envsubst < ${volumeTemplate} > ${volume}
    echo "===== ${volume} ====="
    cat ${volume}
    echo "===== END ====="

    cp $sanity .
}

setup

#Build and push the sanity container
echo "===== Building Container ====="
buildah bud -t localhost/seagate-exos-x-csi/csi-sanity .
podman login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing --tls-verify=false
podman tag localhost/seagate-exos-x-csi/csi-sanity default-route-openshift-image-registry.apps-crc.testing/seagate/csi-sanity
podman push default-route-openshift-image-registry.apps-crc.testing/seagate/csi-sanity --tls-verify=false

#Retrieve the controller UID, needed to mount the CSI socket from the CRC VM into the container
controller_pod_name=$(oc get pods -n seagate -o name | grep seagate-exos-x-csi-controller-server)
controller_pod_uid=$(oc get $controller_pod_name -o=jsonpath='{.metadata.uid}' -n seagate)
sed "s/{{CONTROLLER_POD_UID}}/$controller_pod_uid/g" sanity-crc-template.yaml > sanity-crc.yaml

set +e

# #Run sanity
echo "===== Deleting Old Sanity Pod ====="
oc delete pod csi-sanity-crc

echo "===== Creating Sanity Pod ====="
oc create -f sanity-crc.yaml

echo "\nTest pod is starting! Once running, use 'oc logs csi-sanity-crc' to get sanity output"

