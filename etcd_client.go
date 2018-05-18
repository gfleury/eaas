package main

import (
	"crypto/tls"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
)

type EtcdClient struct {
	client       *clientv3.Client
	tlsConfig    *tls.Config
	etcdUsername string
	etcdPassword string
	hosts        []string
}

func (e *EtcdClient) newEtcdV3Client() error {
	var err error
	e.hosts = strings.Split(coalesceEnv("ETCD_URI", "127.0.0.1:2379"), ",")
	e.etcdUsername = coalesceEnv("ETCD_USERNAME", "root")
	e.etcdPassword = coalesceEnv("ETCD_PASSWORD", "123")

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

	e.client, err = e.connection()

	return err
}

func (e *EtcdClient) connection() (*clientv3.Client, error) {
	var err error
	if e.client == nil {
		e.client, err = clientv3.New(clientv3.Config{
			Endpoints: e.hosts,
			Username:  e.etcdUsername,
			Password:  e.etcdPassword,
			TLS:       e.tlsConfig,
		})
	}
	return e.client, err
}
