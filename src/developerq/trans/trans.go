package trans

import (
	"golang.org/x/net/html"
//	"github.com/Unknwon/goconfig"
//	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os/exec"
//	"fmt"
	"bytes"
	"strconv"
//	"strings"
	"io/ioutil"
//	"net/http"
)


func TranslateText(input string) string {
	cmd := exec.Command("rm", "input")
	cmd.Run()

	binput := []byte(input)
	err := ioutil.WriteFile("input", binput, 0644)
	if err != nil {
		return ""
	}
	cmd = exec.Command("rm", "output")
	cmd.Run()

	cmd = exec.Command("php", "resource/developerq/t.php")
	cmd.Run()
	b, err := ioutil.ReadFile("output")
	output := string(b)
	//unescape json
	output, _ = strconv.Unquote(output)
	return output
}

func TranslateHTMLNode(n *html.Node) string {
	output := ""
	//for batch trans
	pending := ""
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		//get html raw
		b := new(bytes.Buffer)
		if err := html.Render(b, c); err != nil {
//			.Logger.Error(err)
		}
		htmlraw := b.String()

		if c.Type == html.ElementNode && c.Data == "p" {
			//add to batch list
			pending = pending + htmlraw
		} else {
			//when node is not <p> trans it, then add htmlraw
			if pending != "" {
				//d.Logger.Info("translation called")
				o := TranslateText(pending)
				if o == "" {
					output = output + pending
				} else {
					output = output + o
				}
			} else {
				output = output + pending
			}
			//add <code> <pre> and emtpy the pending
			output = output + htmlraw
			pending = ""
		}
	}
	return output
}
