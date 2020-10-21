package app

import (
	flag "github.com/spf13/pflag"
	"os"
)

type GossipPropagationFlags struct {
	MyDevice            string
	Port                int
	EtcdHost            string
	EtcdPort            int
	Join                bool
	MonitorDir          string
	IP                  string
	DebugMode           bool
	MyIP                string
	GossipNodeNum       int
	ConsystencyInterval int
}

const (
	defaultListenPort          = 10039
	defaultEtcdHost            = "localhost"
	defaultListenEtcdPort      = 30079
	defaultJoinFlag            = false
	defaultMonitorDir          = "/var/local/distributed-service-discovery"
	defaultDebugMode           = false
	defaultGossipNodeNum       = 3
	defaultConsistencyInterval = 900
	packageName                = "gossip"
	componentGossipPropagation = "gossip-propagation-d"
)

func newGossipPropagationFlags() *GossipPropagationFlags {
	defaultMyDevice, _ := os.Hostname()
	return &GossipPropagationFlags{
		MyDevice:            defaultMyDevice,
		Port:                defaultListenPort,
		EtcdHost:            defaultEtcdHost,
		EtcdPort:            defaultListenEtcdPort,
		Join:                defaultJoinFlag,
		MonitorDir:          defaultMonitorDir,
		DebugMode:           defaultDebugMode,
		GossipNodeNum:       defaultGossipNodeNum,
		ConsystencyInterval: defaultConsistencyInterval,
	}
}

func (f *GossipPropagationFlags) set(fs *flag.FlagSet) {
	fs.StringVarP(&f.MyDevice, "deviceName", "n", f.MyDevice, "my device name")

	fs.IntVarP(&f.Port, "port", "p", f.Port, "listen port")
	fs.StringVar(&f.EtcdHost, "etcdhost", f.EtcdHost, "etcd host")
	fs.IntVar(&f.EtcdPort, "etcdport", f.EtcdPort, "etcd listen port")

	// TODO: if we set "--join false", then we get true from GossipPropagationFlags.Join. this should be fixed.
	fs.BoolVarP(&f.Join, "join", "j", f.Join, "joining an existing cluster or starting a new cluster")
	fs.StringVarP(&f.IP, "ip", "i", f.IP, "using service discovery or manually setting device ip within cluster")

	fs.StringVar(&f.MyIP, "myip", f.IP, "my ip to use communication between nodes")

	fs.BoolVarP(&f.DebugMode, "debug", "d", f.DebugMode, "whether output debug log or not")

	// TODO: should get from db(etcd, titania...)
	fs.StringVarP(&f.MonitorDir, "monitor", "m", f.MonitorDir, "directory path to monitor for service discovery")

	fs.IntVarP(&f.GossipNodeNum, "gossip", "g", f.GossipNodeNum, "gossip nodes")
	fs.IntVarP(&f.ConsystencyInterval, "consistency", "c", f.ConsystencyInterval, "consystency interval")
}
