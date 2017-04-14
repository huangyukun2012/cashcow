package socrawler
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
	t "developerq/trans"
	m "developerq/model"
	u "utils"
)

var Logger *logging.Logger

const SO_LIST_URL = "http://stackoverflow.com/questions/tagged/%s?page=%d&sort=votes&pagesize=50"
const SO_BASE_URL = "http://stackoverflow.com"

const SE_LIST_URL = "http://%s/questions?page=%d&sort=votes"
const SE_BASE_URL = "http://%s"

const AU_LIST_URL = "http://askubuntu.com/questions?page=%d&sort=votes"
const AU_BASE_URL = "http://askubuntu.com"

var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var cfg *goconfig.ConfigFile

//Mysql
func Init(dbc *sql.DB) {

	logSvc := logging.NewLogServcie()
	logSvc.ConfigDefaultLogger("/tmp/developerq", "socrawler", logging.INFO, logging.ROTATE_DAILY)
	logSvc.Serve()
	//defer logSvc.Stop()
	Logger = logSvc.GetLogger("default")

	db = dbc
}


func checkErr(err error) {
	if err != nil {
		Logger.Error(err.Error())
		panic(err.Error())
	}
}


func SetQuestionDetail(n *html.Node, article *m.Article) {
	//for post text
	if n.Type == html.ElementNode && n.Data == "div" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "post-text" {
				flag = true
			}
		}
		if flag == true {
			b := new(bytes.Buffer)
			if err := html.Render(b, n); err != nil {
				Logger.Error(err.Error())
			}
			article.Question = b.String()
			//translate
			article.QuestionCN = t.TranslateHTMLNodeWithPrefix(n, "so")
		}
	}

	//for tag list
	if n.Type == html.ElementNode && n.Data == "div" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "post-taglist" {
				flag = true
			}
		}
		if flag == true {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				for _, a := range c.Attr {
					if a.Key == "href" {
						tags := strings.Split(a.Val, "/")
						//Logger.Info(tags)
						tag := tags[len(tags) - 1]
						//Logger.Info(tag)
						article.Tags = article.Tags + tag + ", "
					}
				}
			}
			//Logger.Info("Tags = ", article.Tags)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		SetQuestionDetail(c, article)
	}

}

func SetViewCount(n *html.Node, article *m.Article) {
	//for answer post text

	if strings.Contains(n.Data, "times") {
		times, _ := strconv.ParseInt(strings.Split(n.Data, " ")[0], 10, 0)
		article.ViewCount = times
		//Logger.Info("ViewCount = ", article.ViewCount)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		SetViewCount(c, article)
	}


}

func SetAnserDetail(n *html.Node, article *m.Article) {
	//for answer post text
	if n.Type == html.ElementNode && n.Data == "div" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "post-text" {
				flag = true
			}
		}
		if flag == true {
			b := new(bytes.Buffer)
			if err := html.Render(b, n); err != nil {
				Logger.Error(err.Error())
			}
			article.Answer = b.String()
			//translate
			article.AnswerCN = t.TranslateHTMLNodeWithPrefix(n, "so")
			//Logger.Info("answerCN = ", article.AnswerCN)
		}
	}

	//for vote count
	if n.Type == html.ElementNode && n.Data == "span" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "itemprop" && a.Val == "upvoteCount" {
				flag = true
			}
		}
		if flag == true {
			if n.FirstChild != nil && n.FirstChild.Data != "" {
				article.VoteCount, _ = strconv.ParseInt(n.FirstChild.Data, 10, 0)
			}
			//Logger.Info("vote count = ", article.VoteCount)

		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		SetAnserDetail(c, article)
	}

}


func SetUnresolvedAnswerDetail(n *html.Node, article *m.Article) {
	//for answer post text
	if n.Type == html.ElementNode && n.Data == "div" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "post-text" {
				flag = true
			}
		}
		if flag == true {
			b := new(bytes.Buffer)
			if err := html.Render(b, n); err != nil {
				Logger.Error(err.Error())
			}
			article.Answer = b.String()
			//translate
			article.AnswerCN = t.TranslateHTMLNodeWithPrefix(n, "so")
			//Logger.Info("answerCN = ", article.AnswerCN)
		}
	}

	//for vote count
	if n.Type == html.ElementNode && n.Data == "span" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "itemprop" && a.Val == "upvoteCount" {
				flag = true
			}
		}
		if flag == true {
			if n.FirstChild != nil && n.FirstChild.Data != "" {
				article.VoteCount, _ = strconv.ParseInt(n.FirstChild.Data, 10, 0)
			}
			//Logger.Info("vote count = ", article.VoteCount)

		}
	}

	for c := n.FirstChild; article.AnswerCN == "" && c != nil; c = c.NextSibling {
		SetUnresolvedAnswerDetail(c, article)
	}

}


func FindUnresolvedQuestion(n *html.Node, article *m.Article) {

	//unresolved content
	if n.Type == html.ElementNode && n.Data == "div" {
		flag := ""
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "answer" {
				flag = "answer"
			}

		}
		if flag == "answer" {
			SetUnresolvedAnswerDetail(n, article)
			return
		}

	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindUnresolvedQuestion(c, article)
	}
}


