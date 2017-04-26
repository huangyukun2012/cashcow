package utils

import (

	"fmt"
	"github.com/yanyiwu/gojieba"
	"math/rand"
	//	_ "github.com/go-sql-driver/mysql"
	//sql "database/sql"
	//	"io/ioutil"
	"logging"
	//	"regexp"
	//	"encoding/json"
	//	"time"
	//	"github.com/garyburd/redigo/redis"
	//"github.com/Unknwon/goconfig"
	//	"strconv"
	//"bytes"
	//	"os"
	//	"bufio"
	//	"io"
	s "strings"
	//m "model"
	//"utils"
	"strconv"
	"time"
	"sync"

)

//global
var LISTMAX int
var PAGEMAX int
var NAVMAX int
var RANDMAX int

var CAT_INT_STR map[int]string

var CAT_STR_INT map[string]int

var CAT_INT_STRCN map[int]string

//var Logger *logging.Logger
/*
func CheckErr(err error) {
	if err != nil {
		Logger.Error("Error...", err.Error())
	}
}
*/

func IntToDateStr(d int64) string {
	tm := time.Unix(d, 0)
	ts := tm.Format("2006-01-02 15:04:05")
	return ts
}

func IntToStr(v int64) string {
	s := strconv.FormatInt(v, 10)
	return s
}


var pmap = map[rune]rune{
	'，':',',
	'。':'.',
	'：':':',
	'“':'"',
	'”':'"',
	'（':'(',
	'）':')',
	'！':'!',
	'《':'<',
	'》':'>',
	'／':'/',
	'？':'?',
	'；':';',
	'【':'[',
	'】':']',
	'「':'[',
	'」':']',
	'—':'_',
	'、':'\\',
	'～':'~',
	'·':'`',
}


func ReplaceCNPunctuation(input string) string {
	if input == "" {
		return ""
	}
	rs := []rune(input)
	for k, r := range rs {
		if v, ok := pmap[r]; ok {
			rs[k] = v
		}
	}
	return string(rs)
}


func InitCateMap() {
	CAT_INT_STR = map[int]string{}
	CAT_INT_STRCN = map[int]string{}
	CAT_STR_INT = map[string]int{}

	CAT_INT_STR[0] = "all"
	CAT_INT_STR[1] = "video"
	CAT_INT_STR[2] = "torrent"
	CAT_INT_STR[3] = "soft"
	CAT_INT_STR[4] = "doc"
	CAT_INT_STR[5] = "music"
	CAT_INT_STR[6] = "picture"
	CAT_INT_STR[7] = "other"

	CAT_INT_STRCN[0] = "全部"
	CAT_INT_STRCN[1] = "视频"
	CAT_INT_STRCN[2] = "种子"
	CAT_INT_STRCN[3] = "软件"
	CAT_INT_STRCN[4] = "文档"
	CAT_INT_STRCN[5] = "音乐"
	CAT_INT_STRCN[6] = "图片"
	CAT_INT_STRCN[7] = "其他"

	CAT_STR_INT["all"] = 0
	CAT_STR_INT["video"] = 1
	CAT_STR_INT["torrent"] = 2
	CAT_STR_INT["soft"] = 3
	CAT_STR_INT["doc"] = 4
	CAT_STR_INT["music"] = 5
	CAT_STR_INT["picture"] = 6
	CAT_STR_INT["other"] = 7
}

func SplitNames(fn string) []string {
	ss := s.Split(fn, "#$#$")
	return ss
}

func SplitItem(fn string) []string {
	ss := s.Split(fn, "@+@+")
	return ss
}

func SizeToStr(size int64) string {

	if size <= 1024 {
		return IntToStr(size) + " B"
	}

	size = size / 1024
	if size <= 1024 {
		return IntToStr(size) + " KB"
	}

	s := float64(size) / 1024
	if s <= 1024 {
		return fmt.Sprintf("%.2f", s) + " MB"
	}

	s = float64(s) / 1024
	return fmt.Sprintf("%.2f", s) + " GB"
}

