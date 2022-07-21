#!/bin/bash

# Usage: crc-sanity.sh
#
# Launch cli sanity in a crc pod. Must be logged in with oc before running. 
# Runs all csi-sanity test cases

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

echo "Test pod is starting! Once running, use 'oc logs csi-sanity-crc' to get sanity output"

counter=0
success=0
continue=1
maxattempts=6

while [ $continue ]
do
    echo ""
    echo "Waiting for test pod to come online"

    testpodstatus=$(oc get pod csi-sanity-crc -o=jsonpath='{.status.phase}' -n seagate)
    echo $testpodstatus
    if [ "$testpodstatus" == "Running" ]; then
        echo "SUCCESS: test pod running"
        success=1
        break
    fi

    if [[ "$counter" -eq $maxattempts ]]; then
        echo ""
        echo "ERROR: Max attempts ($maxattempts) reached and test pod is not running."
        echo ""
        oc get pod csi-sanity-crc
        break
    else
        sleep 5
    fi

    ((counter++))
done

log_output_file="csi-sanity.log"
if [ "$success" -ne 0 ]; then
    echo "csi sanity in progress, logs will be tailed to $log_output_file"
    oc logs csi-sanity-crc -f > $log_output_file &
fi
