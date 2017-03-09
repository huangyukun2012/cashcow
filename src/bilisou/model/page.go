package model


import (
//	"fmt"
//	"github.com/siddontang/go/log"
	u "bilisou/utils"
)


type PageVar struct {
	Type          string
	CategoryInt   int
	Category      string
	CategoryCN    string
	Keyword       string

	//for paging
	Current       int
	Start         int
	End    int
	Previous int
	Next   int
	Before []int
	After  []int


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
	Keywords          []Keyword
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
