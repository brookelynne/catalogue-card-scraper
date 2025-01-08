package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const (
	id    = "id"
	class = "class"

	mainContainerID   = "main-container"
	marcViewID        = "marc_view"
	tagIndicatorClass = "tag_ind"
	tagClass          = "tag"
	subfieldsClass    = "subfields"
	subCodeClass      = "sub_code"
)

var (
	tagsWanted = []string{
		"090",
		"100",
		"240",
		"245",
		"260",
		"264",
		"300",
		// 500 only if it ends with `5|`
		"590",
		"591",
		// Any 7XX that ends with `5|`
	}

	traversalError = errors.New("unable to properly traverse page")
)

func main() {
	controlNumber := getControlNumber()
	uri := fmt.Sprintf("https://iucat.iu.edu/catalog/%s/librarian_view", controlNumber)

	fmt.Println("a" + controlNumber)

	resp, err := http.Get(uri)
	if err != nil {
		log.Fatalf("Error fetching item: %s\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if err = parse(body); err != nil {
		log.Fatalf("Error parsing item: %s\n", err)
	}
}

func getControlNumber() string {
	args := os.Args
	if len(args) < 2 {
		// TODO: TEST MODE ONLY
		//return "20972946"
		//return "19858379"
		log.Fatal("Missing control number; please supply a control number after the program, example:\n" +
			"./catalogue-card-scraper 19858379")

	}
	return args[1]
}

func parse(body []byte) error {
	page, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	// TODO: can you make this a little more reliable/concrete, to ensure we've found body?
	bodyTag := page.FirstChild.NextSibling.LastChild
	if bodyTag == nil || bodyTag.Data != "body" {
		return fmt.Errorf("%e (failed to find body)", traversalError)
	}
	mainContainer := getFirstChildWithAttr(bodyTag, id, mainContainerID)
	if mainContainer == nil {
		return fmt.Errorf("%e (failed to find %s)", traversalError, mainContainerID)
	}

	table := mainContainer.LastChild.PrevSibling // Should be the table; can we make this more concrete?
	rows := getFirstChildWithAttr(table, id, marcViewID)
	if rows == nil {
		return fmt.Errorf("%e (failed to find %s)", traversalError, marcViewID)
	}

	for row := range rows.ChildNodes() {
		if !rowIsWanted(row) {
			continue
		}

		if v := getFirstChildWithAttr(row, class, subfieldsClass); v != nil {
			fmt.Println(getSubfieldsAsString(v))
		}
	}

	return nil
}

func getFirstChildWithAttr(n *html.Node, attrName, attrValue string) *html.Node {
	if n == nil {
		return nil
	}
	for n = range n.ChildNodes() {
		for _, attr := range n.Attr {
			if attr.Key == attrName && attr.Val == attrValue {
				return n
			}
		}
	}
	return nil
}

func rowIsWanted(n *html.Node) bool {
	tagIndicator := getFirstChildWithAttr(n, class, tagIndicatorClass)
	if tagIndicator == nil {
		return false
	}
	tag := getFirstChildWithAttr(tagIndicator, class, tagClass)
	if tag == nil {
		return false
	}
	tagNumber := strings.TrimSpace(tag.FirstChild.Data)
	for _, wanted := range tagsWanted {
		if tagNumber == wanted {
			return true
		}
	}
	// 500 or any 700 with `5|`
	if (tagNumber == "500" || tagNumber[0] == '7') && subfieldsContains5Pipe(n) {
		return true
	}
	return false
}

func subfieldsContains5Pipe(n *html.Node) bool {
	n = getFirstChildWithAttr(n, class, subfieldsClass)
	for child := range n.ChildNodes() {
		for _, a := range child.Attr {
			if a.Key == class && a.Val == subCodeClass && child.FirstChild.Data == "5|" {
				return true
			}
		}
	}
	return false
}

func getSubfieldsAsString(n *html.Node) (subfields string) {
	if n == nil {
		return ""
	}
	for n = range n.ChildNodes() {
		if n.Data == "span" {
			continue
		}
		if text := strings.TrimSpace(n.Data); text != "" && text != "UNAUTHORIZED" {
			if subfields != "" {
				subfields += " "
			}
			subfields += text
		}
	}
	return
}
