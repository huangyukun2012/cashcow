package model

import (
	"database/sql"
	u "utils"
	"logging"
)

//global
var MIN_USER int
var MAX_USER int
var MIN_SHARE int
var MAX_SHARE int
var MIN_KEYWORD int
var MAX_KEYWORD int
var Logger *logging.Logger

type UserInfo struct {
	Id           int64
	Avatar_url   string
	Fans_count   int64
	Follow_count int64
	Intro        string
	Pubshare_count  int64
	Uk            string
	Uname         string
}


type User struct {
	UK          string
	Uname       string
	FansCount   string
	FollowCount string
	PubshareCount    string
	AvatarURL   string
	Intro       string
}

type Keyword struct {
	Keyword string
	Count   int64
}


func UserInfoToUser(uinfo UserInfo) User {
/*	Id           int64
	Avatar_url   string
	Fans_count   int64
	Follow_count int64
	Intro        string
	Pubshare_count  int64
	Uk            int64
	Uname         string
*/
	user := User{}
	user.UK         =  uinfo.Uk
	user.Uname      =  uinfo.Uname
	user.FansCount  =  u.IntToStr(uinfo.Fans_count)
	user.FollowCount = u.IntToStr(uinfo.Follow_count)
	user.PubshareCount   = u.IntToStr(uinfo.Pubshare_count)
	user.AvatarURL   =  uinfo.Avatar_url
	user.Intro       =  uinfo.Intro
	return user

}

func SetCategory(pv *PageVar, category int){
	pv.CategoryInt = category

	cat, ok := u.CAT_INT_STR[category]
	if ok {
		pv.Category = cat
	}

	cat, ok = u.CAT_INT_STRCN[category]
	if ok {
		pv.CategoryCN = cat
	}
}


func GetKeywordMaxMinID(db *sql.DB) (int, int) {
	var max int
	var min int
	sql := "select max(id), min(id) from keyword"
	rows, _ := db.Query(sql)
	//u.CheckErr(err)
	for rows.Next() {
		rows.Scan(&max, &min)
	}
	rows.Close()
	return max, min
}


func GetShareMaxMinID(db *sql.DB) (int, int) {
	var max int
	var min int
	sql := "select max(id), min(id) from sharedata"
	rows, _ := db.Query(sql)
	//u.CheckErr(err)
	for rows.Next() {
		rows.Scan(&max, &min)
	}
	rows.Close()
	return max, min
}

func GetUserMaxMINID(db *sql.DB) (int, int) {
	var max int
	var min int
	sql := "select max(id), min(id) from uinfo"
	rows, _ := db.Query(sql)
	//u.CheckErr(err)
	for rows.Next() {
		rows.Scan(&max, &min)
	}
	rows.Close()
	return max, min
}

/*
func UpdateCategory(db *sql.DB) {
	max, min := GetShareMaxMinID(db)
	for i:=min; i <= max; i ++ {
		s := "select title from sharedata where id = %d"
		s = fmt.Sprintf(s, i)
		rows, _ := db.Query(s)
		//u.CheckErr(err)
		var tt sql.NullString
		for rows.Next() {
			rows.Scan(&tt)
		}
		rows.Close()
		if tt.Valid {
			c := u.GetCategoryFromName(tt.String)
			us := "update sharedata set category = ? where id = ?"
			//us = fmt.Sprintf(us, c, i)
			stmt, _ := db.Prepare(us)
			stmt.Exec(c,i)
			stmt.Close()
			//res.RowsAffected()
//			log.Info(us)
		}
	}
}

*/

