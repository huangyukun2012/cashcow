package main
import (
//	"fmt"
//	"bytes"
//	"strconv"
//	"strings"
//	"io/ioutil"
//	"net/http"
//	"logging"
//	"encoding/json"
	//t "developerq/trans"
	//m "developerq/model"
	g "developerq/ghcrawler"
	//u "developerq/utils"
	"os"
)


func main() {
	arg := os.Args[1]
	g.Init()
	if arg == "url" {
		//arg2 := os.Args[2]
		g.CrawlGHURL(20000)
	}

}
