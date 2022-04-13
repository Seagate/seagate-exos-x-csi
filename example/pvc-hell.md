# pvc-hell testing notes

## Initial Configuration
- Testing first with a single node Kubernetes clsuter just to validate the scripts and testing and see the results

- export helmpath=~/github.com/Seagate/seagate-exos-x-csi/helm/csi-charts/
- helm install seagate-csi $helmpath -f values.yaml
- kubectl get pods -n seagate
- kubectl create -f <secret.yaml>
- kubectl create -f <storageclass.yaml>


## Run Tests

```
./pvc-hell.sh apply 1
```

## Clean Up

./pvc-hell.sh delete 200


## Notes
- secret.yaml points to the correct storage array, this testing used Gallium
- storageclass.yaml used a name of 'my-marvelous-storage'