package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"

	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

var _ = check.Suite(&S{})

type S struct {
	muxer http.Handler
}

func Test(t *testing.T) { check.TestingT(t) }

func (s *S) SetUpSuite(c *check.C) {
	s.muxer = buildMux()
}

func (s *S) TearDownSuite(c *check.C) {
	session().DB(dbName()).DropDatabase()
}

func (s *S) SetUpTest(c *check.C) {
	collection().RemoveAll(nil)
}

func (s *S) TestAdd(c *check.C) {
	body := strings.NewReader("name=something")
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
}

func (s *S) TestAddReservedName(c *check.C) {
	name := dbName()
	body := strings.NewReader("name=" + name)
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusForbidden)
	c.Assert(recorder.Body.String(), check.Equals, "Reserved name")
}

func (s *S) TestBindShouldReturnLocalhostWhenThePublicHostEnvIsNil(c *check.C) {
	body := strings.NewReader("app-host=localhost")
	request, err := http.NewRequest("POST", "/resources/myapp/bind-app", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	os.Setenv("MONGODB_URI", "")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["ETCD_HOSTS"], check.Equals, "127.0.0.1:2379")
	c.Assert(data["ETCD_APP_SCHEMA_PATH"], check.Equals, "/domain/config/myapp")
	c.Assert(data["ETCD_SECRET_SCHEMA_PATH"], check.Equals, "/domain/secret/myapp")
	c.Assert(data["ETCD_USER"], check.Not(check.HasLen), 0)
	c.Assert(data["ETCD_PASSWORD"], check.Not(check.HasLen), 0)
	coll := collection()
	expected := dbBind{
		AppHost:  "localhost",
		Name:     "myapp",
		User:     data["ETCD_USER"],
		Password: data["ETCD_PASSWORD"],
	}
	var bind dbBind
	q := bson.M{"name": "myapp"}
	defer coll.Remove(q)
	err = coll.Find(q).One(&bind)
	c.Assert(err, check.IsNil)
	c.Assert(bind, check.DeepEquals, expected)
}

func (s *S) TestBind(c *check.C) {
	body := strings.NewReader("app-host=localhost")
	request, err := http.NewRequest("POST", "/resources/myapp/bind-app", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	publicHost := "127.0.0.1:2379"
	os.Setenv("ETCD_URI", publicHost)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	defer func() {
		database := session().DB("myapp")
		database.RemoveUser("myapp")
		database.DropDatabase()
		collection().Remove(bson.M{"name": "myapp"})
	}()
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
	result, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	data := map[string]string{}
	json.Unmarshal(result, &data)
	c.Assert(data["ETCD_HOSTS"], check.Equals, publicHost)
	c.Assert(data["ETCD_APP_SCHEMA_PATH"], check.Equals, "/domain/config/myapp")
	c.Assert(data["ETCD_SECRET_SCHEMA_PATH"], check.Equals, "/domain/secret/myapp")
	c.Assert(data["ETCD_USER"], check.Not(check.HasLen), 0)
	c.Assert(data["ETCD_PASSWORD"], check.Not(check.HasLen), 0)
	tlsInfo := transport.TLSInfo{
		CertFile:           "",
		KeyFile:            "",
		TrustedCAFile:      "",
		InsecureSkipVerify: true,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	session, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(data["ETCD_HOSTS"], ","),
		Username:  data["ETCD_USER"],
		Password:  data["ETCD_PASSWORD"],
		TLS:       tlsConfig,
	})
	c.Assert(err, check.IsNil)
	defer session.Close()
	_, err = session.Get(context.TODO(), data["ETCD_APP_SCHEMA_PATH"]+"/hello")
	c.Assert(err, check.IsNil)
}

func (s *S) TestBindNoAppHost(c *check.C) {
	body := strings.NewReader("")
	request, err := http.NewRequest("POST", "/resources/myapp/bind-app", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, "Missing app-host")
}

func (s *S) TestUnbind(c *check.C) {
	name := "myapp"
	data, err := bind(name, "localhost")
	c.Assert(err, check.IsNil)
	defer func() {
		database := session().DB(name)
		database.DropDatabase()
	}()
	body := strings.NewReader("app-host=localhost")
	request, err := http.NewRequest("DELETE", "/resources/myapp/bind-app", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	tlsInfo := transport.TLSInfo{
		CertFile:           "",
		KeyFile:            "",
		TrustedCAFile:      "",
		InsecureSkipVerify: true,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	_, err = clientv3.New(clientv3.Config{
		Endpoints: strings.Split(data["ETCD_HOSTS"], ","),
		Username:  data["ETCD_USER"],
		Password:  data["ETCD_PASSWORD"],
		TLS:       tlsConfig,
	})
	c.Assert(err, check.NotNil)
}

func (s *S) TestBindUnit(c *check.C) {
	request, err := http.NewRequest("POST", "/resources/myapp/bind", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (s *S) TestUnbindUnit(c *check.C) {
	request, err := http.NewRequest("DELETE", "/resources/myapp/bind", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (s *S) TestRemoveShouldRemoveBinds(c *check.C) {
	name := "myapp"
	collection().Insert(dbBind{Name: name})
	database := session().DB(name)
	database.AddUser(name, "", false)
	request, err := http.NewRequest("DELETE", "/resources/myapp", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	count, err := collection().Find(bson.M{"name": name}).Count()
	c.Assert(err, check.IsNil)
	c.Assert(count, check.Equals, 0)
}

func (s *S) TestStatus(c *check.C) {
	name := "myapp"
	database := session().DB(name)
	database.AddUser(name, "", false)
	defer func() {
		database.RemoveUser("myapp")
		database.DropDatabase()
	}()
	request, err := http.NewRequest("GET", "/resources/myapp/status", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNoContent)
}

func errorHandler(w http.ResponseWriter, r *http.Request) error {
	return errors.New("some error")
}

func httpErrorHandler(w http.ResponseWriter, r *http.Request) error {
	return &httpError{code: 400, body: "please provide a name"}
}

func simpleHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprint(w, "success")
	return nil
}

func (s *S) TestHandlerReturns500WhenInternalHandlerReturnsAnError(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, check.IsNil)
	Handler(errorHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 500)
	c.Assert(recorder.Body.String(), check.Equals, "some error\n")
}

func (s *S) TestHandlerWithHTTPError(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, check.IsNil)
	Handler(httpErrorHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 400)
	c.Assert(recorder.Body.String(), check.Equals, "please provide a name\n")
}

func (s *S) TestHandlerShouldPassAnHandlerWithoutError(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/apps", nil)
	c.Assert(err, check.IsNil)
	Handler(simpleHandler).ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	c.Assert(recorder.Body.String(), check.Equals, "success")
}

func (s *S) TestHTTPError(c *check.C) {
	var err error = &httpError{code: 404, body: "not found"}
	c.Assert(err.Error(), check.Equals, "HTTP error (404): not found")
}
