package weixin
import (
	"golang.org/x/net/html"
	"github.com/Unknwon/goconfig"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"fmt"
	"bytes"
//	"strconv"
	"strings"
	"io/ioutil"
	"io"
	"net/http"
	"logging"
//	"encoding/json"
//	t "developerq/trans"
	m "developerq/model"
//	u "utils"
	"os"
)

var Logger *logging.Logger


const WEIXIN_SEARCH_URL = "http://weixin.sogou.com/weixin?usip=null&query=%s&from=tool&ft=&tsn=1&et=&interation=null&type=2&wxid=&page=%d&ie=utf8"
const BAIDU_SEARCH_URL = "http://www.baidu.com/s?ie=utf-8&f=3&rsv_bp=1&tn=baidu&wd=\"%s\""

var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var cfg *goconfig.ConfigFile
var githubusername, githubpassword string

//Mysql
func Init(dbc *sql.DB) {

	logSvc := logging.NewLogServcie()
	logSvc.ConfigDefaultLogger("/tmp/developerq", "weixin", logging.INFO, logging.ROTATE_DAILY)
	logSvc.Serve()
	//defer logSvc.Stop()
	Logger = logSvc.GetLogger("default")
	//m.Logger = Logger
	db = dbc

}

func CheckBlog(title string) bool {
	if strings.Contains(title, "聘") || strings.Contains(title, "<") ||
		strings.Contains(title, ">") {
		return false
	}
	return true

	/*

	url := fmt.Sprintf(BAIDU_SEARCH_URL, title)
	resp, err := http.Get(url)
	fmt.Println(url)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	text := string(body)
	fmt.Println(text)
	if strings.Contains(text, "很抱歉，没有找到与") {
		fmt.Println("found " + title)
		return true
	} else {
		fmt.Println("skip " + title)
		return false
	}
	*/
}

func DownloadImage(url string, filename string) error {
	res, err := http.Get(url)
	Logger.Info("Download image " + url + " " + filename)
	if err != nil {
		Logger.Info("failed to download, " + err.Error())
		return err
	}
	file, err := os.Create("imgpool/" + filename + ".jpg")
	if err != nil {
		Logger.Info("failed to create file" + err.Error())
		return err
	}
	io.Copy(file, res.Body)
	return nil
}


func FindImageAndDownload(n *html.Node, blog *m.Blog) {

	if n.Type == html.ElementNode && n.Data == "img" {
		Logger.Info("found image!!!")
		newurl := ""
		flag := true

		for i, a := range n.Attr {
			if a.Key == "data-src" {
				url := a.Val

				t := time.Now()
				filename := fmt.Sprintf("%d", int64(t.UnixNano()))
				err := DownloadImage(url, filename)
				if err != nil {
					return
				}
				newurl = "/imgpool/" + filename + ".jpg"
			}
			if a.Key == "src" {
				n.Attr[i].Val = newurl
				flag = false
			}
		}
		if flag {
			ta := html.Attribute{}
			ta.Key = "src"
			ta.Val = newurl
			n.Attr = append(n.Attr,ta)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindImageAndDownload(c, blog)
	}
}

func FindBlog(n *html.Node, blog *m.Blog) {
	if n.Type == html.ElementNode && n.Data == "h2" {
		for _, a := range n.Attr {
			if a.Key == "class" 	&&  a.Val == "rich_media_title" {
				Logger.Info("find title")
				if c := n.FirstChild; c.Data != "" {
					b := new(bytes.Buffer)
					if err := html.Render(b, c); err != nil {
						Logger.Error(err.Error())
					}
					blog.Title = b.String()
					blog.Title = strings.TrimSpace(blog.Title)
					blog.Valid =  CheckBlog(blog.Title)
				}
			}
		}

	}

	if n.Type == html.ElementNode && n.Data == "div" {
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "rich_media_content " && blog.Valid {
				FindImageAndDownload(n, blog)
				b := new(bytes.Buffer)
				if err := html.Render(b, n); err != nil {
					Logger.Error(err.Error())
				}
				blog.Content = b.String()
			}
		}

	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindBlog(c, blog)
	}
}

func FindBlogURL(n *html.Node, blogs *[]m.Blog) {
	// title
	if n.Type == html.ElementNode && n.Data == "h3" {
		blog := m.Blog{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "a" {
				for _, a := range c.Attr {
					if a.Key == "href" {
						blog.URL = a.Val
					}
				}

			}
		}
		*blogs = append(*blogs, blog)
	} else {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			FindBlogURL(c, blogs)
		}
	}
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


func Start(dbc *sql.DB) {
	Init(dbc)
	CrawlBlog()
}

func HttpGet(url string, headers map[string]string) (result []byte, err error) {

	client := &http.Client{}
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	req.SetBasicAuth (githubusername, githubpassword)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		Logger.Error("数据读取异常")
		return nil, err
	}
	defer resp.Body.Close()
	return body, nil
}

func CrawlBlog() {
	for true {
		var blogkeywords = []string{}
		sql := "select keyword from blogseed where flag = 0"

		rows, err := db.Query(sql)
		defer rows.Close()
		if err != nil {
			Logger.Error(err.Error())
			return
		}
		var key string
		for rows.Next() {
			rows.Scan( &key)
			if key != "" {
				blogkeywords = append(blogkeywords, key)
			}
		}

		if len(blogkeywords) == 0 {
			stmt, _ := db.Prepare("update blogseed set flag=0")
			stmt.Exec()
			stmt.Close()

		}

		for k, keyword := range blogkeywords {
			stmt, _ := db.Prepare("update blogseed set flag=1 where keyword=?")
			stmt.Exec(keyword)
			stmt.Close()

			page := 1

			for page < 2  && k < 4 {
				blogs := []m.Blog{}

				var url = fmt.Sprintf(WEIXIN_SEARCH_URL, keyword, page)
				page = page + 1
				doc, err := GetHTMLNodeFromURL(url)

				if err != nil {
					Logger.Error(err.Error())
				}

				Logger.Info(url)

				FindBlogURL(doc, &blogs)

				for _, blog := range blogs {
					doc, err := GetHTMLNodeFromURL(blog.URL)
					blog.Tag = keyword
					blog.Valid = true

					if err != nil {
						Logger.Error(err.Error())
					}
					FindBlog(doc, &blog)
					//fmt.Printf("%+v", blog)
					if blog.Valid {
						blog.FillAll()
						err = blog.Save(db)
						if err != nil {
							Logger.Error(err.Error())
						}
					}
					fmt.Println("blog.Title = " + blog.Title)
					time.Sleep(200*time.Second)

				}

				time.Sleep(3*time.Minute)
			}
			time.Sleep(6*time.Minute)
		}
	}
}

func main() {
	//Init()
	//CrawlBlog()
}
