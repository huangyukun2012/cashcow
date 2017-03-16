package model

import (
//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
//	"encoding/json"
	u "developerq/utils"
)


func GenerateSearchArticlePageVar(esclient *es.Client, keyword string, page int) *PageVar {
	if page <= 0 || page > 300 {
		return nil
	}

	pv := PageVar{}
	pv.Type = "search"
	pv.Keyword = keyword

	boolQuery := es.NewBoolQuery()
	boolQuery.Should(es.NewMatchQuery("title_raw", keyword))
	boolQuery.Should(es.NewMatchQuery("title_cn_raw", keyword))
	boolQuery.Should(es.NewMatchQuery("question_raw", keyword))
	boolQuery.Should(es.NewMatchQuery("question_cn_raw", keyword))
	boolQuery.Should(es.NewMatchQuery("answer_raw", keyword))
	boolQuery.Should(es.NewMatchQuery("answer_cn_raw", keyword))


	start := u.PAGEMAX * (page - 1)
	if start <= 0 {
		start = 1
	}
	var size int64
	pv.SearchArticle, size = SearchArticle(esclient, boolQuery, start, u.PAGEMAX, "")

	/*
	if len(pv.SearchArticle) == 0 {
		pv.Type = "lost"
	}
	*/
	pv.TotalFound = size

	pv.End = int(size) / 20 + 1
	pv.Current = page
	if pv.End > 300 {
		pv.End = 300
	}

	SetBA(&pv)
	pv.RandomArticle = GenerateRandomArticle(esclient, 10, "")
	return &pv
}
