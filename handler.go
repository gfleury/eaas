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
	appName := r.FormValue("app-name")
	if appName == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing app-name")
		return
	}
	env, err := bind(name, appName)
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

}

func UnbindApp(c *gin.Context) {
	r := c.Request
	w := c.Writer
	r.Method = "POST"
	name := c.Param("name")
	appName := r.FormValue("app-name")
	err := unbind(name, appName)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func UnbindUnit(c *gin.Context) {

}

func Remove(c *gin.Context) {
	//r := c.Request
	w := c.Writer
	name := c.Param("name")
	collection().RemoveAll(bson.M{"name": name})
	w.WriteHeader(http.StatusOK)
}

func Status(c *gin.Context) {
	//r := c.Request
	w := c.Writer
	if err := session().Ping(); err != nil {
		c.AbortWithError(http.StatusAccepted, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ListPlans(c *gin.Context) {
	plans := []Plan{
		{
			Description: "Default ETCD Plan",
			Name:        "default",
		},
	}
	c.JSON(http.StatusOK, plans)
}

func UpdateServiceInstance(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})

}

func GetServiceInstance(c *gin.Context) {
	name := c.Param("name")
	if name == "plans" {
		ListPlans(c)
		return
	}
	c.JSON(http.StatusNotFound, []string{})
}