func GetCategoryFromName(name string) int {
	name = s.ToLower(name)

	if s.Contains(name, "mp4") || s.Contains(name, "mkv") || s.Contains(name, "avi") || s.Contains(name, "wmv") || s.Contains(name, "mpg") ||s.Contains(name, "mpeg")  ||s.Contains(name, "rmvb") ||s.Contains(name, "mov") ||s.Contains(name, "flv"){
		return 1
	}

	if s.Contains(name, "torrent") {
		return 2
	}

	if s.Contains(name, "apk") || s.Contains(name, "exe") || s.Contains(name, "dmg") {
		return 3
	}

	if s.Contains(name, "doc") || s.Contains(name, "ppt") || s.Contains(name, "xls") || s.Contains(name, "txt") || s.Contains(name, "pdf") {
		return 4
	}


	if s.Contains(name, "wav") || s.Contains(name, "mp3") || s.Contains(name, "m4a") || s.Contains(name, "acc") {
		return 5
	}

	if s.Contains(name, "bmp") || s.Contains(name, "jpg") || s.Contains(name, "jpeg") || s.Contains(name, "png") || s.Contains(name, "gif") {
		return 6
	}
	return 7
}


//for seo
var Jb * gojieba.Jieba
func InitJieba() {
	Jb = gojieba.NewJieba()
}


func ConvertNumber(n int64) string {
	var str string
	if n < 10000 {
		str = fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		str = fmt.Sprintf("%.0f", float64(n)/float64(1000)) + "k"
	} else {
		str = fmt.Sprintf("%.1f", float64(n)/float64(1000000)) + "m"
	}
	return str
}

var BilisouStat = make(map[string]int)
var BilisouCount = 0
var LBilisou sync.Mutex

var DeveloperqStat = make(map[string]int)
var DeveloperqCount = 0
var LDeveloperq sync.Mutex

func UpdateDeveloperqStat(ipaddrport string, log *logging.Logger) {
	ss := s.Split(ipaddrport, ":")
	if len(ss) != 2 {
		return
	}
	ipaddr := ss[0]

	LDeveloperq.Lock()
	_, ok := DeveloperqStat[ipaddr]
	if ok {
		DeveloperqStat[ipaddr] = DeveloperqStat[ipaddr] + 1
	} else {
		DeveloperqStat[ipaddr] = 1
	}
	LDeveloperq.Unlock()

	DeveloperqCount = DeveloperqCount + 1
	if DeveloperqCount % 1000 == 0 {
		DumpMap(DeveloperqStat, DeveloperqCount, log)
	}
}


func UpdateBilisouStat(ipaddrport string , log *logging.Logger) {
	ss := s.Split(ipaddrport, ":")
	if len(ss) != 2 {
		return
	}
	ipaddr := ss[0]

	_, ok := BilisouStat[ipaddr]
	LBilisou.Lock()
	if ok {
		BilisouStat[ipaddr] = BilisouStat[ipaddr] + 1
	} else {
		BilisouStat[ipaddr] = 1
	}
	LBilisou.Unlock()

	BilisouCount = BilisouCount + 1
	if BilisouCount % 1000 == 0 {
		DumpMap(BilisouStat, BilisouCount, log)
	}
}

func DumpMap(stat map[string]int, count int, log *logging.Logger) {
	log.Info("Total Request Count = %d", count)
	log.Info("IP Count = %d", len(stat))
	for key, value := range stat {
		log.Info("[%s] = %d", key, value)
	}
}

func GetRandoms(start int, end int, size int) []int {
	res := []int{}

	ranges := end - start

	if ranges <= 0 {
		ranges = 1
	}

	for i:=0; i<size; i++ {
		rand.Seed(time.Now().UnixNano())
		res = append(res, rand.Intn(ranges) + start)
	}
	return res
}
