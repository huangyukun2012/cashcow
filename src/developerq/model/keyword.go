package model

import (
	//"fmt"
	//	u "utils"
	//"github.com/siddontang/go/log"
	"database/sql"
)

func KeywordHit(db *sql.DB, keyword string) {
	count := 1
	rows, _ := db.Query("select count from keyword where keyword = ?", keyword)
	for rows.Next() {
		rows.Scan(&count)
		count = count + 1;
	}

	if count == 1 {
		stmt, _ := db.Prepare("update keyword set count = ?  where keyword = ?")
		stmt.Exec(count, keyword)
		stmt.Close()
	} else {
		stmt, _ := db.Prepare("insert into keyword(keyword) values(?)")
		stmt.Exec(keyword)
		stmt.Close()
	}
}
