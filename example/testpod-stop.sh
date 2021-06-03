#!/bin/bash

source common.sh

echo "[] testpod-stop"

banner "Delete Resources"

runCommand "kubectl delete deployment exosx-controller-server"
runCommand "kubectl delete daemonsets exosx-node-server"
runCommand "kubectl delete serviceaccounts csi-provisioner"
runCommand "kubectl delete configmaps init-node"
runCommand "kubectl delete clusterrole external-provisioner-runner-exosx"
runCommand "kubectl delete ClusterRoleBinding csi-provisioner-role-exosx"
runCommand "kubectl delete role external-provisioner-cfg-exosx"
runCommand "kubectl delete rolebinding csi-provisioner-role-cfg-exosx"

runCommand "helm uninstall test-release"
runCommand "kubectl delete pods test-pod --grace-period=0 --force"
runCommand "kubectl delete secrets exosx-secrets"
runCommand "kubectl delete pvc exosx-pvc"
runCommand "kubectl delete sc exosx-storageclass"
runCommand "kubectl delete pods --all"

banner "Check Resources"

runCommand "kubectl get configmaps"
runCommand "kubectl get serviceaccounts"
runCommand "kubectl get daemonsets"
runCommand "kubectl get deployments"
runCommand "kubectl get pvc"
runCommand "kubectl get pvc"
runCommand "kubectl get sc"
runCommand "kubectl get secrets"
runCommand "kubectl get pods"
