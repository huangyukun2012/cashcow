package model

import (
//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
//	"encoding/json"
//	"github.com/siddontang/go/log"
	u "developerq/utils"
)


func GenerateShowArticlePageVar(esclient *es.Client, uk int64 ) *PageVar {
	pv := PageVar{}
	pv.Type = "show"
//	pv.Keyword = keyword

	//boolQuery := es.NewBoolQuery()

	query := es.NewTermQuery("uk", uk)

	//var size int64
	as, _ := SearchArticle(esclient, query, 0, u.PAGEMAX, "")

	if len(as) == 0 {
		pv.Type = "lost"
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
	} else {
		pv.Article = as[0]
		pv.RandomArticle = GenerateRandomArticle(esclient, 10, pv.Article.TitleRaw)
	}

	return &pv
}
