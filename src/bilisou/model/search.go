package model

import (
//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
//	"encoding/json"
//	"github.com/siddontang/go/log"
	u "bilisou/utils"
)


func GenerateSearchPageVar(esclient *es.Client, category int, keyword string, page int) *PageVar {
	if page <= 0 {
		return nil
	}

	pv := PageVar{}
	pv.Type = "search"
	pv.Keyword = keyword

	boolQuery := es.NewBoolQuery()

	query := es.NewMatchQuery("title", keyword)
	boolQuery.Should(query)

	if category != 0 {
		boolQuery.Must(es.NewTermQuery("category", category))
	}

	start := u.PAGEMAX * (page - 1)
	if start <= 0 {
		start = 1
	}
	var size int64
	pv.SearchShares, size = SearchShare(esclient, boolQuery, start, u.PAGEMAX, "")
	//log.Info(pv.SearchShares

	if len(pv.SearchShares) == 0 {
		pv.Type = "lost"
	}

	pv.End = int(size) / 20 + 1
	if pv.End > 30000 {
		pv.End = 30000
	}
	pv.Current = page

	SetBA(&pv)
	SetCategory(&pv, category)

	pv.RandomUsers = GenerateRandomUsers(esclient, 24)
	pv.Keywords = GenerateRandomKeywords(esclient, 30)
	return &pv
}
