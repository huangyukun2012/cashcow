package model

import (
//	"fmt"
	u "bilisou/utils"
	es "gopkg.in/olivere/elastic.v3"
)


func GenerateUserPageVar(esclient *es.Client, uk string, page int) *PageVar {
	if page <= 0 {
		return nil
	}

	pv := PageVar{}
	pv.Type = "user"

	query := es.NewTermQuery("uk", uk)
	users, size := SearchUser(esclient, query, 0, 10)

	if len(users) < 1 {
		pv.Type = "lost"
	} else {
		pv.User = users[0]
	}

	query = es.NewTermQuery("uk", uk)

	start := u.PAGEMAX * (page - 1)
	if start <= 0 {
		start = 1
	}

	pv.UserShares, size = SearchShare(esclient, query, start, u.PAGEMAX, "last_scan")

	pv.Current = page
	pv.End = int(size) / 20 + 1

	SetBA(&pv)

	pv.RandomShares = GenerateRandomShares(esclient, 0, 10, "")
	pv.RandomUsers = GenerateRandomUsers(esclient, 24)
	pv.Keywords = GenerateRandomKeywords(esclient, 30)

	return &pv
}
