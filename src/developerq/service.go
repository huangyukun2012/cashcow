package developerq

import (
	"fmt"
	"github.com/gorilla/mux"
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
//	"strings"
	m "developerq/model"
	u "utils"
	s "developerq/socrawler"
	//r "developerq/rhcrawler"
	//g "developerq/ghcrawler"
	w "weixin"

	es "gopkg.in/olivere/elastic.v3"
	"io/ioutil"
	"logging"
)


var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var esclient *es.Client
var cfg *goconfig.ConfigFile
var Logger *logging.Logger

//all the templates
var listArticleTemplate *template.Template
var searchArticleTemplate *template.Template
var tagArticleTemplate *template.Template
var listTagTemplate *template.Template
var listReadMeTemplate *template.Template
var showReadMeTemplate *template.Template
var showArticleTemplate *template.Template
var listBlogTemplate *template.Template
var showBlogTemplate *template.Template

var notFoundTemplate *template.Template


//Mysql Redis ES init
func Init() {

	logSvc := logging.NewLogServcie()
	logSvc.ConfigDefaultLogger("/tmp/developerq", "developerq", logging.INFO, logging.ROTATE_DAILY)
	logSvc.Serve()
//	defer logSvc.Stop()
	Logger = logSvc.GetLogger("default")
	//u.Logger = Logger
	m.Logger = Logger

	cfg, ConfError = goconfig.LoadConfigFile("config/developerq.ini")
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

	if ConfError != nil {
		Logger.Error("读取数据库server错误")
	}

	if ConfError != nil {
		Logger.Error("读取数据库port错误")
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
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(20)


	//init es
	esclient, err = es.NewClient()
	if err != nil {
		Logger.Error("failed to create es client")
	}
	m.MaxArticle, m.MinArticle = m.GetArticleMaxMinID(db)

	InitTemplates()
	go s.Start(db)
	//go r.Start(db)
	//go g.Start(db)
	go w.Start(db)
}

func InitTemplates() {

	header, err := ioutil.ReadFile("resource/developerq/templates/header.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	list, err := ioutil.ReadFile("resource/developerq/templates/list.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	search, err := ioutil.ReadFile("resource/developerq/templates/search.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	show, err := ioutil.ReadFile("resource/developerq/templates/show.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	tag, err := ioutil.ReadFile("resource/developerq/templates/tag.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	listtag, err := ioutil.ReadFile("resource/developerq/templates/listtag.html")
	if err != nil {
		Logger.Error(err.Error())
	}


	foot, err := ioutil.ReadFile("resource/developerq/templates/foot.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	sidebar, err := ioutil.ReadFile("resource/developerq/templates/sidebar.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	listreadme, err := ioutil.ReadFile("resource/developerq/templates/listreadme.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	showreadme, err := ioutil.ReadFile("resource/developerq/templates/readme.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	listblog, err := ioutil.ReadFile("resource/developerq/templates/listblog.html")
	if err != nil {
		Logger.Error(err.Error())
	}

	showblog, err := ioutil.ReadFile("resource/developerq/templates/blog.html")
	if err != nil {
		Logger.Error(err.Error())
	}


	notfound, err := ioutil.ReadFile("resource/developerq/templates/notfound.html")
	if err != nil {
		Logger.Error(err.Error())
	}


	listArticleTemplate = template.Must(template.New("tmp").Parse(string(header) + string(list) + string(sidebar) + string(foot)))
	searchArticleTemplate = template.Must(template.New("tmp").Parse(string(header) + string(search) + string(sidebar) + string(foot)))
	showArticleTemplate = template.Must(template.New("tmp").Parse(string(header) + string(show) + string(sidebar) + string(foot)))
	//fmt.Println(string(header) + string(show) + string(foot))
	tagArticleTemplate = template.Must(template.New("tmp").Parse(string(header) + string(tag) + string(sidebar) + string(foot)))
	listTagTemplate = template.Must(template.New("tmp").Parse(string(header) + string(listtag) + string(sidebar) + string(foot)))
	listReadMeTemplate = template.Must(template.New("tmp").Parse(string(header) + string(listreadme) + string(sidebar) + string(foot)))
	showReadMeTemplate = template.Must(template.New("tmp").Parse(string(header) + string(showreadme) + string(sidebar) + string(foot)))

	listBlogTemplate = template.Must(template.New("tmp").Parse(string(header) + string(listblog) + string(sidebar) + string(foot)))
	showBlogTemplate = template.Must(template.New("tmp").Parse(string(header) + string(showblog) + string(sidebar) + string(foot)))

	notFoundTemplate = template.Must(template.New("tmp").Parse(string(header) + string(notfound) + string(sidebar) + string(foot)))
}


func SetURL(url string, pv *m.PageVar) error {
	b, err := json.Marshal(pv)
	if err != nil {
		return err
	}
	str := string(b)
	err = u.SetRedis("linuxman" + url, str)
	Logger.Info("Set Cache for %s", url)
	return err
}

func GetURL(url string) (*m.PageVar, error){
	str, err := u.GetRedis("linuxman" + url)
	if err != nil {
		return nil, err
	}

	pv := m.PageVar{}
	err = json.Unmarshal([]byte(str), &pv)
	Logger.Info("Get Cache for %s", url)
	return &pv, err
}


func Index(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)
	pv, err := GetURL("home")
	if err == nil && pv != nil {
		render(w, listArticleTemplate, pv)
	} else {
		pv := m.ListArticlePage(db, esclient, 1)
		SetURL("home", pv)
		render(w, listArticleTemplate, pv)
	}
}


func Dummy(w http.ResponseWriter, r *http.Request) {
	return
}




func ListArticle(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url)
		render(w, listArticleTemplate, pv)
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
		pp = 1
	}
	pv = m.ListArticlePage(db, esclient, pp)

	if pv != nil {
		render(w, listArticleTemplate, pv)
	}

	SetURL(r.URL.Path, pv)
}


func ListReadMe(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url)
		render(w, listReadMeTemplate, pv)
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
		pp = 1
	}
	pv = m.ListReadMePage(db, esclient, pp)

	if pv != nil {
		render(w, listReadMeTemplate, pv)
	}

	SetURL(r.URL.Path, pv)
}


func ListBlog(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url)
		render(w, listBlogTemplate, pv)
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
		pp = 1
	}
	pv = m.ListBlogPage(db, esclient, pp)

	if pv != nil {
		render(w, listBlogTemplate, pv)
	}

	SetURL(r.URL.Path, pv)
}



