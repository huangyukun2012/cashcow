package model

import (
//	"fmt"
// 	u "utils"
//	"github.com/siddontang/go/log"
	es "gopkg.in/olivere/elastic.v3"

)


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
