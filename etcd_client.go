package main

import (
	"crypto/tls"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
)

type EtcdClient struct {
	client    *clientv3.Client
	tlsConfig *tls.Config
}

func (e *EtcdClient) newEtcdV3Client() error {
	var err error
	hosts := strings.Split(coalesceEnv("ETCD_URI", "127.0.0.1:2379"), ",")
	etcdUsername := coalesceEnv("ETCD_USERNAME", "root")
	etcdPassword := coalesceEnv("ETCD_PASSWORD", "123")

	tlsInfo := transport.TLSInfo{
		CertFile:           "",
		KeyFile:            "",
		TrustedCAFile:      "",
		InsecureSkipVerify: true,
	}

	e.tlsConfig, err = tlsInfo.ClientConfig()

	if err != nil {
		return err
	}

	e.client, err = clientv3.New(clientv3.Config{
		Endpoints: hosts,
		Username:  etcdUsername,
		Password:  etcdPassword,
		TLS:       e.tlsConfig,
	})

	return err
}
