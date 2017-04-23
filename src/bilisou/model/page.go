package model


import (
//	"fmt"
//	u "utils"
)


type PageVar struct {
	Type          string
	CategoryInt   int
	Category      string
	CategoryCN    string
	Keyword       string
	Tag           string
	Blog          Blog
	ListBlog         []Blog
	SideBarBlog   []Blog

	//for paging
	Current       int
	Start         int
	End    int
	Previous int
	Next   int
	Before []int
	After  []int
	SearchTime int64
	SearchResult int64


	User              User
	Share             Share
	SearchShares        []Share
	ListShares        []Share
	ListUsers         []User
	RandomUsers       []User
	UserShares        []Share
	RandomShares      []Share
	RandomSharesCategory      []Share
	RandomSharesSimilar      []Share
	Keywords          []string
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
