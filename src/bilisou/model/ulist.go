package model

import (
//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
//	"github.com/siddontang/go/log"
	u "bilisou/utils"
)


func GenerateUlistPageVar(esclient *es.Client, page int) *PageVar {
	if page <= 0 {
		return nil
	}

	pv := PageVar{}
	pv.Type = "ulist"


	query := es.NewTermQuery("search", 1)

	start := u.PAGEMAX * (page - 1)
	if start <= 0 {
		start = 1
	}
	var size int64
	pv.ListUsers, size = SearchUser(esclient, query, start, u.PAGEMAX)

	if len(pv.ListUsers) == 0 {
		pv.Type = "lost"
	}

	pv.End = int(size) / 20 + 1
	pv.Current = page
	SetBA(&pv)

	pv.RandomShares = GenerateRandomShares(esclient, 0, 10, "")
	pv.RandomUsers = GenerateRandomUsers(esclient, 24)
	pv.Keywords = GenerateRandomKeywords(esclient, 30)
	return &pv
}
