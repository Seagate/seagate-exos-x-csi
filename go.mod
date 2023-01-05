module github.com/Seagate/seagate-exos-x-csi

go 1.19

require (
	github.com/Seagate/csi-lib-iscsi v1.0.3
	github.com/Seagate/csi-lib-sas v1.0.2
	github.com/Seagate/seagate-exos-x-api-go v1.0.11
	github.com/container-storage-interface/spec v1.4.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/kubernetes-csi/csi-test v0.0.0-20191016154743-6931aedb3df0
	github.com/onsi/gomega v1.10.5
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.13.0
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	google.golang.org/grpc v1.50.0
	k8s.io/klog/v2 v2.80.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/nxadm/tail v1.4.4 // indirect
	github.com/onsi/ginkgo v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591 // indirect
	golang.org/x/sys v0.0.0-20221010170243-090e33056c14 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20221010155953-15ba04fc1c0e // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// replace github.com/Seagate/seagate-exos-x-api-go => ../seagate-exos-x-api-go
// replace github.com/Seagate/csi-lib-iscsi => ../csi-lib-iscsi
// replace github.com/Seagate/csi-lib-sas => ../csi-lib-sas