func FindQuestion(n *html.Node, article *m.Article) {
	// title
	if n.Type == html.ElementNode && n.Data == "a" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "question-hyperlink" {
				flag = true
			}
		}
		if flag == true {
			if c := n.FirstChild; c != nil && article.Title == "" {
				article.Title = string(c.Data)
				//fmt.Println("Title = ", question.Title)
				article.TitleCN = t.TranslateTextWithPrefix(article.Title, "so")
				//Logger.Info("TitleCN = ", article.TitleCN)
			}
		}

	}

	// ext_id
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, a := range n.Attr {
			if a.Key == "data-questionid" {
				article.ExtID, _ = strconv.ParseInt(a.Val, 10, 0)
				//Logger.Info("extid = ", article.ExtID)
			}
		}
	}


	//question content
	if n.Type == html.ElementNode && n.Data == "div" {
		flag := ""
		for _, a := range n.Attr {

			if a.Key == "id" && a.Val == "question" {
				flag = "question"
			}
			if a.Key == "class" && a.Val == "answer accepted-answer" {
				flag = "answer"
			}

		}
		if flag == "question" {
			SetQuestionDetail(n, article)
		}
		if flag == "answer" {
			SetAnserDetail(n, article)
		}

	}

	//for viewcount
	if n.Type == html.ElementNode && n.Data == "table" {
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == "qinfo" {
				SetViewCount(n, article)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindQuestion(c, article)
	}
}

func FindURLs(n *html.Node, urls *[]string){
	if n.Type == html.ElementNode && n.Data == "a" {
		flag := false
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "question-hyperlink" {
				flag = true
			}
		}
		if flag == true {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if !strings.HasPrefix(a.Val, "http") {
						*urls = append(*urls, string(a.Val))
						Logger.Info("found url = %s", a.Val)
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		FindURLs(c, urls)
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


func CrawlSOQuestion(start int) {

	for {
		//update every 2 minute
		rows, _ := db.Query("select url from sourl where flag=0 and id > ? limit 1", start)
		var url string
		for rows.Next() {
			err = rows.Scan( &url)
			checkErr(err)
		}
		rows.Close()
		//mark processing
		stmt, _ := db.Prepare("update sourl set flag=2 where url=?")
		stmt.Exec(url)
		stmt.Close()

		article := m.Article{}
		article.Source = "so"
		fullurl := url
		Logger.Info("Crawl question = %s", fullurl)

		doc, err := GetHTMLNodeFromURL(fullurl)
		if err != nil {
			Logger.Error("Failed with url %s, %s", fullurl, err.Error())
			//mark 3 when curl failed
			stmt, _ := db.Prepare("update sourl set flag=3 where url=?")
			stmt.Exec(url)
			stmt.Close()
			Logger.Info("Finished article %d", article.ExtID)
			//return
			continue
		} else {
			Logger.Info("Downloaded url %s", url)
		}
		Logger.Info("Parsing question = %s", fullurl)
		FindQuestion(doc, &article)
		if article.Answer == "" {
			Logger.Info("Failed to find accept ac = %s", url)
			FindUnresolvedQuestion(doc, &article)
		}
		article.URL = fullurl
		article.FillAll()
		err = article.Save(db)
		if err != nil {
			Logger.Error("Failed at article %d", article.ExtID)
			//mark 4 when insert failed
			stmt, _ := db.Prepare("update sourl set flag=4 where url=?")
			stmt.Exec(url)
			stmt.Close()
			continue
		}
		//mark 1 when it's done
		stmt, _ = db.Prepare("update sourl set flag=1 where url=?")
		stmt.Exec(url)
		stmt.Close()
		Logger.Info("Finished article %d", article.ExtID)

		//deley 2 minutes
		time.Sleep(time.Minute*2)
	}


}


func CrawlSEURL(db *sql.DB, hostname string, start int,  end int) {

	for i := start; i < end; i ++ {
		url := fmt.Sprintf(SE_LIST_URL,hostname, i)
		doc, err := GetHTMLNodeFromURL(url)
		checkErr(err)

		urls := []string{}
		FindURLs(doc, &urls)
		if len(urls) == 0 {
			Logger.Info("Failed to get URL sleep ping...")
			time.Sleep(3 * time.Minute)
			i --;
			continue
		}
		for _, url := range urls {
			if _, err = u.Get(url); err != nil {


				url = SE_BASE_URL + url

				rows, _ := db.Query("select url  from sourl where url = ?", url)
				scanurl := ""
				for rows.Next() {
					err = rows.Scan(&scanurl)
				}
				rows.Close()
				url = fmt.Sprintf(url, hostname)
				if scanurl == ""  {
					_, err = db.Exec("INSERT into sourl(url,flag, type) values(?,?,?)", url, 0, hostname)
					checkErr(err)
					u.Set(url, "i")
				} else {
					Logger.Info("Ignore url = %s", url)
				}
			}
		}
		fmt.Println(urls)
	}
}

func Start(dbc *sql.DB) {
	Init(dbc)

	CrawlSOQuestion(1)
}

/*
func main() {

	if(len(os.Args) < 2) {
		fmt.Println("please choose 'url' or question' as parameter")
		return
	}

	arg := os.Args[1]


	Init()

	if arg == "url" {
		arg2 := os.Args[2]
		arg3, _ := strconv.ParseInt(os.Args[3], 10, 0)
		arg4, _ := strconv.ParseInt(os.Args[4], 10, 0)
		CrawlSEURL(db, arg2, int(arg3), int(arg4))

	}

	if arg == "question" {
		arg2, _ := strconv.ParseInt(os.Args[2], 10, 0)
		CrawlSOQuestion(db, int(arg2))
	}
}
*/
