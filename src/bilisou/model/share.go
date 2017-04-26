package model

import (
	"fmt"
// 	u "utils"
//	"github.com/siddontang/go/log"
	es "gopkg.in/olivere/elastic.v3"
	"database/sql"
	"strings"
	u "utils"
	t "html/template"
//	"logging"
	"time"
	"math/rand"
	//	"strconv"
	"errors"

)

/*
func GenerateSharePageVar(esclient *es.Client, dataid string) *PageVar {
	pv := PageVar{}
	pv.Type = "share"

	query := es.NewTermQuery("data_id", dataid)
	Shares, _ := SearchShare(esclient, query, 0, 10, "")

	if len(Shares) == 0 {
		//return nil
		pv.Type = "lost"
	} else {
		pv.Share = Shares[0]
	}

	pv.RandomSharesSimilar = GenerateRandomShares(esclient, 0, 10, pv.Share.Title)
	pv.RandomSharesCategory = GenerateRandomShares(esclient, 0, 10, "")
	pv.UserShares =	GenerateUserShares(esclient, pv.Share.UK, 10)

	pv.RandomShares = GenerateRandomShares(esclient, 0, 10, "")
	pv.RandomUsers = GenerateRandomUsers(esclient, 24)
	pv.Keywords = GenerateRandomKeywords(esclient, 30)
	return &pv
}

*/


type Share struct {
	ShareID   string    `json:"share_id"`
	HTitle    t.HTML
	DataID    string    `json:"data_id"`
	Title     string    `json:"title"`
	FeedType  string //专辑：album 文件或者文件夹：share
	AlbumID   string
	Category  string
	CategoryInt int64   `json:"category"`
	CategoryCN  string
	FeedTime   int64    `json:"feed_time"`
	FeedTimeStr   string
	Size      int64    `json:"size"`
	SizeStr   string    `json:"size_str"`
	FilenamesRaw string `json:"filenames"`
	Filenames []string
	BTFilenames []BTFilename
	Abstract t.HTML
	FileCount int
	UK        string    `json:"uk"`
	Uname     string    `json:"uname"`
	ViewCount int64    `json:"view_count"`
	LikeCount int64    `json:"like_count"`
	LastScan  int64     `json:"last_scan"`

	Link      string `json:"link"`
	HLink     t.HTML
	Source    int    `json:"source"`
	Description string `json:"description"`

	Valid bool
	LastScanStr  string
	SeoKeywords []string
}

type BTFilename struct {
	Name string
	Size string
}

func (share *Share)FillHtml()  {
	share.HTitle = t.HTML(share.Title)
	share.HLink = t.HTML(share.Link)
	share.Category = u.CAT_INT_STR[int(share.CategoryInt)]
	share.CategoryCN = u.CAT_INT_STRCN[int(share.CategoryInt)]
	share.FeedTimeStr  = u.IntToDateStr(share.FeedTime)

	items := u.SplitNames(share.FilenamesRaw)
	if share.Source == 0 {
		abs := ""
		for _, item := range items {
			if strings.TrimSpace(item) != "" {
				abs = abs + item
				share.Filenames = append(share.Filenames, item)
			}
		}
		share.Abstract = t.HTML(abs)
	} else if share.Source == 2 {
		abs := ""
		for _, item := range items {
			ss := u.SplitItem(item)
			if len(ss) > 1 {
				btname := BTFilename{}
				btname.Name = ss[0]
				btname.Size = ss[1]
				abs = abs + ss[0]
				share.BTFilenames = append(share.BTFilenames, btname)
			}
		}
		share.Abstract = t.HTML(abs)
	}

	if share.Source == 0 {
		share.FileCount = len(share.Filenames)
	} else if share.Source == 2 {
		share.FileCount = len(share.BTFilenames)
	}

	if share.Source == 0 {
		share.SizeStr = u.SizeToStr(share.Size)
	}

	share.LastScanStr = u.IntToDateStr(share.LastScan)
	if share.Title != "" {
		keywords := u.Jb.CutForSearch(share.Title, true)
		for _, keyword := range keywords {
			share.SeoKeywords = append(share.SeoKeywords, keyword)
		}
	}
	return
}



