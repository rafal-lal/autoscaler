package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/power-manager/server"
	kube_flag "k8s.io/component-base/cli/flag"
	klog "k8s.io/klog/v2"
)

// MultiStringFlag is a flag for passing multiple parameters using same flag
type MultiStringFlag []string

// String returns string representation of the node groups.
func (flag *MultiStringFlag) String() string {
	return "[" + strings.Join(*flag, " ") + "]"
}

// Set adds a new configuration.
func (flag *MultiStringFlag) Set(value string) error {
	*flag = append(*flag, value)
	return nil
}

func multiStringFlag(name string, usage string) *MultiStringFlag {
	value := new(MultiStringFlag)
	flag.Var(value, name, usage)
	return value
}

var (
	// flags needed by the external grpc provider service
	address = flag.String("address", ":8086", "The address to expose the grpc service.")
	keyCert = flag.String("key-cert", "",
		"The path to the certificate key file. Empty string for insecure communication.")
	cert = flag.String("cert", "",
		"The path to the certificate file. Empty string for insecure communication.")
	cacert = flag.String("ca-cert", "",
		"The path to the ca certificate file. Empty string for insecure communication.")

	cloudConfig = flag.String("cloud-config", "",
		"The path to the cloud provider configuration file. Empty string for no configuration file.")
	clusterName    = flag.String("cluster-name", "", "Autoscaled cluster name, if available")
	nodeGroupsFlag = multiStringFlag(
		"nodes",
		"sets min,max size and other configuration data for a node group in a format accepted by cloud provider. "+
			"Can be used multiple times. Format: <min>:<max>:<other...>")
)

func main() {
	klog.InitFlags(nil)
	kube_flag.InitFlags()

	var s *grpc.Server

	// tls config
	var serverOpt grpc.ServerOption
	if *keyCert == "" || *cert == "" || *cacert == "" {
		klog.V(1).Info("no cert specified, using insecure")
		s = grpc.NewServer()
	} else {

		certificate, err := tls.LoadX509KeyPair(*cert, *keyCert)
		if err != nil {
			klog.Fatalf("failed to read certificate files: %s", err)
		}
		certPool := x509.NewCertPool()
		bs, err := os.ReadFile(*cacert)
		if err != nil {
			klog.Fatalf("failed to read client ca cert: %s", err)
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			klog.Fatal("failed to append client certs")
		}
		transportCreds := credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{certificate},
			ClientCAs:    certPool,
		})
		serverOpt = grpc.Creds(transportCreds)
		s = grpc.NewServer(serverOpt)
	}

	// format is id:min:max:label,label...
	id := strings.Split((*nodeGroupsFlag)[0], ":")[0]
	min, _ := strconv.Atoi(strings.Split((*nodeGroupsFlag)[0], ":")[1])
	max, _ := strconv.Atoi(strings.Split((*nodeGroupsFlag)[0], ":")[2])
	nodeGroupLabels := strings.Split((*nodeGroupsFlag)[0], ":")[3]
	labels := strings.Split(nodeGroupLabels, ",")

	srv := server.NewBaremetalProvider(&server.BaremetalNodeGroupConfig{
		ID:          id,
		MinSizeConf: min,
		MaxSizeConf: max,
		Labels:      labels,
	})

	// listen
	lis, err := net.Listen("tcp", *address)
	if err != nil {
		klog.Fatalf("failed to listen: %s", err)
	}

	// serve
	protos.RegisterCloudProviderServer(s, srv)
	klog.V(1).Infof("Server ready at: %s\n", *address)
	if err := s.Serve(lis); err != nil {
		klog.Fatalf("failed to serve: %v", err)
	}

}
