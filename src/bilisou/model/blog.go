package model

import (
	"fmt"
//	"time"
	"errors"
	"database/sql"
	t "html/template"
//	"math/rand"
	//"strings"
//	u "developerq/utils"
//	h "github.com/jaytaylor/html2text"
//	"logging"
	es "gopkg.in/olivere/elastic.v3"
	"github.com/russross/blackfriday"
	"time"
	"math/rand"
//	t "html/template"
	"strings"
	u "utils"

)

type Blog struct {
	UK          int64
	Title       string
	Abstract    string
	URL         string
	Content     string
	Tag        string
	UpdateTime  int64
	UpdateTimeStr string
	Author      string
	HTitle        t.HTML
	HContent       t.HTML
	SeoKeywords    []string
	Valid       bool
	Category    int
	CategoryStr  string
}


func (blog *Blog)Save(db *sql.DB) error {
	//if blog.Title == "" || blog.ContentCN == "" {
	if false {
		return errors.New("Empty blog " + blog.URL)
	} else {
		url := "empty"
		if blog.Title == "" {
			return errors.New("Empty blog " + blog.URL)
		}

		Logger.Info("url = %s", url)
		Logger.Info("Insert blog url = %s", blog.URL)
		_, err := db.Exec("INSERT into blog (uk, title, content,  url, tag,  update_time, author, category) values (?,?,?,?,?,?,?, ?)", blog.UK, blog.Title,  blog.Content, blog.URL, blog.Tag,  blog.UpdateTime, blog.Author, blog.Category)
		return err
	}

}

var CatMap = map[int]string {
	1 : "搞笑段子",
	2 : "糗事百科",
	3 : "明星八卦",
	4 : "深度好文",
	5 : "时事点评",
	6 : "心灵鸡汤",
	7 : "养生专家",
	8 : "游戏世界",
}

//running
func (blog * Blog) FillHtml()*Blog {
	//replace < /

	blog.HContent = t.HTML(blackfriday.MarkdownCommon([]byte(blog.Content)))


	blog.SeoKeywords = []string{}
	keywords := u.Jb.Cut(blog.Title, true)
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			blog.SeoKeywords = append(blog.SeoKeywords, keyword)
		}
	}

	keywords = u.Jb.Cut(blog.Title, true)
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			blog.SeoKeywords = append(blog.SeoKeywords, keyword)
		}
	}
	blog.UpdateTimeStr = u.IntToDateStr(blog.UpdateTime/1000000000)

	blog.CategoryStr = CatMap[blog.Category]

	return blog
}



func GetBlogCount(db *sql.DB) int{
	count := 0
	rows, _ := db.Query("select count(id) from blog")
	for rows.Next() {
		rows.Scan(&count)
	}
	rows.Close()
	return count
}


func GetSideBarBlog(db *sql.DB) []Blog {
	sql := "select count(id) from blog"
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		Logger.Error(err.Error())
		return nil
	}
	var size int
	for rows.Next() {
		rows.Scan( &size)
	}


	if size <= 0 {
		size = 1
	}
	rand.Seed(time.Now().UnixNano())
	start := rand.Intn(size )

	where := fmt.Sprintf(" limit %d, 10", start)
	return GetBlogs(db, where)
}


func GetBlogs(db *sql.DB, where string) []Blog {
	sql := "select uk, url, update_time, title, tag,  category, abstract from blog order by update_time desc" + where;

	rows, _ := db.Query(sql)
	defer rows.Close()
	blogs := []Blog{}
	/*
	if err != nil {
		Logger.Error(err.Error())
		return blogs
	}
	*/

	for rows.Next() {
		blog := Blog{}
		rows.Scan( &blog.UK, &blog.URL, &blog.UpdateTime,  &blog.Title, &blog.Tag,  &blog.Category,  &blog.Abstract)

		blog.FillHtml()
		blogs = append(blogs, blog)
	}
	return blogs
}


func GetBlog(db *sql.DB, uk int64) *Blog {
	sql := "select uk, url, update_time, title,  category, content  from blog  where uk = ?";
	rows, err := db.Query(sql, uk)
	defer rows.Close()

	if err != nil {
		fmt.Println(err.Error())
		Logger.Error(err.Error())
		return nil
	}

	for rows.Next() {
		blog := Blog{}
		rows.Scan( &blog.UK, &blog.URL, &blog.UpdateTime,  &blog.Title, &blog.Category, &blog.Content)
		blog.FillHtml()
		return &blog
	}
	return nil
}


func ListBlogPage(db *sql.DB,esclient *es.Client, page int, category int) *PageVar {

	pv := PageVar{}
	pv.Type = "listblog"
	pv.CategoryInt = category

	if page <= 0  {
		pv.Type = "lost"
		pv.SideBarBlog = GetSideBarBlog(db)
		return nil
	}

	count := GetBlogCount(db)
	pv.End = count / 20
	where := fmt.Sprintf(" limit %d, 20", (page - 1) * 20)
	if category != 0 {
		where = where + fmt.Sprintf(" and category = %d", category)
	}

	pv.ListBlog = GetBlogs(db, where)

	if len(pv.ListBlog) == 0 {
		pv.Type = "lost"
		pv.SideBarBlog = GetSideBarBlog(db)
		return &pv
	}

	pv.Current = page

	SetBA(&pv)
	pv.SideBarBlog = GetSideBarBlog(db)

	return &pv
}



func ShowBlogPage(db *sql.DB,esclient *es.Client, uk int64) *PageVar {
	pv := PageVar{}
	pv.Type = "blog"
	blog := GetBlog(db, uk)

	if blog == nil {
		pv.Type = "lost"
		pv.SideBarBlog = GetSideBarBlog(db)
		return &pv
	}

	pv.Blog = *blog
	pv.SideBarBlog = GetSideBarBlog(db)
	return &pv
}



//crawling
func (blog *Blog)FillAll() {
	//uk
	t := time.Now()
	rand.Seed(t.UnixNano())
	r := int64(rand.Intn(19850720))
	blog.UK = t.Unix() + r

	//update time
	//blog.UpdateTime = t.Format("2006-01-02 15:04:05")
	blog.UpdateTime = int64(t.UnixNano())
}
