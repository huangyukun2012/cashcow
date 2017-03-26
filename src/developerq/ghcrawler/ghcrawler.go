package ghcrawler
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
	"encoding/json"
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
	logSvc.ConfigDefaultLogger("/tmp/developerq", "ghcrawler", logging.INFO, logging.ROTATE_DAILY)
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




func FindReadMe(n *html.Node, readme *m.ReadMe) {
	// title
	if n.Type == html.ElementNode && n.Data == "article" {
		fmt.Println("find article")
		b := new(bytes.Buffer)
		if err := html.Render(b, n); err != nil {
			Logger.Error(err.Error())
		}
		readme.Content = b.String()
		//translate
		readme.ContentCN = t.TranslateHTMLNodeWithPrefix(n, "gh")

	} else {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			FindReadMe(c, readme)
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


func CrawlGHReadMe(db *sql.DB, start int) {

	for {
		readme := m.ReadMe{}

		rows, err := db.Query("select url, name, description, stars, fork, follow, language from githuburl where flag=0 and id > ? limit 1", start)
		if err != nil {
			Logger.Error(err.Error())
			return
		}

		for rows.Next() {
			err = rows.Scan( &readme.URL, &readme.Name, &readme.Description, &readme.Stars, &readme.Fork,& readme.Follow, &readme.Language)
			checkErr(err)
		}
		//mark processing
		stmt, _ := db.Prepare("update githuburl set flag=2 where url=?")
		stmt.Exec(readme.URL)
		stmt.Close()
		fmt.Println(readme.URL)
		Logger.Info("Crawl readme = %s", readme.URL)

		doc, err := GetHTMLNodeFromURL(readme.URL)
		//trans
		readme.Title = readme.Description
		readme.TitleCN = t.TranslateTextWithPrefix(readme.Title, "gh")
		if err != nil {
			Logger.Error("Failed with url %s, %s", readme.URL, err.Error())
			//mark 3 when curl failed
			stmt, _ := db.Prepare("update githuburl set flag=3 where url=?")
			stmt.Exec(readme.URL)
			stmt.Close()
			Logger.Info("Finished readme %s", readme.URL)
			//return
			continue
		} else {
			Logger.Info("Downloaded url %s", readme.URL)
		}
		Logger.Info("Parsing readme = %s", readme.URL)
		FindReadMe(doc, &readme)


		readme.FillAll()
		//fmt.Printf("%+v\n", readme)
		err = readme.Save(db)

		if err != nil {
			Logger.Error("Failed at readme %s, err = %s", readme.URL, err.Error())
			//mark 4 when insert failed
			stmt, _ := db.Prepare("update githuburl set flag=4 where url=?")
			stmt.Exec(readme.URL)
			stmt.Close()
			continue
		}
		//mark 1 when it's done
		stmt, _ = db.Prepare("update githuburl set flag=1 where url=?")
		stmt.Exec(readme.URL)
		stmt.Close()
		Logger.Info("Finished readme %s", readme.URL)

		//deley 2 minutes
		time.Sleep(time.Minute*2)
	}


}


func Start() {
	Init()
	//CrawlSEURL(db, arg2, int(arg3), int(arg4))
	CrawlGHReadMe(db, 1)
}


var Header = map[string]string{
	"User-Agent":"Awesome-Octocat-App",
	"Accept": "application/vnd.github.v3+json",

}

type SearchResult struct {
	Items []Item  `json:"items"`
}

type Item struct {
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	Description string     `json:"description"`
	Url         string     `json:"html_url"`
	Fork        int        `json:"forks_count"`
	Stars       int        `json:"stargazers_count"`
	Follow      int        `json:"watchers_count"`
	Language    string     `json:"language"`
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

func CrawlGHURL(star int) {
	step := 30
	first := true
	for star > 1 {
		page := 1
		for true {

			var url string
			if first {
				url = fmt.Sprintf(GH_API_BIG, star, page)

			} else {
				url = fmt.Sprintf(GH_API, star - step, star, page)
			}
			page = page + 1
			fmt.Println(url)
			result, _ := HttpGet(url, Header)
			fmt.Println(string(result))

			sr := SearchResult{}

			err := json.Unmarshal(result, &sr)
			if err != nil {
				fmt.Println(err.Error())
				Logger.Error("Failed to read search result, %s", err.Error())
				break
			}
			//fmt.Printf("%+v\n", sr)

			if len(sr.Items) == 0 {
				break
			}

			for _, item := range sr.Items {

				//item.Url = "https://github.com/" + item.FullName
				fmt.Printf("%+v\n", item)
				_, err = db.Exec("insert into githuburl (url, name, description, stars, fork, follow, language) values (?, ?, ?, ?, ?, ?, ?)", item.Url, item.Name, item.Description, item.Stars, item.Fork, item.Follow, item.Language)
				fmt.Println("Insert URL == " + item.Url)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			//sleep for paging
			time.Sleep(1*time.Second)
		}

		if first {
			first = false
		}

		star = star - step - 1

		if star < 10000 {
			step = 20
		}

		if star < 5000 {
			step = 10
		}

		if star < 2000 {
			step = 5
		}

		if star < 1000 {
			step = 2
		}

		if star < 500 {
			step = 1
		}


		fmt.Printf("star = %d, step = %d\n", star, step)
		time.Sleep(15*time.Second)
	}
}

func main() {

	if(len(os.Args) < 2) {
		fmt.Println("please choose 'url' or readme' as parameter")
		return
	}

	arg := os.Args[1]


	Init()

	if arg == "url" {
		//arg2 := os.Args[2]
		CrawlGHURL(20000)
	}

	if arg == "readme" {
		arg2, _ := strconv.ParseInt(os.Args[2], 10, 0)
		CrawlGHReadMe(db, int(arg2))
	}
}
