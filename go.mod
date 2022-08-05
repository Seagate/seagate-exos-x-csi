module github.com/Seagate/seagate-exos-x-csi

go 1.16

require (
	github.com/Seagate/csi-lib-iscsi v1.0.3
	github.com/Seagate/csi-lib-sas v1.0.1
	github.com/Seagate/seagate-exos-x-api-go v1.0.8-0.20220531203625-3d1a38b18ac6
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/container-storage-interface/spec v1.4.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/kubernetes-csi/csi-test v0.0.0-20191016154743-6931aedb3df0
	github.com/onsi/gomega v1.10.5
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/sys v0.0.0-20211124211545-fe61309f8881 // indirect
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
	google.golang.org/grpc v1.42.0
	k8s.io/klog v1.0.0
)

// replace github.com/Seagate/seagate-exos-x-api-go => ../seagate-exos-x-api-go
// replace github.com/Seagate/csi-lib-iscsi => ../csi-lib-iscsi
// replace github.com/Seagate/csi-lib-sas => ../csi-lib-sas
