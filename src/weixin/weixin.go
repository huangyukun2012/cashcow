package main
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
	"net/http"
	"logging"
//	"encoding/json"
	t "developerq/trans"
	m "developerq/model"
	u "developerq/utils"
//	"os"
)

var Logger *logging.Logger




const WEIXIN_SEARCH_URL = "http://weixin.sogou.com/weixin?usip=null&query=%s&from=tool&ft=&tsn=1&et=&interation=null&type=2&wxid=&page=%d&ie=utf8"




var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var cfg *goconfig.ConfigFile
var githubusername, githubpassword string

//Mysql
func Init() {

	logSvc := logging.NewLogServcie()
	logSvc.ConfigDefaultLogger("/tmp/developerq", "weixin", logging.INFO, logging.ROTATE_DAILY)
	logSvc.Serve()
	//defer logSvc.Stop()
	Logger = logSvc.GetLogger("default")
	m.Logger = Logger

	cfg, ConfError = goconfig.LoadConfigFile("config/developerq.ini")
	if ConfError != nil {
		panic("配置文件config.ini不存在,请将配置文件复制到运行目录下")
	}

	username, ConfError = cfg.GetValue("MySQL", "username")
	if ConfError != nil {
		panic("读取数据库username错误")
	}
	password, ConfError = cfg.GetValue("MySQL", "password")
	if ConfError != nil {
		panic("读取数据库password错误")
	}
	url, ConfError = cfg.GetValue("MySQL", "url")
	if ConfError != nil {
		panic("读取数据库url错误")
	}

	githubusername, ConfError = cfg.GetValue("Github", "username")
	if ConfError != nil {
		panic("error reading github username")
	}
	githubpassword, ConfError = cfg.GetValue("Github", "password")
	if ConfError != nil {
		panic("error reading github password")
	}

	var dataSourceName bytes.Buffer
	dataSourceName.WriteString(username)
	dataSourceName.WriteString(":")
	dataSourceName.WriteString(password)
	dataSourceName.WriteString("@")
	dataSourceName.WriteString(url)
	db, err = sql.Open("mysql", dataSourceName.String())
	if err != nil {
		Logger.Error(err.Error())
	}

	if err := db.Ping(); err != nil {
		panic("数据库连接出错,请检查配置账号密码是否正确")
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(30)
	u.InitRedis()
}


func checkErr(err error) {
	if err != nil {
		Logger.Error(err.Error())
		panic(err.Error())
	}
}




func FindBlog(n *html.Node, blog *m.Blog) {
	// title
	if n.Type == html.ElementNode && n.Data == "article" {
		fmt.Println("find article")
		b := new(bytes.Buffer)
		if err := html.Render(b, n); err != nil {
			Logger.Error(err.Error())
		}
		blog.Content = b.String()
		//translate
		blog.ContentCN = t.TranslateHTMLNodeWithPrefix(n, "gh")

	} else {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			FindBlog(c, blog)
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

var blogkeywords = []string{"linux", "redis",}
/*
func CrawlBlog(db *sql.DB, start int) {

	for {
		blog := m.Blog{}


		doc, err := GetHTMLNodeFromURL(blog.URL)
		//trans
		blog.Title = blog.Description
		blog.TitleCN = t.TranslateTextWithPrefix(blog.Title, "gh")
		if err != nil {
			Logger.Error("Failed with url %s, %s", blog.URL, err.Error())
			//mark 3 when curl failed
			stmt, _ := db.Prepare("update githuburl set flag=3 where url=?")
			stmt.Exec(blog.URL)
			stmt.Close()
			Logger.Info("Finished blog %s", blog.URL)
			//return
			continue
		} else {
			Logger.Info("Downloaded url %s", blog.URL)
		}
		Logger.Info("Parsing blog = %s", blog.URL)
		FindBlog(doc, &blog)


		blog.FillAll()
		//fmt.Printf("%+v\n", blog)
		err = blog.Save(db)

		if err != nil {
			Logger.Error("Failed at blog %s, err = %s", blog.URL, err.Error())
			//mark 4 when insert failed
			stmt, _ := db.Prepare("update githuburl set flag=4 where url=?")
			stmt.Exec(blog.URL)
			stmt.Close()
			continue
		}
		//mark 1 when it's done
		stmt, _ = db.Prepare("update githuburl set flag=1 where url=?")
		stmt.Exec(blog.URL)
		stmt.Close()
		Logger.Info("Finished blog %s", blog.URL)

		//deley 2 minutes
		time.Sleep(time.Minute*2)
	}


}
*/

func Start() {
//	Init()
	//CrawlSEURL(db, arg2, int(arg3), int(arg4))
//	CrawlGHBlog(db, 1)
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
	for _, keyword := range blogkeywords {
		page := 1

		for page < 4 {

			var url = fmt.Sprintf(WEIXIN_SEARCH_URL, keyword, page)
			page = page + 1
			fmt.Println(url)
			result, err := HttpGet(url, nil)
			if err != nil {
				//log.Info(err.Error())
				fmt.Println(err.Error())
			}

			fmt.Println(string(result))

			//parse search result


			//sleep for paging
			time.Sleep(2*time.Second)
		}
	}
}

func main() {
	Init()
	CrawlBlog()

}
