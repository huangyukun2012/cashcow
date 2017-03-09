package model

import (
	"time"
	"errors"
	"database/sql"
	"math/rand"
	t "html/template"
	"strings"
	u "developerq/utils"
	h "github.com/jaytaylor/html2text"
	"logging"
	"github.com/russross/blackfriday"

)

var Logger *logging.Logger

type Article struct {
	ExtID            int64   `json:"exit_id"`
	Title            string  `json:"title"`
	Question  string  `json:"question"`
	Answer    string  `json:"answer"`
	Tags             string  `json:"tags"`
	OriginalURL      string  `json:"orignal_url"`
	Source           string  `json:"source"`
	ViewCount        int64   `json:"view_count"`
	VoteCount        int64   `json:"vote_count"`
	TitleCN           string `json:"title_cn"`
	QuestionCN string `json:"question_cn"`
	AnswerCN   string `json:"answer_cn"`
	UK              int64    `json:"uk"`
	UpdateTime      string    `json:"update_time"`
	QuestionRaw     string    `json:"question_raw"`
	QuestionCNRaw   string    `json:"question_cn_raw"`
	AnswerRaw       string    `json:"answer_raw"`
	AnswerCNRaw     string    `json:"answer_cn_raw"`
	TitleCNRaw      string    `json:"title_cn_raw"`
	TitleRaw        string    `json:"title_raw"`
	ScanTime        int64    `json:"scan_time"`

	URL             string   `json:"url"`
	Flag            int       `json:"flag"`

	TagsList        []string
	HQuestionCN     t.HTML
	HAnswerCN       t.HTML
	HTitleCN        t.HTML
	Abstract        string
	HQuestionCNRaw  t.HTML

	HTitleHighlight  t.HTML
	HQuestionHighlight  t.HTML
	SeoKeywords     []string
}

var MaxArticle int
var MinArticle int

//running
func (article * Article) FillHtml()*Article {


	//replace < /
	article.QuestionCN = strings.Replace(article.QuestionCN, "</ ", "</", -1)
	article.TitleCN = strings.Replace(article.TitleCN, "</ ", "</", -1)
	article.AnswerCN = strings.Replace(article.AnswerCN, "</ ", "</", -1)

	article.HTitleCN = t.HTML(blackfriday.MarkdownCommon([]byte(article.TitleCN)))
	article.HQuestionCN = t.HTML(blackfriday.MarkdownCommon([]byte(article.QuestionCN)))
	article.HAnswerCN = t.HTML(blackfriday.MarkdownCommon([]byte(article.AnswerCN)))

	//for search
//	article.HTitleCNRaw = t.HTML(article.TitleCNRaw)
//	article.HQuestionCNRaw = t.HTML(article.QuestionCNRaw)


//	Logger.Info("HTitleCNRaw = ", article.HTitleCNRaw, "Title = ", article.TitleCNRaw)



	tags := strings.Split(article.Tags, ",")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			article.TagsList = append(article.TagsList, tag)
		}
	}

	article.SeoKeywords = article.TagsList
	keywords := u.Jb.Cut(article.TitleCNRaw, true)
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			article.SeoKeywords = append(article.SeoKeywords, keyword)
		}
	}

	keywords = u.Jb.Cut(article.TitleRaw, true)
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			article.SeoKeywords = append(article.SeoKeywords, keyword)
		}
	}
	//Logger.Info("keyword = ", article.SeoKeywords)
	return article
}


type Keyword struct {
	Keyword string `json:"keyword"`
	Count   int64  `json:"count"`
}

