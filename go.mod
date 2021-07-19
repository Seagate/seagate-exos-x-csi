module github.com/Seagate/seagate-exos-x-csi

go 1.16

require (
	github.com/Seagate/csi-lib-iscsi v0.0.0-00010101000000-000000000000
	github.com/Seagate/seagate-exos-x-api-go v0.0.0-00010101000000-000000000000
	github.com/container-storage-interface/spec v1.4.0
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/kubernetes-csi/csi-test v0.0.0-20191016154743-6931aedb3df0
	github.com/onsi/gomega v1.10.5
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	google.golang.org/grpc v1.29.1
	k8s.io/klog v1.0.0
)

replace github.com/Seagate/seagate-exos-x-api-go => ../seagate-exos-x-api-go

replace github.com/Seagate/csi-lib-iscsi => ../csi-lib-iscsi
