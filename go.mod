module github.com/Seagate/seagate-exos-x-csi

go 1.20

require (
	github.com/Seagate/csi-lib-iscsi v1.1.0
	github.com/Seagate/csi-lib-sas v1.0.2
	github.com/Seagate/seagate-exos-x-api-go/v2 v2.2.0
	github.com/container-storage-interface/spec v1.8.0
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/kubernetes-csi/csi-test v0.0.0-20191016154743-6931aedb3df0
	github.com/onsi/gomega v1.28.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.17.0
	golang.org/x/sync v0.3.0
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.33.0
	k8s.io/klog/v2 v2.100.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/namsral/flag v1.7.4-pre // indirect
	github.com/nxadm/tail v1.4.4 // indirect
	github.com/onsi/ginkgo v1.12.1 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//replace github.com/Seagate/seagate-exos-x-api-go/v2 => ./seagate-exos-x-api-go
// replace github.com/Seagate/csi-lib-iscsi => ../csi-lib-iscsi
// replace github.com/Seagate/csi-lib-sas => ../csi-lib-sas
