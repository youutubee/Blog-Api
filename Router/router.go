package router

import (
	controller "goblogapi/Controller"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/signup", controller.CreateUser).Methods("POST")
	return r
}
