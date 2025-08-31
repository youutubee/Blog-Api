package main

import (
	"fmt"
	router "goblogapi/Router"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Welcome to a blog api in Golang")
	log.Fatal(http.ListenAndServe(":8082", router.Router()))
}
