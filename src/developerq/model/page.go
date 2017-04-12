package model


import (
//	"fmt"
//	u "developerq/utils"
)


type PageVar struct {
	Type              string
	Keyword           string
	Tag               string
	Tags              []string
	//for show
	Article           Article
	Blog              Blog
	//for search
	SearchArticle     []Article
	TotalFound        int64
	//for list
	ListArticle       []Article
	ListBlog          []Blog
	//for tag
	//ListTagArticle     []Article
	//for 404
	RandomArticle     []Article

	ReadMe            ReadMe
	ListReadMe        []ReadMe

	SideBarReadMe        []ReadMe
	SideBarBlog        []Blog

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
	left := 5
	right := 5
	if(pv.Current > 2) {
		pv.Previous = pv.Current - 1;
	}
	if pv.Next < pv.End {
		pv.Next = pv.Current + 1
	}

	pl := pv.Current - left
	pr := pv.Current + 1

	for ; pl < pv.Current; pl ++ {
		if pl > 0 {
			pv.Before = append(pv.Before, pl)
		} else {
			right = right + 1
		}
	}

	for ; pr < pv.Current + right && pr <pv.End; pr ++ {
		pv.After = append(pv.After, pr)
	}

/*
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
*/

}