//crawling
func (article *Article)FillAll() {
	//uk
	t := time.Now()
	rand.Seed(t.UnixNano())
	r := int64(rand.Intn(19850720))
	article.UK = t.Unix() + r

	//update time
	article.UpdateTime = t.Format("2006-01-02 15:04:05")
	article.ScanTime = int64(t.UnixNano())

	//replace cn prunctuation
	article.TitleCN = u.ReplaceCNPunctuation(article.TitleCN)
	article.QuestionCN = u.ReplaceCNPunctuation(article.QuestionCN)
	article.AnswerCN = u.ReplaceCNPunctuation(article.AnswerCN)

	//replace < /
	article.QuestionCN = strings.Replace(article.QuestionCN, "</ ", "</", -1)
	article.TitleCN = strings.Replace(article.TitleCN, "</ ", "</", -1)
	article.AnswerCN = strings.Replace(article.AnswerCN, "</ ", "</", -1)

	//create raw
	article.QuestionRaw, _ = h.FromString(article.Question)
	article.QuestionCNRaw, _ = h.FromString(article.QuestionCN)
	article.AnswerRaw, _ = h.FromString(article.Answer)
	article.AnswerCNRaw, _ = h.FromString(article.AnswerCN)
	article.TitleRaw, _ = h.FromString(article.Title)

	article.TitleCNRaw, _ = h.FromString(article.TitleCN)

}

func (article *Article)Save(db *sql.DB) error {
	if article.TitleCN == "" || article.AnswerCN == "" ||article.QuestionCN == "" || article.Title == "" ||article.Question == "" {
		return errors.New("Empty article " + article.URL)
	} else {
		ext_id := -1
		rows, err := db.Query("select ext_id from article where ext_id = ?", article.ExtID)
		if err == nil {
			for rows.Next() {
				rows.Scan(&ext_id)
			}
		}

		if ext_id == -1 {
			Logger.Info("Insert article ext_id = %d", article.ExtID)
			_, err := db.Exec("INSERT into article (uk,update_time, title, question, answer, tags, url,  view_count, like_count, source, flag, title_cn, question_cn, answer_cn, vote_count, question_raw, question_cn_raw, answer_raw, answer_cn_raw,title_raw, title_cn_raw, scan_time, ext_id) values(?, ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", article.UK, article.UpdateTime, article.Title, article.Question, article.Answer, article.Tags, article.URL, article.ViewCount, 1, article.Source, 0, article.TitleCN, article.QuestionCN, article.AnswerCN, article.VoteCount, article.QuestionRaw, article.QuestionCNRaw, article.AnswerRaw, article.AnswerCNRaw, article.TitleRaw, article.TitleCNRaw, article.ScanTime, article.ExtID)
			return err
		} else {
			Logger.Info("update article ext_id = %d", article.ExtID)
			_, err := db.Exec("update article set uk= ?,update_time = ?, title = ?, question = ?, answer = ?, tags = ?, url = ?,  view_count = ?, like_count = ?, source = ?, flag = ?, title_cn = ?, question_cn = ?, answer_cn = ?, vote_count = ?, question_raw = ?, question_cn_raw = ?, answer_raw = ?, answer_cn_raw = ?, title_raw = ?, title_cn_raw = ?, scan_time = ? where ext_id = ?", article.UK, article.UpdateTime, article.Title, article.Question, article.Answer, article.Tags, article.URL, article.ViewCount, 1, article.Source, 0, article.TitleCN, article.QuestionCN, article.AnswerCN, article.VoteCount, article.QuestionRaw, article.QuestionCNRaw, article.AnswerRaw, article.AnswerCNRaw, article.TitleRaw, article.TitleCNRaw, article.ScanTime, article.ExtID)
			return err
		}
	}

}

func GetArticleMaxMinID(db *sql.DB) (int, int) {
	var max int
	var min int
	sql := "select max(id), min(id) from article"
	rows, err := db.Query(sql)
	u.CheckErr(err)
	for rows.Next() {
		err = rows.Scan(&max, &min)
	}
	return max, min
}

func ViewArticle(db *sql.DB, uk int64) {
	count := 1
	rows, _ := db.Query("select view_count from article where uk = ?", uk)
	for rows.Next() {
		rows.Scan(&count)
		count = count + 1;
	}
	stmt, _ := db.Prepare("update article set view_count = ?  where uk = ?")
	stmt.Exec(count, count)
	stmt.Close()
}
