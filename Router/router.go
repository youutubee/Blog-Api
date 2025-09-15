package router

import (
	controller "goblogapi/Controller"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/signup", controller.CreateUser).Methods("POST")
	r.HandleFunc("/login", controller.LoginOneUser).Methods("POST")
	r.HandleFunc("/insertblog", controller.InsertOneBlog).Methods("POST")
	return r
}
