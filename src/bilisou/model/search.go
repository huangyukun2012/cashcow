package model

import (
//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
//	"encoding/json"
//	"github.com/siddontang/go/log"
	u "utils"
	"database/sql"
)


func GenerateSearchPageVar(esclient *es.Client, db *sql.DB, category int, keyword string, page int) *PageVar {
	if page <= 0 {
		return nil
	}

	pv := PageVar{}
	pv.Type = "search"
	pv.Keyword = keyword

	boolQuery := es.NewBoolQuery()

	query := es.NewMultiMatchQuery(keyword, "title", "filenames")
	//query := es.NewMultiMatchQuery(keyword, "title")
	//query := es.NewSimpleQueryStringQuery(keyword)
	boolQuery.Should(query)

	/*
	if category != 0 {
		boolQuery.Must(es.NewTermQuery("category", category))
	}
	*/

	start := u.PAGEMAX * (page - 1)
	if start <= 0 {
		start = 1
	}
	var size int64
	var time int64

	pv.SearchShares, size, time= SearchShare(esclient, boolQuery, start, u.PAGEMAX, "")
	if size > 1000 {
		size = 1000
	}
	pv.SearchTime = time
	pv.SearchResult = size

	if len(pv.SearchShares) == 0 {
		pv.Type = "lost"
	}

	pv.End = int(size) / 20 + 1

	pv.Current = page

	SetBA(&pv)
	SetCategory(&pv, category)

//	pv.RandomUsers = GenerateRandomUsers(esclient, 24)
	pv.Keywords = GetRandomKeywords(db, 10)
	return &pv
}
