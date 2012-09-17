package main

import (
	"github.com/bmizerany/pat"
	"log"
	"net/http"
)

func main() {
	m := pat.New()
	m.Post("/resources", http.HandlerFunc(AddInstance))
	m.Del("/resources/:name", http.HandlerFunc(RemoveInstance))
	log.Fatal(http.ListenAndServe(":3333", m))
}