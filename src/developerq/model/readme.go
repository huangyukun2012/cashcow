package model

import (
	//	"fmt"
//	"time"
	"errors"
	"database/sql"
//	"math/rand"
//	t "html/template"
	//"strings"
//	u "developerq/utils"
//	h "github.com/jaytaylor/html2text"
//	"logging"
//	"github.com/russross/blackfriday"

)

type ReadMe struct {
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
	Follow      int
	Language    string
	UpdateTime  int64
	Flag        int
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


		Logger.Info("url = %s", url)
		if url != readme.URL {
			Logger.Info("Insert readme url = %s", readme.URL)
			_, err := db.Exec("INSERT into readme (title,title_cn, content, content_cn, flag, url, tags, name, description, stars, fork, follow, language, update_time) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)", readme.Title, readme.TitleCN, readme.Content, readme.ContentCN, readme.Flag, readme.URL, readme.Tags, readme.Name, readme.Description, readme.Stars, readme.Follow, readme.Fork, readme.Language, readme.UpdateTime)
			return err
		} else {
			Logger.Info("update readme url = %s", readme.URL)
			_, err := db.Exec("update readme  set title = ?,title_cn = ?, content = ?, content_cn = ?, flag = ?, tags = ?, name = ?, description = ?, stars = ?, fork = ?, follow = ?, language = ?, update_time = ? where url = ?", readme.Title, readme.TitleCN, readme.Content, readme.ContentCN, readme.Flag, readme.Tags, readme.Name, readme.Description, readme.Stars, readme.Fork, readme.Follow, readme.Language, readme.UpdateTime, readme.URL)
			return err
		}
	}

}
