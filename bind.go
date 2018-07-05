package main

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"net/url"

	"github.com/coreos/etcd/auth/authpb"
	"github.com/coreos/etcd/clientv3"
)

var etcdSess *EtcdClient

// dbBind represents a bind stored in the database.
type dbBind struct {
	*BindAppForm
	Name     string `bson:",omitempty"`
	User     string `bson:",omitempty"`
	Password string `bson:",omitempty"`
}

type env map[string]string

var locker = multiLocker()

func etcdSession() (*EtcdClient, error) {
	var err error
	if etcdSess == nil {
		etcdClientInstance := EtcdClient{}
		etcdSess = &etcdClientInstance
		err = etcdSess.newEtcdV3Client()
	}

	if etcdSess.client == nil {
		_, err = etcdSess.connection()
	}
	return etcdSess, err
}

func bind(name, appName string) (env, error) {
	locker.Lock(name)
	defer locker.Unlock(name)
	bind, err := newBind(name, appName)
	if err != nil {
		return nil, err
	}
	hosts := coalesceEnv("ETCD_URI", "127.0.0.1:2379")

	u, err := url.Parse(hosts)
	if err == nil {
		hosts = u.Hostname()
	}

	caCert := coalesceEnv("ETCD_CA_ROOT", "")

	secretPath := coalesceEnv("ETCD_SECRET_PATH", "/domain/secret/%s")
	configPath := coalesceEnv("ETCD_CONFIG_PATH", "/domain/config/%s")

	data := map[string]string{
		"ETCD_HOSTS":              hosts,
		"ETCD_USER":               bind.User,
		"ETCD_PASSWORD":           bind.Password,
		"ETCD_APP_SCHEMA_PATH":    fmt.Sprintf(configPath, appName),
		"ETCD_SECRET_SCHEMA_PATH": fmt.Sprintf(secretPath, appName),
	}

	if len(caCert) > 0 {
		data["ETCD_CA_ROOT"] = caCert
	}

	return env(data), nil
}

func newBind(name, appName string) (dbBind, error) {
	password := newPassword()
	username := appName + newPassword()[:8]
	err := addUser(appName, username, password)
	if err != nil {
		return dbBind{}, err
	}
	item := dbBind{
		BindAppForm: &BindAppForm{
			AppName: appName,
		},
		User:     username,
		Name:     name,
		Password: password,
	}
	err = collection().Insert(item)
	if err != nil {
		return dbBind{}, err
	}
	return item, nil
}

func addUser(name, username, password string) error {
	session, err := etcdSession()
	if err != nil {
		return err
	}

	_, err = session.client.UserAdd(context.TODO(), username, password)
	if err != nil {
		return err
	}
	_, err = session.client.RoleAdd(context.TODO(), username)
	if err != nil {
		session.client.UserDelete(context.TODO(), username)
		return err
	}

	secretPath := coalesceEnv("ETCD_SECRET_PATH", "/domain/secret/%s")
	configPath := coalesceEnv("ETCD_CONFIG_PATH", "/domain/config/%s")

	_, err = session.client.RoleGrantPermission(context.TODO(), username, fmt.Sprintf(configPath, name), clientv3.GetPrefixRangeEnd(fmt.Sprintf(configPath, name)), clientv3.PermissionType(authpb.Permission_Type_value["read"]))
	if err != nil {
		session.client.UserDelete(context.TODO(), username)
		session.client.RoleDelete(context.TODO(), username)
		return err
	}

	_, err = session.client.RoleGrantPermission(context.TODO(), username, fmt.Sprintf(secretPath, name), clientv3.GetPrefixRangeEnd(fmt.Sprintf(secretPath, name)), clientv3.PermissionType(authpb.Permission_Type_value["read"]))
	if err != nil {
		session.client.UserDelete(context.TODO(), username)
		session.client.RoleDelete(context.TODO(), username)
		return err
	}

	_, err = session.client.UserGrantRole(context.TODO(), username, username)
	if err != nil {
		session.client.UserDelete(context.TODO(), username)
		session.client.RoleDelete(context.TODO(), username)
		return err
	}

	_, err = session.client.Put(context.TODO(), fmt.Sprintf(configPath, name)+"/hello", "Hello World, access granted!")

	return err
}

func unbind(name, appName string) error {
	locker.Lock(name)
	defer locker.Unlock(name)
	coll := collection()
	bind := dbBind{
		BindAppForm: &BindAppForm{
			AppName: appName,
		},
		Name: name,
	}
	err := coll.Find(bind).One(&bind)
	if err != nil {
		return err
	}
	err = coll.Remove(bind)
	if err != nil {
		return err
	}
	return removeUser(bind.User)
}

func removeUser(username string) error {
	session, err := etcdSession()
	if err != nil {
		return err
	}

	_, err = session.client.UserDelete(context.TODO(), username)
	if err != nil {
		return err
	}
	_, err = session.client.RoleDelete(context.TODO(), username)
	return err
}

func newPassword() string {
	var random [32]byte
	rand.Read(random[:])
	h := sha512.New()
	return fmt.Sprintf("%x", h.Sum(random[:]))
}
