package bili
import (
	"golang.org/x/net/html"
	"github.com/Unknwon/goconfig"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
//	"time"
//	"regexp"
	"fmt"
	"bytes"
//	"strconv"
	"strings"
	"io/ioutil"
//	"io"
	"net/http"
	"logging"
//	"encoding/json"
//	t "developerq/trans"
	m "bilisou/model"
//	u "utils"
	//"os"
)

var Logger *logging.Logger



var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var cfg *goconfig.ConfigFile

//Mysql
func Init(dbc *sql.DB) {

	logSvc := logging.NewLogServcie()
	logSvc.ConfigDefaultLogger("/tmp/bilisou", "bili", logging.INFO, logging.ROTATE_DAILY)
	logSvc.Serve()
	//defer logSvc.Stop()
	Logger = logSvc.GetLogger("default")
	//m.Logger = Logger
	db = dbc

}


func GetHTMLNodeFromFile(filename string)( *html.Node, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	doc, err := html.Parse(strings.NewReader(string(buf)))
	return doc, err
}

func GetHTMLNodeFromURL(url string)( *html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return nil, err
	}
	doc, err := html.Parse(strings.NewReader(string(body)))
	return doc, err
}


func FindTag(n *html.Node, share *m.Share) {
	if n.Type == html.ElementNode && n.Data == "a" {
		Logger.Info("find tag")
		if c := n.FirstChild; c != nil && c.Data != "" {
			share.FilenamesRaw = share.FilenamesRaw + c.Data + "#$#$"
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindShare(c, share)
	}

}

func FindShare(n *html.Node, share *m.Share) {
	if n.Type == html.ElementNode && n.Data == "h1" {
		Logger.Info("find title")
		if c := n.FirstChild; c != nil && c.Data != "" {
			b := new(bytes.Buffer)
			if err := html.Render(b, c); err != nil {
				Logger.Error(err.Error())
			}
			share.Title = b.String()
			share.Title = strings.TrimSpace(share.Title)
		}

	}


	if n.Type == html.ElementNode && n.Data == "div" {
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == "v_desc" {
				//FindImageAndDownload(n, share)
				b := new(bytes.Buffer)
				if err := html.Render(b, n); err != nil {
					Logger.Error(err.Error())
				}
				share.Description = b.String()
				//fmt.Println("desc = " + share.Description)

			}
		}
	}


	if n.Type == html.ElementNode && n.Data == "meta" {
		for _, a := range n.Attr {
			fmt.Println(a)
			if a.Key == "name"  && a.Val == "keywords" {
				for _, b := range n.Attr {
					if b.Key == "content" {
						share.FilenamesRaw = b.Val
						fmt.Println("share tag = " + share.FilenamesRaw)
					}
				}
			}

			if a.Key == "name"  && a.Val == "description" {
				for _, b := range n.Attr {
					if b.Key == "content" {
						share.Description = b.Val
						fmt.Println("share tag = " + share.Description)
					}
				}
			}
		}
	}


	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindShare(c, share)
	}
}


func CrawlShare() {
	for  {
		sql := "select path, current from bttrack"
		rows, err := db.Query(sql)
		defer rows.Close()
		if err != nil {
			Logger.Error(err.Error())
			return
		}
		var path string
		var current int
		for rows.Next() {
			rows.Scan(&path, &current)
		}

		current = current + 1
		stmt, _ := db.Prepare("update bttrack set current = ?")
		stmt.Exec(current)
		stmt.Close()

		///filename := path + strconv.Itoa(current) + ".html"
		url := fmt.Sprintf("http://www.bilibili.com/video/av%d/", current)

		Logger.Info("process bt url = " + url)

		if true {
			// path/to/whatever exists
			doc, err := GetHTMLNodeFromURL(url)
			if err != nil {
				Logger.Error(err.Error())
				continue
			}

			share := m.Share{}
			share.Source = 4
			//share.CategoryInt = 1
			share.Valid = true
			share.ShareID = fmt.Sprintf("%d", current)
			FindShare(doc, &share)
			share.CategoryInt = 2
			if share.Valid {
				share.FillAll()
				err = share.Save(db)
				if err != nil {
					Logger.Error(err.Error())
				}
			}
		}
	}
}

func Start(dbc *sql.DB) {
	Init(dbc)
	CrawlShare()
}
