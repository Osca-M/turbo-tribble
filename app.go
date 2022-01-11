package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func apiRoot(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "Welcome to our API Root!")
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", apiRoot)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	handleRequests()
}
