package rhcrawler
import (
	"golang.org/x/net/html"
	"github.com/Unknwon/goconfig"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"io/ioutil"
	"net/http"
	"logging"
	//"encoding/json"
	t "developerq/trans"
	m "developerq/model"
	u "developerq/utils"
	"os"
)

var Logger *logging.Logger




const GH_API = "https://api.github.com/search/repositories?q=stars:%d..%d&page=%d&per_page=90'"

const GH_API_BIG = "https://api.github.com/search/repositories?q=stars:%d..1000000&page=%d&per_page=90'"



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
	logSvc.ConfigDefaultLogger("/tmp/developerq", "rhcrawler", logging.INFO, logging.ROTATE_DAILY)
	logSvc.Serve()
	//defer logSvc.Stop()
	Logger = logSvc.GetLogger("default")


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




func FindKB(n *html.Node, article *m.Article) {
	// title
	if n.Type == html.ElementNode && n.Data == "section" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "field_kcs_issue_txt" {
				flag = true
			}
		}
		if flag == true {
			b := new(bytes.Buffer)
			if err := html.Render(b, n); err != nil {
				Logger.Error(err.Error())
			}
			article.Question = b.String()
			//article.Question = strings.Replace(article.Question, "<section class=\"field_kcs_issue_txt\">", "<div class=\"post-text\" itemprop=\"text\">", -1)
			article.Question = strings.Replace(article.Question, "<section class=\"field_kcs_issue_txt\">", "", -1)

			//article.Question = strings.Replace(article.Question, "</section>", "</div>", -1)
			article.Question = strings.Replace(article.Question, "</section>", "", -1)

			//translate
			article.QuestionCN = t.TranslateHTMLNodeWithPrefix(n, "rh")
			article.QuestionCN = article.QuestionCN + article.Question
			article.QuestionCN = "<div class=\"post-text\" itemprop=\"text\">" + article.QuestionCN + "</div>"

		}

		article.Answer = `"<div class="post-text" itemprop="text"><h2>请关注DeveloperQ公众号以获得问题解决方法</h>
<img src="/static/img/qr_m.jpg"  alt="DeveloperQ公众号" /></div>"`
		//translate
		//article.ContentCN = t.TranslateHTMLNodeGH(n)
		article.AnswerCN = article.Answer
	}

	if n.Type == html.ElementNode && n.Data == "h1" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "title" {
				flag = true
			}
		}
		if flag == true {
			if c := n.FirstChild; c != nil && article.Title == "" {
				article.Title = string(c.Data)
				article.TitleCN = t.TranslateTextWithPrefix(article.Title, "rh")
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindKB(c, article)
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


func CrawlRHKB(db *sql.DB, start int) {

	for {
		//update every 2 minute
		article := m.Article{}
		article.Source = "rh"

		rows, _ := db.Query("select url, ext_id from rhurl where flag=0 and id > ? limit 1", start)
		for rows.Next() {
			err = rows.Scan(&article.URL, &article.ExtID)
			checkErr(err)
		}
		//mark processing
		stmt, _ := db.Prepare("update rhurl set flag=2 where url=?")
		stmt.Exec(article.URL)
		stmt.Close()

		Logger.Info("Crawl question = %s", article.URL)

		doc, err := GetHTMLNodeFromURL(article.URL)
		if err != nil {
			Logger.Error("Failed with url %s, %s", article.URL, err.Error())
			//mark 3 when curl failed
			stmt, _ := db.Prepare("update rhurl set flag=3 where url=?")
			stmt.Exec(article.URL)
			stmt.Close()
			Logger.Info("Finished article %d", article.ExtID)
			//return
			continue
		} else {
			Logger.Info("Downloaded url %s", article.URL)
		}
		Logger.Info("Parsing question = %s", article.URL)
		FindKB(doc, &article)
		article.Tags = "rhel,linux,redhat"

		article.FillAll()
		//fmt.Printf("%+v\n", article)
		err = article.Save(db)
		if err != nil {
			Logger.Error("Failed at article %d, %s", article.ExtID, err.Error())
			//mark 4 when insert failed
			stmt, _ := db.Prepare("update rhurl set flag=4 where url=?")
			stmt.Exec(url)
			stmt.Close()
			continue
		}
		//mark 1 when it's done
		stmt, _ = db.Prepare("update rhurl set flag=1 where url=?")
		stmt.Exec(url)
		stmt.Close()
		Logger.Info("Finished article %d", article.ExtID)

		//deley 2 minutes
		time.Sleep(time.Minute*2)
		//time.Sleep(time.Second*2)
	}
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

func Start() {
	Init()
	CrawlRHKB(db, 0)
}

func main() {

	if(len(os.Args) < 2) {
		fmt.Println("please choose 'url' or article' as parameter")
		return
	}

	arg := os.Args[1]


	Init()

	if arg == "kb" {
		arg2, _ := strconv.ParseInt(os.Args[2], 10, 0)
		CrawlRHKB(db, int(arg2))
	}
}
