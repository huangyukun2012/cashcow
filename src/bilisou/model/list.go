package model

import (
//	"fmt"
	u "bilisou/utils"
//	"github.com/siddontang/go/log"
	es "gopkg.in/olivere/elastic.v3"
)



func GenerateListPageVar(esclient *es.Client, category int,  page int) *PageVar {
	pv := PageVar{}
	if page <= 0 {
		pv.Type = "lost"
		pv.RandomShares = GenerateRandomShares(esclient, 0, 10, "")
		pv.RandomUsers = GenerateRandomUsers(esclient, 24)
		pv.Keywords = GenerateRandomKeywords(esclient, 30)

		return &pv
	}


	pv.Type = "list"

	boolQuery := es.NewBoolQuery()
	query := es.NewTermQuery("search", 1)
	boolQuery.Should(query)
	if category != 0 {
		boolQuery.Must(es.NewTermQuery("category", category))
	}


	start := u.PAGEMAX * (page - 1)
	if start <= 0 {
		start = 1
	}

	var size int64
	pv.ListShares, size = SearchShare(esclient, boolQuery, start, u.PAGEMAX, "last_scan")

	if len(pv.ListShares) == 0 {
		pv.Type = "lost"
	}

	pv.End = int(size) / 20 + 1
	if pv.End > 30000 {
		pv.End = 30000
	}
	pv.Current = page

	SetBA(&pv)
	SetCategory(&pv, category)


	pv.RandomShares = GenerateRandomShares(esclient, 0, 10, "")
	pv.RandomUsers = GenerateRandomUsers(esclient, 24)
	pv.Keywords = GenerateRandomKeywords(esclient, 30)
	return &pv
}
