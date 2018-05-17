package main

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"fmt"

	"github.com/coreos/etcd/auth/authpb"
	"github.com/coreos/etcd/clientv3"
)

var etcdSess *EtcdClient

// dbBind represents a bind stored in the database.
type dbBind struct {
	Name     string `bson:",omitempty"`
	User     string `bson:",omitempty"`
	AppHost  string `bson:",omitempty"`
	Password string `bson:",omitempty"`
}

type env map[string]string

var locker = multiLocker()

func etcdSession() *EtcdClient {
	if etcdSess == nil {
		etcdClientInstance := EtcdClient{}
		etcdSess = &etcdClientInstance
		err := etcdSess.newEtcdV3Client()
		if err != nil {
			panic(err)
		}
	}
	return etcdSess
}

func bind(name, appName string) (env, error) {
	locker.Lock(name)
	defer locker.Unlock(name)
	bind, err := newBind(name, appName)
	if err != nil {
		return nil, err
	}
	hosts := coalesceEnv("ETCD_URI", "127.0.0.1:2379")

	secretPath := coalesceEnv("ETCD_SECRET_PATH", "/domain/secret/%s")
	configPath := coalesceEnv("ETCD_CONFIG_PATH", "/domain/config/%s")

	data := map[string]string{
		"ETCD_HOSTS":              hosts,
		"ETCD_USER":               bind.User,
		"ETCD_PASSWORD":           bind.Password,
		"ETCD_APP_SCHEMA_PATH":    fmt.Sprintf(configPath, appName),
		"ETCD_SECRET_SCHEMA_PATH": fmt.Sprintf(secretPath, appName),
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
	item := dbBind{AppHost: appName, User: username, Name: name, Password: password}
	err = collection().Insert(item)
	if err != nil {
		return dbBind{}, err
	}
	return item, nil
}

func addUser(name, username, password string) error {
	_, err := etcdSession().client.UserAdd(context.TODO(), username, password)
	if err != nil {
		return err
	}

	_, err = etcdSession().client.RoleAdd(context.TODO(), username)
	if err != nil {
		etcdSession().client.UserDelete(context.TODO(), username)
		return err
	}

	secretPath := coalesceEnv("ETCD_SECRET_PATH", "/domain/secret/%s")
	configPath := coalesceEnv("ETCD_CONFIG_PATH", "/domain/config/%s")

	_, err = etcdSession().client.RoleGrantPermission(context.TODO(), username, fmt.Sprintf(configPath, name), clientv3.GetPrefixRangeEnd(fmt.Sprintf(configPath, name)), clientv3.PermissionType(authpb.Permission_Type_value["read"]))
	if err != nil {
		etcdSession().client.UserDelete(context.TODO(), username)
		etcdSession().client.RoleDelete(context.TODO(), username)
		return err
	}

	_, err = etcdSession().client.RoleGrantPermission(context.TODO(), username, fmt.Sprintf(secretPath, name), clientv3.GetPrefixRangeEnd(fmt.Sprintf(secretPath, name)), clientv3.PermissionType(authpb.Permission_Type_value["read"]))
	if err != nil {
		etcdSession().client.UserDelete(context.TODO(), username)
		etcdSession().client.RoleDelete(context.TODO(), username)
		return err
	}

	_, err = etcdSession().client.UserGrantRole(context.TODO(), username, username)
	if err != nil {
		etcdSession().client.UserDelete(context.TODO(), username)
		etcdSession().client.RoleDelete(context.TODO(), username)
		return err
	}

	_, err = etcdSession().client.Put(context.TODO(), fmt.Sprintf(configPath, name)+"/hello", "Hello World, access granted!")

	return err
}

func unbind(name, appName string) error {
	locker.Lock(name)
	defer locker.Unlock(name)
	coll := collection()
	bind := dbBind{Name: name, AppHost: appName}
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
	_, err := etcdSession().client.UserDelete(context.TODO(), username)
	if err != nil {
		return err
	}
	_, err = etcdSession().client.RoleDelete(context.TODO(), username)
	return err
}

func newPassword() string {
	var random [32]byte
	rand.Read(random[:])
	h := sha512.New()
	return fmt.Sprintf("%x", h.Sum(random[:]))
}
