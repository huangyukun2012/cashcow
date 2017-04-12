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
	u "developerq/utils"

)

type Blog struct {
	UK          int64
	Title       string
	Abstract    string
	URL         string
	Content     string
	Tag        string
	UpdateTime  int64
	Author      string
	HTitle        t.HTML
	HContent       t.HTML
	SeoKeywords    []string
}


func (blog *Blog)Save(db *sql.DB) error {
	//if blog.Title == "" || blog.ContentCN == "" {
	if false {
		return errors.New("Empty blog " + blog.URL)
	} else {
		url := "empty"
		rows, err := db.Query("select url from blog where url = ?", blog.URL)
		if err == nil {
			for rows.Next() {
				rows.Scan(&url)
			}
		}
		rows.Close()
		fmt.Println("update time = ", blog.UpdateTime)
		fmt.Println("url time = ", blog.URL)


		Logger.Info("url = %s", url)
		if url != blog.URL {
			Logger.Info("Insert blog url = %s", blog.URL)
			_, err := db.Exec("INSERT into blog (uk, title, content,  url, tag, name,  update_time, author) values (?,?,?,?,?,?,?,?)", blog.UK, blog.Title,  blog.Content, blog.URL, blog.Tag,  blog.UpdateTime, blog.Author)
			return err
		} else {
			Logger.Info("update blog url = %s", blog.URL)
			_, err := db.Exec("update blog set uk = ?, title = ?, content = ?, tag = ?,   update_time = ?, author = ? where url = ?",blog.UK, blog.Title, blog.Content,  blog.Tag,  blog.UpdateTime, blog.Author, blog.URL )
			return err
		}
	}

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
	where := fmt.Sprintf(" limit 0, 20")
	return GetBlogs(db, where)
}

func GetBlogs(db *sql.DB, where string) []Blog {
	sql := "select uk, url, update_time, title, abstract from blog order by update_time desc" + where;
	rows, err := db.Query(sql)
	blogs := []Blog{}
	if err != nil {
		fmt.Println(err.Error())
		Logger.Error(err.Error())
		return blogs
	}

	for rows.Next() {
		blog := Blog{}
		rows.Scan( &blog.UK, &blog.URL, &blog.UpdateTime,  &blog.Title, &blog.Abstract)
		blog.FillHtml()
		blogs = append(blogs, blog)
	}
	rows.Close()
	return blogs
}


func GetBlog(db *sql.DB, uk int64) *Blog {
	sql := "select uk, url, update_time, title,  content, abstract from blog  where uk = ?";
	rows, err := db.Query(sql, uk)

	if err != nil {
		fmt.Println(err.Error())
		Logger.Error(err.Error())
		return nil
	}

	for rows.Next() {
		blog := Blog{}
		rows.Scan( &blog.UK, &blog.URL, &blog.UpdateTime,  &blog.Title,  &blog.Content, &blog.Abstract)
		blog.FillHtml()
		return &blog
	}
	rows.Close()
	return nil
}


func ListBlogPage(db *sql.DB,esclient *es.Client, page int) *PageVar {
	pv := PageVar{}
	pv.Type = "listblog"

	if page <= 0  {
		pv.Type = "lost"
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
		return nil
	}

	count := GetBlogCount(db)
	pv.End = count / 20
	where := fmt.Sprintf(" limit %d, 20", (page - 1) * 20)
	pv.ListBlog = GetBlogs(db, where)

	if len(pv.ListBlog) == 0 {
		pv.Type = "lost"
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
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
	fmt.Println(uk)

	if blog == nil {
		pv.Type = "lost"
		fmt.Println("lost=====")
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
		return &pv
	}

	pv.Blog = *blog
	pv.RandomArticle = GenerateRandomArticle(esclient, 10, pv.Blog.Title)
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
