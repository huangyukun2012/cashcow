package model

import (
//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
//	"encoding/json"
//	"github.com/siddontang/go/log"
	u "developerq/utils"
)


func GenerateListArticlePageVar(esclient *es.Client, page int) *PageVar {
	if page <= 0 || page > 300 {
		return nil
	}

	pv := PageVar{}
	pv.Type = "list"

	query := es.NewMatchAllQuery()

	start := u.PAGEMAX * (page - 1)
	if start <= 0 {
		start = 1
	}
	var size int64
	pv.ListArticle, size = SearchArticle(esclient, query, start, u.PAGEMAX, "scan_time")

	if len(pv.ListArticle) == 0 {
		pv.Type = "lost"
	}

	pv.End = int(size) / 20 + 1
	if pv.End > 300 {
		pv.End = 300
	}
	pv.Current = page

	SetBA(&pv)
	pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
//	pv.RandomUsers = GenerateRandomUsers(esclient, 24)
//	pv.Keywords = GenerateRandomKeywords(esclient, 30)
	return &pv
}
