package main

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/Seagate/seagate-exos-x-csi/pkg/node"
	"k8s.io/klog/v2"
)

var bind = flag.String("bind", fmt.Sprintf("unix:///var/run/%s/csi-node.sock", common.PluginName), "RPC bind URI (can be a UNIX socket path or any URI)")
var chroot = flag.String("chroot", "", "Chroot into a directory at startup (used when running in a container)")

func main() {
	klog.InitFlags(nil)
	klog.EnableContextualLogging(true)
	flag.Set("logtostderr", "true")
	// Parse command line options, verbosity level is adjusted after this call
	flag.Parse()

	if *chroot != "" {
		if err := syscall.Chroot(*chroot); err != nil {
			panic(err)
		}
	}

	klog.InfoS("starting node controller", "version", common.Version)
	node.New().Start(*bind)
}
