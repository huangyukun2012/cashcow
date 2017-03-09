package main

import (
	"github.com/gorilla/mux"
	"net/http"
	ds "developerq"
	bs "bilisou"
	//"indexer"
)


func main() {

	//start indexer in backend
	//go indexer.Start()

	mx := mux.NewRouter()
	dmx := mx.Host("localhost").Subrouter()
	bmx := mx.Host("localhostb").Subrouter()

	//start developerq
	ds.Start(dmx)
	//start bilisou
	bs.Start(bmx)
	http.ListenAndServe("0.0.0.0:8080", mx)


}
