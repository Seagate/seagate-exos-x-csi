package main

import (
	"flag"
	"fmt"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/controller"
	"k8s.io/klog"
)

var bind = flag.String("bind", fmt.Sprintf("unix:///var/run/%s/csi-controller.sock", common.PluginName), "RPC bind URI (can be a UNIX socket path or any URI)")

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()
	klog.Infof("starting storage controller (%s)", common.Version)
	controller.New().Start(*bind)
}
