package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func Add(c *gin.Context) {
	r := c.Request
	w := c.Writer
	name := r.FormValue("name")
	if name == dbName() {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Reserved name")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func BindApp(c *gin.Context) {
	r := c.Request
	w := c.Writer
	name := c.Param("name")
	appHost := r.FormValue("app-host")
	if appHost == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing app-host")
		return
	}
	env, err := bind(name, appHost)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(env)
	if err != nil {
		c.AbortWithError(500, err)
	}
}

func BindUnit(c *gin.Context) {
	return
}

func UnbindApp(c *gin.Context) {
	r := c.Request
	w := c.Writer
	r.Method = "POST"
	name := c.Param("name")
	appHost := r.FormValue("app-host")
	err := unbind(name, appHost)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func UnbindUnit(c *gin.Context) {
	return
}

func Remove(c *gin.Context) {
	//r := c.Request
	w := c.Writer
	name := c.Param("name")
	collection().RemoveAll(bson.M{"name": name})
	err := session().DB(name).DropDatabase()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func Status(c *gin.Context) {
	//r := c.Request
	w := c.Writer
	if err := session().Ping(); err != nil {
		c.AbortWithError(500, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return
}

type httpError struct {
	code int
	body string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP error (%d): %s", e.code, e.body)
}
