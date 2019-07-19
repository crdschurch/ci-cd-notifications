package main

import (
	"net/http"
	"ci-cd-notifications/deploystatus"
)

func main() {
    http.HandleFunc("/", deploystatus.Handler)
    http.ListenAndServe(":8080", nil)
}