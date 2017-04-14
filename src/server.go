package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	ds "developerq"
	bs "bilisou"
	u "utils"
	//"indexer"
)

func RedirectPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://www.developerq.com" + r.URL.Path, http.StatusMovedPermanently)
}

func InitMain() {

	//utils init
	u.LISTMAX = 300
	u.PAGEMAX = 20
	u.NAVMAX = 5
	u.RANDMAX = 10

	u.InitCateMap()
	u.InitJieba()
	u.InitRedis()

}

func main() {

	InitMain()
	//start indexer in backend
	//go indexer.Start()
	ds.Init()
	bs.Init()

	mx := mux.NewRouter()
	//dmx := mx.Host("localhost").Subrouter()
	//bmx := mx.Host("localhostb").Subrouter()
	dmx := mx.Host("developerq.com").Subrouter()
	ds.Start(dmx)
	dmx = mx.Host("www.developerq.com").Subrouter()
	ds.Start(dmx)

	dmx = mx.Host("wenti.info").Subrouter()
	dmx.NotFoundHandler = http.HandlerFunc(RedirectPage)

	dmx = mx.Host("www.wenti.info").Subrouter()
	dmx.NotFoundHandler = http.HandlerFunc(RedirectPage)

	bmx := mx.Host("bilisou.com").Subrouter()
	bs.Start(bmx)
	bmx = mx.Host("www.bilisou.com").Subrouter()
	bs.Start(bmx)


	fmt.Println("start listening at *:80")
	http.ListenAndServe("0.0.0.0:80", mx)
	select{}

}