//crawling
func (share *Share)FillAll() {
	t := time.Now()
	rand.Seed(t.UnixNano())
	r := int64(rand.Intn(19850720))

	share.DataID = fmt.Sprintf("%d%d", r, t.UnixNano)

	//update time
	//share.UpdateTime = t.Format("2006-01-02 15:04:05")
	share.LastScan = int64(t.Unix())
}


func (share *Share) Save(db *sql.DB) error {
	if share.Title == "" {
		return errors.New("Empty share " + share.Title)
	}

	Logger.Info("save share dataid = %s", share.ShareID)

	_, err := db.Exec("INSERT into sharedata (uk, title, data_id, share_id, category, last_scan, size, size_str, view_count, filenames, description, link, source, feed_time, uname) values (?,?,?,?,?,?,?,?,?, ?,?,?,?,?,?)", share.UK, share.Title,  share.DataID, share.ShareID, share.Category, share.LastScan, share.Size, share.SizeStr, share.ViewCount, share.FilenamesRaw, share.Description, share.Link, share.Source, share.FeedTime, share.Uname)
	return err

}




func GetShareCount(db *sql.DB, where string) int{
	count := 0
	rows, _ := db.Query("select max(id) from sharedata " + where)
	for rows.Next() {
		rows.Scan(&count)
	}
	rows.Close()
	return count
}




func GetShares(db *sql.DB, where string) []Share {
	sql := "select title, data_id, filenames, size, size_str, last_scan, category, source from sharedata " + where;
	fmt.Println(sql)

	rows, _ := db.Query(sql)
	defer rows.Close()
	shares := []Share{}
	/*
	if err != nil {
		Logger.Error(err.Error())
		return shares
	}
	*/

	for rows.Next() {
		share := Share{}
		rows.Scan( &share.Title, &share.DataID, &share.FilenamesRaw, &share.Size, &share.SizeStr, &share.LastScan,  &share.Category,  &share.Source)

		share.FillHtml()
		shares = append(shares, share)
	}
	return shares
}


func GetShare(db *sql.DB, dataID string) *Share {
	sql := "select uk, share_id, data_id, title, size, size_str, last_scan, filenames, view_count, description, link, source, feed_time, uname  from sharedata  where data_id = ? limit 0, 1";
	rows, err := db.Query(sql, dataID)
	defer rows.Close()

	if err != nil {
		Logger.Error(err.Error())
		return nil
	}

	for rows.Next() {
		share := Share{}
		rows.Scan( &share.UK, &share.ShareID, &share.DataID,  &share.Title, &share.Size, &share.SizeStr, &share.LastScan, &share.FilenamesRaw, &share.ViewCount, &share.Description, &share.Link, &share.Source, &share.FeedTime, &share.Uname)
		share.FillHtml()
		return &share
	}
	return nil
}




func ListSharePage(db *sql.DB, page int, category int) *PageVar {

	pv := PageVar{}
	pv.Type = "listshare"
	pv.CategoryInt = category

	if page <= 0 || category < 0 || category > 8  {
		pv.Type = "lost"
		//pv.SideBarShare = GetSideBarShare(db)
		return nil
	}
	where := ""
	//if category != 0 {
	//	where = fmt.Sprintf(" where category = %d  ", category)
	//}

	count := GetShareCount(db, where)
	pv.End = count / 20

	where = where + fmt.Sprintf(" where id < %d  order by id desc limit 0, 20", count - (page - 1) * 20)

	pv.ListShares = GetShares(db, where)

	if len(pv.ListShares) == 0 {
		pv.Type = "lost"
		//pv.SideBarShare = GetSideBarShare(db)
		return &pv
	}

	pv.Current = page

	SetBA(&pv)
	//pv.SideBarShare = GetSideBarShare(db)
	pv.Keywords = GetRandomKeywords(db, 10)
	return &pv
}



func ShowSharePage(db *sql.DB, esclient *es.Client, dataID string) *PageVar {
	pv := PageVar{}
	pv.Type = "share"
	share := GetShare(db, dataID)

	if share == nil {
		pv.Type = "lost"
//		pv.SideBarShare = GetSideBarShare(db)
		return &pv
	}

	pv.Share = *share
	//pv.SideBarShare = GetSideBarShare(db)
	pv.RandomSharesSimilar = GenerateRandomShares(esclient, 0, 10, pv.Share.Title)
	pv.Keywords = GetRandomKeywords(db, 6)


	return &pv
}
