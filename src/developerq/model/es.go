package model

import (
	//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
	"encoding/json"
	"github.com/siddontang/go/log"
	t "html/template"
//	u "utils"
	"strings"
	"math/rand"
	"time"
	"html"
)



func SearchArticle(esclient *es.Client, query es.Query, start int, size int, sort string)([]Article, int64) {
	searchResult := Search(esclient, "developerq_article", query, start, size, sort)
	if searchResult == nil {
		return nil, 0
	}

	articles := []Article{}
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {
			article := Article{}

			err := json.Unmarshal(*hit.Source, &article)
			if err != nil {
				log.Error("Failed to read search result", err)
			}

			if hit.Highlight != nil && len(hit.Highlight) != 0 {
				if hl, found := hit.Highlight["title_cn_raw"]; found {
					// escape original tag
					hl[0] = html.EscapeString(hl[0])
					// replace highligst tag
					hl[0] = strings.Replace(hl[0], "#markstart#", "<mark>", -1)
					hl[0] = strings.Replace(hl[0], "#markend#", "</mark>", -1)
					article.HTitleHighlight = t.HTML(hl[0])
				} else {
					article.TitleCN = html.EscapeString(article.TitleCN)
					article.HTitleHighlight = t.HTML(article.TitleCN)
				}

				if hl, found := hit.Highlight["question_cn_raw"]; found {
					hl[0] = html.EscapeString(hl[0])
					// replace highligst tag
					hl[0] = strings.Replace(hl[0], "#markstart#", "<mark>", -1)
					hl[0] = strings.Replace(hl[0], "#markend#", "</mark>", -1)
					article.HQuestionHighlight = t.HTML(hl[0])
				} else {
					article.QuestionCNRaw = html.EscapeString(article.QuestionCNRaw)
					article.HQuestionHighlight = t.HTML(article.QuestionCNRaw)
				}
			} else {
			//	article.TitleCN = html.EscapeString(article.TitleCN)
				//article.QuestionCNRaw = html.EscapeString(article.QuestionCNRaw)
				article.HTitleHighlight = t.HTML(article.TitleCN)
				article.HQuestionHighlight = t.HTML(article.QuestionCNRaw)

			}

			article.FillHtml()

			articles = append(articles, article)
		}
	} else {
		return nil, 0
	}

	return articles, searchResult.Hits.TotalHits
}


func SearchKeyword(esclient *es.Client, query es.Query, start int, size int)([]Keyword, int64) {
	searchResult := Search(esclient, "developerq_keyword", query , start, size, "count")
	if searchResult == nil {
		return nil, 0
	}
	keywords := []Keyword{}
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {
			k := Keyword{}
			err := json.Unmarshal(*hit.Source, &k)
			if err != nil {
				log.Error("Failed to read search result", err)
			}
			keywords = append(keywords, k)
		}
	} else {
		return nil, 0
	}
	return keywords, searchResult.Hits.TotalHits
}

func Search(esclient *es.Client, index string,  query es.Query, start int, size int, sort string) *es.SearchResult {
	// Specify highlighter
	if start < 0 {
		start = 0
	}
	if size <=0 {
		size = 1
	}

	hl := es.NewHighlight()
	hl = hl.Fields(es.NewHighlighterField("title_cn_raw"), es.NewHighlighterField("question_cn_raw"))
	hl = hl.PreTags("#markstart#").PostTags("#markend#")
	hl = hl.Encoder("utf-8")

	searchService := esclient.Search().
		Index(index).
		Highlight(hl).
		Query(query).
		From(start).Size(size).
		Pretty(true)

	if sort != "" {
		searchService = searchService.Sort(sort, false)
	}

	searchResult, err := searchService.Do()                // execute
	if err != nil {
		log.Info(err)
		return nil
	}

	log.Info("Query took ", searchResult.TookInMillis, " msec")
	// Here's how you iterate through the search results with full control over each step.
	log.Info("Found a total of ", searchResult.Hits.TotalHits)
	return searchResult
}

func GenerateRandomArticle(esclient *es.Client, size int, keyword string) []Article{
	start := 0
	boolQuery := es.NewBoolQuery()
	if keyword != "" {
		//qsq := es.NewQueryStringQuery(keyword)
		//qsq = qsq.Escape(true)

		mq1 := es.NewMatchQuery("title_raw", keyword)
		mq2 := es.NewMatchQuery("question_raw", keyword)

		boolQuery.Should(mq1)
		boolQuery.Should(mq2)
		start = 1
	} else {
		for i := 0; i < size; i ++ {
			rand.Seed(time.Now().UnixNano())
			id := rand.Intn(MaxArticle - MinArticle) + MinArticle
			boolQuery.Should(es.NewTermQuery("id", id))
		}
	}

	randomArticle, _ := SearchArticle(esclient, boolQuery, start, size, "")
	return randomArticle
}
