package model


import (
//	"fmt"
	u "developerq/utils"
)


type PageVar struct {
	Type              string
	Keyword           string
	Tag               string
	Tags              []string
	//for show
	Article           Article
	//for search
	SearchArticle     []Article
	TotalFound        int64
	//for list
	ListArticle       []Article
	//for tag
	//ListTagArticle     []Article
	//for 404
	RandomArticle     []Article

	ReadMe            ReadMe
	ListReadMe        []ReadMe

	SideBarReadMe        []ReadMe

	//for paging
	Current       int
	Start         int
	End    int
	Previous int
	Next   int
	Before []int
	After  []int
}

func SetBA(pv *PageVar) {
	if(pv.Current > 2) {
		pv.Previous = pv.Current - 1;
	}
	if pv.Next < pv.End {
		pv.Next = pv.Current + 1
	}

	pp := pv.Current - u.NAVMAX

	if pp < 0 {
		pp = 1
	}

	for ; (pv.Current > pp) && (pp >= 1); pp ++ {
		pv.Before = append(pv.Before, pp)
	}

	pp = pv.Current + 1
	for ; ((pv.Current + u.NAVMAX) >= pp) && (pp <= pv.End); pp ++ {
		pv.After = append(pv.After, pp)
	}

}
