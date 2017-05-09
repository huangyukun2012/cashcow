package bilisou

import (

	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"net/http"
	"html/template"
	_ "github.com/go-sql-driver/mysql"
	sql "database/sql"
	"strconv"
//	"regexp"
	"encoding/json"
//	"time"
//	"github.com/garyburd/redigo/redis"
	"github.com/Unknwon/goconfig"
//	"strconv"
	"bytes"
//	"os"
//	"bufio"
//	"io"
	"strings"
	u "utils"
	"logging"
	m "bilisou/model"
	c "bilisou/crawler"
	//w "bilisou/weixin"
	b "bilisou/bt"
	es "gopkg.in/olivere/elastic.v3"
	"io/ioutil"
)

type FileData struct {
	ID int
	UID int
	Title string
}

var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var esclient *es.Client
var cfg *goconfig.ConfigFile

//var templateContent *template.Template
//var blogTemplate *template.Template

var showTemplate *template.Template
var listTemplate *template.Template
var searchTemplate *template.Template
var homeTemplate *template.Template
var lostTemplate *template.Template

var bshowTemplate *template.Template
var blistTemplate *template.Template

var regPage string
var loginPage string


var Logger *logging.Logger

func Init() {
	//init log
	logSvc := logging.NewLogServcie()
	logSvc.ConfigDefaultLogger("/tmp/bilisou", "bilisou", logging.INFO, logging.ROTATE_DAILY)
	logSvc.Serve()
	//	defer logSvc.Stop()
	Logger = logSvc.GetLogger("default")
	m.Logger = Logger


	cfg, ConfError = goconfig.LoadConfigFile("config/bilisou.ini")
	if ConfError != nil {
		Logger.Error("配置文件config.ini不存在,请将配置文件复制到运行目录下")
	}

	username, ConfError = cfg.GetValue("MySQL", "username")
	if ConfError != nil {
		Logger.Error("读取数据库username错误")
	}
	password, ConfError = cfg.GetValue("MySQL", "password")
	if ConfError != nil {
		Logger.Error("读取数据库password错误")
	}
	url, ConfError = cfg.GetValue("MySQL", "url")
	if ConfError != nil {
		Logger.Error("读取数据库url错误")
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
		panic("Error Connection database...")
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(30)


	//init es
	esclient, err = es.NewClient()
	if err != nil {
		Logger.Error("failed to create es client")
	}
//	m.TotalShares = m.GetTotalShares(esclient)
//	m.TotalUsers = m.GetTotalUsers(esclient)

	m.MAX_USER, m.MIN_USER = m.GetUserMaxMINID(db)
	m.MAX_SHARE, m.MIN_SHARE = m.GetShareMaxMinID(db)
	m.MAX_KEYWORD, m.MIN_KEYWORD = m.GetKeywordMaxMinID(db)

	if ConfError != nil {
		Logger.Error("读取数据库server错误")
	}

	if ConfError != nil {
		Logger.Error("读取数据库port错误")
	}

	InitTemplates()

	go c.Start(db)
	//go w.Start(db)
	go b.Start(db)


}


func SetURL(url string, pv *m.PageVar) error {
	b, err := json.Marshal(pv)
	if err != nil {
		return err
	}
	str := string(b)
	err = u.SetRedis("bilisou" + url, str)
	Logger.Info("Set Cache for %s", url)
	return err
}

func GetURL(url string) (*m.PageVar, error){

	str, err := u.GetRedis("bilisou" + url)
	if err != nil {
		return nil, err
	}

	pv := m.PageVar{}
	err = json.Unmarshal([]byte(str), &pv)
	Logger.Info("Get Cache for %s ", url)
	return &pv, err
}


func InitTemplates() {
	home, err := ioutil.ReadFile("resource/bilisou/templates/home.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	list, err := ioutil.ReadFile("resource/bilisou/templates/list.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	blist, err := ioutil.ReadFile("resource/bilisou/templates/blist.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	bshow, err := ioutil.ReadFile("resource/bilisou/templates/bshow.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	search, err := ioutil.ReadFile("resource/bilisou/templates/search.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	show, err := ioutil.ReadFile("resource/bilisou/templates/show.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	lost, err := ioutil.ReadFile("resource/bilisou/templates/404.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	reg, err := ioutil.ReadFile("resource/bilisou/templates/reg.html")
	if err != nil {
		Logger.Error(err.Error())
	}
	regPage = string(reg)

	login, err := ioutil.ReadFile("resource/bilisou/templates/login.html")
	if err != nil {
		Logger.Error(err.Error())
	}
	loginPage = string(login)


	listTemplate = template.Must(template.New("tmp").Parse(string(list)))
	searchTemplate = template.Must(template.New("tmp").Parse(string(search)))
	homeTemplate = template.Must(template.New("tmp").Parse(string(home)))
	showTemplate = template.Must(template.New("tmp").Parse( string(show)))
	lostTemplate = template.Must(template.New("tmp").Parse( string(lost)))
	bshowTemplate = template.Must(template.New("tmp").Parse( string(bshow)))
	blistTemplate = template.Must(template.New("tmp").Parse( string(blist)))

}


func SetURLBlog(url string, pv *m.PageVar) error {
	b, err := json.Marshal(pv)
	if err != nil {
		return err
	}
	str := string(b)
	err = u.SetRedis("blog" + url, str)
	Logger.Info("Set Cache for blog %s", url)
	return err
}

func GetURLBlog(url string) (*m.PageVar, error){
	str, err := u.GetRedis("blog" + url)
	if err != nil {
		return nil, err
	}

	pv := m.PageVar{}
	err = json.Unmarshal([]byte(str), &pv)
	Logger.Info("Get Cache for blog %s ", url)
	return &pv, err
}


func Index(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s, url = %s", r.RemoteAddr, r.URL)

	/*
	pv, err := GetURL("home")
	if err == nil && pv != nil {
		render(w, homeTemplate, pv)
	} else {
		pv := m.GenerateListPageVar(esclient, 0, 1)
	*/
	pv := m.PageVar{}
	pv.Username = getUserName(r)
	pv.Keywords = m.GetRandomKeywords(db, 5)
	render(w, homeTemplate, pv)
}


func IndexBlog(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s, url = %s", r.RemoteAddr, r.URL)
	pv, err := GetURLBlog("blog home")

	if err == nil && pv != nil {
		pv.Username = getUserName(r)
		render(w, blistTemplate, pv)
	} else {
		pv := m.ListBlogPage(db, esclient, 1, 0)
		SetURL("blog home", pv)
		pv.Username = getUserName(r)
		render(w, blistTemplate, pv)
	}

}


func ListShare(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s, url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		pv.Username = getUserName(r)
		render(w, listTemplate, pv)
		return
	}

	vars := mux.Vars(r)
	p := vars["page"]
	if p == "" {
		p = "1"
	}

	pp, err:=strconv.Atoi(p)
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	pv = m.ListSharePage(db, pp, 0)
	if pv != nil {
		pv.Username = getUserName(r)
		render(w, listTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}

func SearchShare(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s, url = %s", r.RemoteAddr, r.URL)
	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("it's from cache %s", url)
		pv.Username = getUserName(r)
		pv.Username = getUserName(r)
		render(w, searchTemplate, pv)
		return
	}

	vars := mux.Vars(r)
	/*
	cat := vars["category"]
	cati, ok:= u.CAT_STR_INT[cat]
	if !ok {
		Logger.Error(err.Error())
		cati = -1
	}
	*/

	keyword := vars["keyword"]
	if keyword == "" {
		Logger.Error(err.Error())
	}

	p := vars["page"]
	if p == "" {
		p = "1"
	}

	pp, err:=strconv.Atoi(p)
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	m.KeywordHit(db,keyword)
	pv = m.GenerateSearchPageVar(esclient, db, 0, keyword, pp)
	if pv != nil {
		pv.Username = getUserName(r)
		render(w, searchTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}

func ShowShare(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s, url = %s", r.RemoteAddr, r.URL)
	// break down the variables for easier assignment
	vars := mux.Vars(r)
	id := vars["dataid"]
	pv := m.ShowSharePage(db, esclient, id)
	if pv != nil {

		pv.Username = getUserName(r)
		render(w, showTemplate, pv)
	}
}

func ListBlog(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURLBlog(r.URL.Path)

	if err == nil && pv != nil {
		Logger.Info("load blog %s from cache", url)
		pv.Username = getUserName(r)
		pv.Username = getUserName(r)
		render(w, blistTemplate, pv)
		return
	}

	vars := mux.Vars(r)
	cat := vars["category"]
	//cati, ok:= u.CAT_STR_INT[cat]
	cati, err := strconv.Atoi(cat)
	if err != nil {
		Logger.Error(err.Error())
		cati = -1
	}

	p := vars["page"]
	if p == "" {
		p = "1"
	}

	pp, err:=strconv.Atoi(p)
	if err != nil {
		Logger.Error(err.Error())
		pp = 1
	}
	pv = m.ListBlogPage(db, esclient, pp, cati)

	if pv != nil {
		pv.Username = getUserName(r)
		render(w, blistTemplate, pv)
	}

	SetURLBlog(r.URL.Path, pv)
}


func ShowBlog(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURLBlog(r.URL.Path)

	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url )
		pv.Username = getUserName(r)
		render(w, bshowTemplate, pv)
		return
	}

	// break down the variables for easier assignment
	vars := mux.Vars(r)
	ukk := vars["uk"]
	uk, err:=strconv.Atoi(ukk)
	if err != nil {
		Logger.Info(err.Error() )
		uk = 0
	}

	pv = m.ShowBlogPage(db, esclient, int64(uk))
	//update viewcount
	//m.ViewArticle(db, int64(uk))
	if pv != nil {
		pv.Username = getUserName(r)
		render(w, bshowTemplate, pv)
	}
	SetURLBlog(r.URL.Path, pv)
}


func NotFound(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s, url = %s", r.RemoteAddr, r.URL)
	pv := &m.PageVar{}
	pv.Type = "lost"
	pv.Username = getUserName(r)
	w.WriteHeader(http.StatusNotFound)
	if pv != nil {
		render(w, lostTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}

/*

func NotFoundBlog(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s, url = %s", r.RemoteAddr, r.URL)
	pv := m.ListBlogPage(db, esclient, 1, 0)
	pv.Type = "lost"
	w.WriteHeader(http.StatusNotFound)
	if pv != nil {
		render(w, blogTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}
*/




func render(w http.ResponseWriter, t *template.Template, data interface{}) {
	/*if err != nil {
		Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}*/
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}



func Robots(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User-agent: *\nDisallow:\n")
}


var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}


func Login(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		fmt.Fprintf(response, loginPage)
		return
	}

	if request.Method == "POST" {
		name := request.FormValue("username")
		pass := request.FormValue("password")
		if name != "" && pass != "" {
			id := -1
			r_name := ""
			r_pass := ""
			rows, err := db.Query("select id, username, password from user where username = ?", name)
			defer rows.Close()
			if err != nil {
				fmt.Println(err.Error())
			}

			for rows.Next() {
				rows.Scan(&id, &r_name, &r_pass)
			}

			if id != -1 && name == r_name && pass == r_pass {

				ss := strings.Split(request.RemoteAddr, ":")
				if len(ss) != 2 {
					return
				}
				ipaddr := ss[0]

				stmt, _ := db.Prepare("update user set last_login = ?  where id = ?")
				stmt.Exec(ipaddr, id)
				stmt.Close()

				fmt.Fprintf(response, "true")
				setSession(name, response)
				return
			}
		}
		fmt.Fprintf(response, "false")
		return
	}

}

type RegMsg struct {
	Status int
	Reason string
}

func Register(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		fmt.Fprintf(response, regPage)
		return
	}

	if request.Method == "POST" {
		msg := RegMsg{}
		email := request.FormValue("email")
		if !strings.Contains(email, ".") && !strings.Contains(email, "@") {
			msg.Status = 0
			msg.Reason = "非法邮箱地址"
			b, _ := json.Marshal(msg)
			fmt.Fprintf(response, string(b))
			return
		}
		name := request.FormValue("username")
		pass := request.FormValue("password")
		if name != "" && pass != "" {
			id := -1
			r_name := ""
			r_pass := ""
			rows, err := db.Query("select id, username, password from user where username = ?", name)
			defer rows.Close()
			if err != nil {
				msg.Status = 0
				msg.Reason = "服务器错误"
				b, _ := json.Marshal(msg)
				fmt.Fprintf(response, string(b))
				return
			}

			for rows.Next() {
				rows.Scan(&id, &r_name, &r_pass)
			}

			if id != -1 {
				msg.Status = 0
				msg.Reason = "用户名已经存在，请换一个用户名注册"
				b, _ := json.Marshal(msg)
				fmt.Fprintf(response, string(b))
				return
			}

			stmt, _ := db.Prepare("insert into user(username, password, email) values(?, ?, ?)")
			stmt.Exec(name, pass, email)
			stmt.Close()
			msg.Status = 1
			b, _ := json.Marshal(msg)
			fmt.Fprintf(response, string(b))
			return
		}
		msg.Status = 0
		msg.Reason = "用户名或密码为空"
		b, _ := json.Marshal(msg)
		fmt.Fprintf(response, string(b))
		return

	}
}


func Logout(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}


/*
func StartBlog(mx *mux.Router) {

//	Init()

	//u.SetURL("aaa", "aabb")
	//Logger.Info(u.GetURL("aa"))
	mx.HandleFunc("", IndexBlog)
	mx.HandleFunc("/", IndexBlog)
	//list
	mx.HandleFunc("/list/{category}", ListBlog)
	mx.HandleFunc("/list/{category}/", ListBlog)
	mx.HandleFunc("/list/{category}/{page}", ListBlog)
	mx.HandleFunc("/list/{category}/{page}/", ListBlog)

	//blog
	mx.HandleFunc("/blog/{uk}", ShowBlog)
	mx.HandleFunc("/blog/{uk}/", ShowBlog)


	//server static
	mx.PathPrefix("/imgpool1").Handler(http.FileServer(http.Dir("")))

	//server static
	mx.PathPrefix("/static").Handler(http.FileServer(http.Dir("resource/bilisou/")))

	//for baidu
	mx.HandleFunc("/robots.txt", Robots)

	//not found
	mx.NotFoundHandler = http.HandlerFunc(NotFound)
}
*/

func Start(mx *mux.Router) {

//	Init()

	//u.SetURL("aaa", "aabb")
	//Logger.Info(u.GetURL("aa"))

	mx.HandleFunc("/", Index)
	//list
	mx.HandleFunc("/list", ListShare)
	mx.HandleFunc("/list/", ListShare)
	mx.HandleFunc("/list/{page}", ListShare)
	mx.HandleFunc("/list/{page}/", ListShare)

	//ulist
/*
	mx.HandleFunc("/ulist", ListUsers)
	mx.HandleFunc("/ulist/", ListUsers)
	mx.HandleFunc("/ulist/{page}", ListUsers)
	mx.HandleFunc("/ulist/{page}/", ListUsers)
*/
	//search
	mx.HandleFunc("/search/{keyword}", SearchShare)
	mx.HandleFunc("/search/{keyword}/", SearchShare)
	mx.HandleFunc("/search/{keyword}/{page}", SearchShare)
	mx.HandleFunc("/search/{keyword}/{page}/", SearchShare)

	//file
	mx.HandleFunc("/file/{dataid}", ShowShare)
	mx.HandleFunc("/file/{dataid}/", ShowShare)


	mx.HandleFunc("/blist", IndexBlog)
	mx.HandleFunc("/blist/", IndexBlog)
	mx.HandleFunc("/blist/{category}", ListBlog)
	mx.HandleFunc("/blist/{category}/", ListBlog)
	mx.HandleFunc("/blist/{category}/{page}", ListBlog)
	mx.HandleFunc("/blist/{category}/{page}/", ListBlog)

	//blog
	mx.HandleFunc("/blog/{uk}", ShowBlog)
	mx.HandleFunc("/blog/{uk}/", ShowBlog)



	//mx.HandleFunc("/", indexPageHandler)
	//mx.HandleFunc("/internal", internalPageHandler)

	mx.HandleFunc("/login", Login)
	mx.HandleFunc("/register", Register)
	mx.HandleFunc("/logout", Logout)

	//server static
	mx.PathPrefix("/static").Handler(http.FileServer(http.Dir("resource/bilisou/")))

	//server static
	mx.PathPrefix("/imgpool1").Handler(http.FileServer(http.Dir("")))

	//for baidu
	mx.HandleFunc("/robots.txt", Robots)

	//not found
	mx.NotFoundHandler = http.HandlerFunc(NotFound)

}
