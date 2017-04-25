package model

import (
	//	"fmt"
	es "gopkg.in/olivere/elastic.v3"
	"encoding/json"
	//"github.com/siddontang/go/log"
//	u "utils"
	"math/rand"
	"time"
)


var TotalShares int64
var TotalUsers int64
var TotalKeywords int64


func SearchShare(esclient *es.Client, query es.Query, start int, size int, sort string)([]Share, int64, int64) {
	searchResult := Search(esclient, "bilisou_sharedata", query, start, size, sort)
	if searchResult == nil {
		return nil, 0, 0
	}

	shares := []Share{}
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {
			sd := Share{}

			err := json.Unmarshal(*hit.Source, &sd)
			if err != nil {
				Logger.Error("Failed to read search result %s", err.Error())
			}

			if hit.Highlight != nil && len(hit.Highlight) != 0 {
				if hl, found := hit.Highlight["title"]; found {
					sd.Title = hl[0]
				}

				//if hl, found := hit.Highlight["filenames"]; found {
				//	sd.FilenamesRaw = hl[0]
				//}
			}
			if len(sd.FilenamesRaw) > 180 {
				sd.FilenamesRaw = sd.FilenamesRaw[0:180] + "..."
			}

			sd.FillHtml()
			//trucate the filenames
					shares = append(shares, sd)
		}
	} else {
		return nil, 0, 0
	}

	return shares, searchResult.Hits.TotalHits, searchResult.TookInMillis
}

func SearchUser(esclient *es.Client, query es.Query, start int, size int)([]User, int64) {
	searchResult := Search(esclient, "bilisou_uinfo", query , start, size, "")
	if searchResult == nil {
		return nil, 0
	}
	users := []User{}
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {
			u := UserInfo{}

			err := json.Unmarshal(*hit.Source, &u)
			if err != nil {
				Logger.Error("Failed to read search result, %s", err.Error())
			}
			user := UserInfoToUser(u)
			users = append(users, user)
		}
	} else {
		return nil, 0
	}
	return users, searchResult.Hits.TotalHits
}


func SearchKeyword(esclient *es.Client, query es.Query, start int, size int)([]Keyword, int64) {
	searchResult := Search(esclient, "bilisou_keyword", query , start, size, "count")
	if searchResult == nil {
		return nil, 0
	}
	keywords := []Keyword{}
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {
			k := Keyword{}
			err := json.Unmarshal(*hit.Source, &k)
			if err != nil {
				Logger.Error("Failed to read search result, %s", err.Error())
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
	hl = hl.Fields(es.NewHighlighterField("title"))
	hl = hl.PreTags("<mark>").PostTags("</mark>")
	hl = hl.Encoder("utf-8")

	/*
	hlf := es.NewHighlight()
	hlf = hl.Fields(es.NewHighlighterField("filenames"))
	hlf = hl.PreTags("<mark>").PostTags("</mark>")
	hlf = hl.Encoder("utf-8")
	*/

	searchService := esclient.Search().
		Index(index).
		Highlight(hl).
	//	Highlight(hlf).
		Query(query).
		From(start).Size(size).
		Pretty(true)

	if sort != "" {
		searchService = searchService.Sort(sort, false)
	}

	searchResult, err := searchService.Do()                // execute
	if err != nil {
		Logger.Error(err.Error())
		return nil
	}

	Logger.Info("Query took %d msec", searchResult.TookInMillis)
	// Here's how you iterate through the search results with full control over each step.
	Logger.Info("Found a total of %d", searchResult.Hits.TotalHits)
	return searchResult
}

func GenerateRandomShares(esclient *es.Client, category int, size int, keyword string) []Share{
	boolQuery := es.NewBoolQuery()
	for i := 0; i < size; i ++ {
		rand.Seed(time.Now().UnixNano())
		id := rand.Intn(MAX_SHARE - MIN_SHARE) + MIN_SHARE
		boolQuery.Should(es.NewTermQuery("id", id))
	}

	if category != 0 {
		boolQuery.Must(es.NewTermQuery("category", category))
	}
	randomShares, _, _ := SearchShare(esclient, boolQuery, 0, size, "")
	return randomShares
}
