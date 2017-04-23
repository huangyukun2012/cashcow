package bt
import (
	"golang.org/x/net/html"
	"github.com/Unknwon/goconfig"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
//	"time"
	"regexp"
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"io/ioutil"
//	"io"
//	"net/http"
	"logging"
//	"encoding/json"
//	t "developerq/trans"
	m "bilisou/model"
	u "utils"
	"os"
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
	logSvc.ConfigDefaultLogger("/tmp/bilisou", "bt", logging.INFO, logging.ROTATE_DAILY)
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

func FindShareFilenames(n *html.Node, share *m.Share) {
	if n.Type == html.ElementNode && n.Data == "tr" {
		b := new(bytes.Buffer)
		if err := html.Render(b, n); err != nil {
			Logger.Error(err.Error())
		}
		tr := b.String()
		if !strings.Contains(tr, "<th>文件名称</th>") {
			var digitsRegexp = regexp.MustCompile(`<td>(.*)<\/td>[\r\n\s]+<td>[\r\n\s]+(.*)<\/td>`)
			res := digitsRegexp.FindStringSubmatch(tr)
			fmt.Printf("res len = %s\n", len(res))
			if len(res) > 2 {
				share.FilenamesRaw = share.FilenamesRaw + res[1] + "s#i#z#e#" + res[2]
				share.FilenamesRaw = share.FilenamesRaw + "b#i#l#i#s#o#u#"
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindShareFilenames(c, share)
	}
}

func FindShare(n *html.Node, share *m.Share) {
	if n.Type == html.ElementNode && n.Data == "h2" {


		Logger.Info("find title")
		if c := n.FirstChild; c != nil && c.Data != "" {
			b := new(bytes.Buffer)
			if err := html.Render(b, c); err != nil {
				Logger.Error(err.Error())
			}
			share.Title = b.String()
			share.Title = strings.TrimSpace(share.Title)
			fmt.Println("title = " + share.Title)
		}

	}

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" && strings.Contains(a.Val, "magnet") {
				share.Link = a.Val
				fmt.Println("Link = " + share.Link)
			}
		}
	}

	if n.Type == html.ElementNode && n.Data == "table" {
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "am-table am-table-bordered am-table-radius am-table-striped" {

				//FindImageAndDownload(n, share)
				b := new(bytes.Buffer)
				if err := html.Render(b, n); err != nil {
					Logger.Error(err.Error())
				}
				table := b.String()
				if strings.Contains(table, "<th>文件名称</th>") {
					FindShareFilenames(n, share)
					fmt.Println("filename = " + share.FilenamesRaw)
				} else {
					rp := regexp.MustCompile(`<td>文件大小<\/td>[\r\n\s]+<td>(.*)<\/td>[\r\n\s]+<\/tr>[\r\n\s]+<tr>[\r\n\s]+<td>创建时间<\/td>`)
					res := rp.FindStringSubmatch(table)
					if len(res) > 1 {
						share.SizeStr = res[1]
					}
					fmt.Println("size = " + share.SizeStr)
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
		if current > 87000 {
			return
		}
		current = current + 1
		stmt, _ := db.Prepare("update bttrack set current = ?")
		stmt.Exec(current)
		stmt.Close()

		filename := path + strconv.Itoa(current) + ".html"

		fmt.Println("process bt file = " + filename)

		if _, err := os.Stat(filename); err == nil {
			// path/to/whatever exists
			doc, err := GetHTMLNodeFromFile(filename)
			if err != nil {
				Logger.Error(err.Error())
				continue
			}

			share := m.Share{}
			share.Source = 2
			share.CategoryInt = 1
			share.Valid = true
			FindShare(doc, &share)
			share.CategoryInt = int64(u.GetCategoryFromName(share.Title))
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