func NotFound(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)
	pv := m.ListArticlePage(db, esclient, 1)
	pv.Type = "lost"
	w.WriteHeader(http.StatusNotFound)
	if pv != nil {
		render(w, notFoundTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}



func TagArticle(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)
	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url)
		render(w, tagArticleTemplate, pv)
		return
	}

	vars := mux.Vars(r)
	tag := vars["tag"]
	p := vars["page"]
	if p == "" {
		p = "1"
	}

	pp, err:=strconv.Atoi(p)
	if err != nil {
		Logger.Info(err.Error())
		pp = 1
	}
	pv = m.ListTagArticlePage(db, esclient, tag,  pp)
	if pv != nil {
		render(w, tagArticleTemplate, pv)
	}
	SetURL(r.URL.Path, pv)

}

func SearchArticle(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	keyword := r.URL.Query().Get("q")
	page := r.URL.Query().Get("page")
//	sort := r.URL.Query().Get("sort")

	pv, err := GetURL(r.URL.Path + "###" + keyword + "###" + page)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url)
		render(w, searchArticleTemplate, pv)
		return
	}

	//todo
	// if keyword is a tag, jump to tag page

	pp, err:=strconv.Atoi(page)
	if err != nil {
		pp = 1
	}

	m.KeywordHit(db,keyword)
	pv = m.SearchArticlePage(db, esclient, keyword, pp)
	if pv != nil {
		render(w, searchArticleTemplate, pv)
	}
	SetURL(r.URL.Path + "###" + keyword + "###" + page, pv)

}

