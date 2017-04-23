package model

import (
	"fmt"
//	u "utils"
	"math/rand"
	"time"

	"database/sql"
)

func KeywordHit(db *sql.DB, keyword string) {
	count := 0
	rows, _ := db.Query("select count from keyword where keyword = ?", keyword)
	for rows.Next() {
		rows.Scan(&count)
		count = count + 1;
	}
	rows.Close()

	if count >= 1 {
		stmt, _ := db.Prepare("update keyword set count = ?  where keyword = ?")
		stmt.Exec(count, keyword)
		stmt.Close()
	} else {
		stmt, _ := db.Prepare("insert into keyword(keyword) values(?)")
		stmt.Exec(keyword)
		stmt.Close()
	}
}

func GetRandomKeywords(db *sql.DB, number int) []string {
	res := []string{}

	sql := "select count(id) from keyword"
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

	sql = "select keyword from keyword" + fmt.Sprintf(" limit %d, %d", start, number)
	rows, err = db.Query(sql)
	defer rows.Close()
	if err != nil {
		Logger.Error(err.Error())
		return nil
	}
	var keyword string
	for rows.Next() {
		rows.Scan( &keyword)
		res = append(res, keyword)
	}

	return res
}
