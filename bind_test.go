package main

import (
	"os"

	"gopkg.in/check.v1"
)

func (s *S) TestBindWithCA(c *check.C) {
	os.Setenv("ETCD_CA_ROOT", "test cert env")
	os.Setenv("ETCD_URI", "https://localhost:3939")
	envs, err := bind("test", "test")
	c.Assert(err, check.IsNil)
	data := map[string]string{
		"ETCD_HOSTS":              "127.0.0.1:2379",
		"ETCD_USER":               "test",
		"ETCD_PASSWORD":           "test",
		"ETCD_APP_SCHEMA_PATH":    "/domain/secret/%s",
		"ETCD_SECRET_SCHEMA_PATH": "/domain/config/%s",
		"ETCD_CA_ROOT":            "test cert env",
	}
	c.Assert(envs, check.DeepEquals, data)
}

func (s *S) TestBindWithoutCA(c *check.C) {
	envs, err := bind("test", "test")
	c.Assert(err, check.IsNil)
	data := map[string]string{
		"ETCD_HOSTS":              "127.0.0.1:2379",
		"ETCD_USER":               "test",
		"ETCD_PASSWORD":           "test",
		"ETCD_APP_SCHEMA_PATH":    "/domain/secret/%s",
		"ETCD_SECRET_SCHEMA_PATH": "/domain/config/%s",
	}
	c.Assert(envs, check.DeepEquals, data)
}
