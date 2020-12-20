package main

import (
	"log"
	"net/http"

	methods "github.com/M1K8/internal/api"

	"github.com/gorilla/mux"
)

func main() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/authorize", methods.Auth)

	myRouter.HandleFunc("/void", methods.Void)

	myRouter.HandleFunc("/capture", methods.Capture)

	myRouter.HandleFunc("/refund", methods.Refund)

	log.Fatal(http.ListenAndServe(":1337", myRouter))
}