func ShowArticle(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url )
		render(w, showArticleTemplate, pv)
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

	pv = m.ShowArticlePage(db, esclient, int64(uk))
	//update viewcount
	m.ViewArticle(db, int64(uk))
	if pv != nil {
		render(w, showArticleTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}


func ShowReadMe(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url )
		render(w, showReadMeTemplate, pv)
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

	pv = m.ShowReadMePage(db, esclient, int64(uk))
	//update viewcount
	//m.ViewArticle(db, int64(uk))
	if pv != nil {
		render(w, showReadMeTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}

func ShowBlog(w http.ResponseWriter, r *http.Request) {
	Logger.Info("ip = %s url = %s", r.RemoteAddr, r.URL)

	pv, err := GetURL(r.URL.Path)
	if err == nil && pv != nil {
		Logger.Info("load %s from cache", url )
		render(w, showBlogTemplate, pv)
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
		render(w, showBlogTemplate, pv)
	}
	SetURL(r.URL.Path, pv)
}



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

func Start(mx *mux.Router) {
//	Init()


	mx.HandleFunc("/", Index)

	//list
	mx.HandleFunc("/list", ListArticle)
	mx.HandleFunc("/list/", ListArticle)
	mx.HandleFunc("/list/{page}", ListArticle)
	mx.HandleFunc("/list/{page}/", ListArticle)

	//list readme
	mx.HandleFunc("/listreadme", ListReadMe)
	mx.HandleFunc("/listreadme/", ListReadMe)
	mx.HandleFunc("/listreadme/{page}", ListReadMe)
	mx.HandleFunc("/listreadme/{page}/", ListReadMe)


	//list readme
	mx.HandleFunc("/listblog", ListBlog)
	mx.HandleFunc("/listblog/", ListBlog)
	mx.HandleFunc("/listblog/{page}", ListBlog)
	mx.HandleFunc("/listblog/{page}/", ListBlog)



	//tag
	mx.HandleFunc("/tag/{tag}", TagArticle)
	mx.HandleFunc("/tag/{tag}/", TagArticle)
	mx.HandleFunc("/tag/{tag}/{page}", TagArticle)
	mx.HandleFunc("/tag/{tag}/{page}/", TagArticle)

	//search
	mx.HandleFunc("/search", SearchArticle)

	//article
	mx.HandleFunc("/article/{uk}", ShowArticle)
	mx.HandleFunc("/article/{uk}/", ShowArticle)


	//readme
	mx.HandleFunc("/readme/{uk}", ShowReadMe)
	mx.HandleFunc("/readme/{uk}/", ShowReadMe)


	//blog
	mx.HandleFunc("/blog/{uk}", ShowBlog)
	mx.HandleFunc("/blog/{uk}/", ShowBlog)

	//img
	mx.HandleFunc("/img/{img}", Dummy)
	mx.HandleFunc("/img/{img}/", Dummy)


	//server static
	mx.PathPrefix("/static").Handler(http.FileServer(http.Dir("resource/developerq/")))

	//server static
	mx.PathPrefix("/imgpool").Handler(http.FileServer(http.Dir("")))


	//admin
	mx.HandleFunc("/admin", ShowArticle)
	mx.HandleFunc("/admin/list", ShowArticle)
	mx.HandleFunc("/admin/list/", ShowArticle)

	mx.HandleFunc("/admin/login", ShowArticle)
	mx.HandleFunc("/admin/login/", ShowArticle)

	//for baidu
	mx.HandleFunc("/robots.txt", Robots)
	//not found
	mx.NotFoundHandler = http.HandlerFunc(NotFound)

}
