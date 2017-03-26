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

type ReadMe struct {
	UK          int64
	Title       string
	TitleCN     string
	URL         string
	Content     string
	ContentCN   string
	Tags         string
	Name        string
	Description string
	Stars       int
	Fork        int
	ForkStr     string
	Follow      int
	FollowStr   string
	Language    string
	UpdateTime  int64
	Flag        int
	LangList    []string

	HTitleCN        t.HTML
	HContentCN       t.HTML
	HContent       t.HTML
	SeoKeywords    []string


}


func (readme *ReadMe)Save(db *sql.DB) error {
	//if readme.Title == "" || readme.ContentCN == "" {
	if false {
		return errors.New("Empty readme " + readme.URL)
	} else {
		url := "empty"
		rows, err := db.Query("select url from readme where url = ?", readme.URL)
		if err == nil {
			for rows.Next() {
				rows.Scan(&url)
			}
		}
		fmt.Println("update time = ", readme.UpdateTime)
		fmt.Println("url time = ", readme.URL)


		Logger.Info("url = %s", url)
		if url != readme.URL {
			Logger.Info("Insert readme url = %s", readme.URL)
			_, err := db.Exec("INSERT into readme (uk, title,title_cn, content, content_cn, flag, url, tags, name, description, stars, fork, follow, language, update_time) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", readme.UK, readme.Title, readme.TitleCN, readme.Content, readme.ContentCN, readme.Flag, readme.URL, readme.Tags, readme.Name, readme.Description, readme.Stars, readme.Follow, readme.Fork, readme.Language, readme.UpdateTime)
			return err
		} else {
			Logger.Info("update readme url = %s", readme.URL)
			_, err := db.Exec("update readme set uk = ?, title = ?,title_cn = ?, content = ?, content_cn = ?, flag = ?, tags = ?, name = ?, description = ?, stars = ?, fork = ?, follow = ?, language = ?, update_time = ? where url = ?",readme.UK, readme.Title, readme.TitleCN, readme.Content, readme.ContentCN, readme.Flag, readme.Tags, readme.Name, readme.Description, readme.Stars, readme.Fork, readme.Follow, readme.Language, readme.UpdateTime, readme.URL)
			return err
		}
	}

}


//running
func (readme * ReadMe) FillHtml()*ReadMe {
	//replace < /

	readme.TitleCN = strings.Replace(readme.TitleCN, "</ ", "</", -1)
	readme.ContentCN = strings.Replace(readme.ContentCN, "</ ", "</", -1)
	readme.ContentCN = strings.Replace(readme.ContentCN, "“", "", -1)
	readme.ContentCN = strings.Replace(readme.ContentCN, "”", "", -1)

	readme.HTitleCN = t.HTML(blackfriday.MarkdownCommon([]byte(readme.TitleCN)))
	readme.HContentCN = t.HTML(blackfriday.MarkdownCommon([]byte(readme.ContentCN)))
	readme.HContent = t.HTML(blackfriday.MarkdownCommon([]byte(readme.Content)))

	langs := strings.Split(readme.Language, ",")
	for _, l := range langs {
		l = strings.TrimSpace(l)
		if l != "" {
			readme.LangList = append(readme.LangList, l)
		}
	}

	readme.SeoKeywords = readme.LangList
	keywords := u.Jb.Cut(readme.TitleCN, true)
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			readme.SeoKeywords = append(readme.SeoKeywords, keyword)
		}
	}

	keywords = u.Jb.Cut(readme.Title, true)
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			readme.SeoKeywords = append(readme.SeoKeywords, keyword)
		}
	}

	//convert viewcount/voteconnt
	readme.ForkStr = u.ConvertNumber(int64(readme.Fork))
	readme.FollowStr = u.ConvertNumber(int64(readme.Follow))

	//Logger.Info("keyword = ", readme.SeoKeywords)
	return readme
}



func GetReadMeCount(db *sql.DB) int{
	count := 0
	rows, _ := db.Query("select count(id) from readme")
	for rows.Next() {
		rows.Scan(&count)
	}
	return count
}

func GetSideBarReadMe(db *sql.DB) []ReadMe {
	where := fmt.Sprintf(" limit 0, 20")
	return GetReadMes(db, where)
}

func GetReadMes(db *sql.DB, where string) []ReadMe {
	sql := "select uk, url, update_time, name, title, title_cn, language, fork, follow from readme order by update_time desc" + where;
	rows, err := db.Query(sql)
	readmes := []ReadMe{}
	if err != nil {
		fmt.Println(err.Error())
		Logger.Error(err.Error())
		return readmes
	}

	for rows.Next() {
		readme := ReadMe{}
		rows.Scan( &readme.UK, &readme.URL, &readme.UpdateTime, &readme.Name, &readme.Title, &readme.TitleCN, &readme.Language, &readme.Fork, &readme.Follow)
		readme.FillHtml()
		readmes = append(readmes, readme)
	}
	return readmes
}


func GetReadMe(db *sql.DB, uk int64) *ReadMe {
	sql := "select uk, url, update_time, name, title, title_cn, language, fork, follow, content, content_cn from readme  where uk = ?";
	rows, err := db.Query(sql, uk)

	if err != nil {
		fmt.Println(err.Error())
		Logger.Error(err.Error())
		return nil
	}

	for rows.Next() {
		readme := ReadMe{}
		rows.Scan( &readme.UK, &readme.URL, &readme.UpdateTime, &readme.Name, &readme.Title, &readme.TitleCN, &readme.Language, &readme.Fork, &readme.Follow, &readme.Content, &readme.ContentCN)
		readme.FillHtml()
		return &readme
	}
	return nil
}


func ListReadMePage(db *sql.DB,esclient *es.Client, page int) *PageVar {
	pv := PageVar{}
	pv.Type = "listreadme"

	if page <= 0  {
		pv.Type = "lost"
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
		return nil
	}

	count := GetReadMeCount(db)
	pv.End = count / 20
	where := fmt.Sprintf(" limit %d, 20", (page - 1) * 20)
	pv.ListReadMe = GetReadMes(db, where)

	if len(pv.ListReadMe) == 0 {
		pv.Type = "lost"
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
		return &pv
	}

	pv.Current = page

	SetBA(&pv)
	pv.SideBarReadMe = GetSideBarReadMe(db)
	return &pv
}



func ShowReadMePage(db *sql.DB,esclient *es.Client, uk int64) *PageVar {
	pv := PageVar{}
	pv.Type = "readme"
	readme := GetReadMe(db, uk)
	fmt.Println(uk)

	if readme == nil {
		pv.Type = "lost"
		fmt.Println("lost=====")
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
		return &pv
	}

	pv.ReadMe = *readme
	pv.RandomArticle = GenerateRandomArticle(esclient, 10, pv.ReadMe.Name)
	pv.SideBarReadMe = GetSideBarReadMe(db)
	return &pv
}



//crawling
func (readme *ReadMe)FillAll() {
	//uk
	t := time.Now()
	rand.Seed(t.UnixNano())
	r := int64(rand.Intn(19850720))
	readme.UK = t.Unix() + r

	//update time
	//readme.UpdateTime = t.Format("2006-01-02 15:04:05")
	readme.UpdateTime = int64(t.UnixNano())

	//replace cn prunctuation
	readme.TitleCN = u.ReplaceCNPunctuation(readme.TitleCN)

	readme.TitleCN = strings.Replace(readme.TitleCN, "</ ", "</", -1)
}
