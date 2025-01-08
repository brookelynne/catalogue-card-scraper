package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	tcn := getTitleControlNumber()
	resp, err := http.Get(fmt.Sprintf("https://iucat.iu.edu/catalog/%s/librarian_view", tcn))
	if err != nil {
		fmt.Println("Error fetching item:", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("No record found for title control number", tcn)
		os.Exit(1)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("a" + tcn)
	if err = parse(body); err != nil {
		fmt.Println("Error parsing item:", err)
		os.Exit(1)
	}
}

func getTitleControlNumber() string {
	args := os.Args
	if len(args) < 2 {
		fmt.Print(getHelpText())
		os.Exit(1)
	}
	if args[1] == "help" || args[1] == "-h" || args[1] == "--help" {
		fmt.Print(getHelpText())
		os.Exit(0)
	}
	tcn := args[1]
	// title control numbers begin with a lower-case 'a'; we'll strip that for use in the URL
	if tcn[0] == 'a' {
		tcn = tcn[1:]
	}
	return tcn
}

func getHelpText() string {
	return `Missing title control number. Please supply a title control number after the program name, example:
    catalogue-card-scraper.exe a19858379
The "a" prefix on the title control number is optional.
For help or change requests, please email brooke.weaver@gmail.com

             .--.           .---.        .-.
         .---|--|   .-.     | L |  .---. |~|    .--.
      .--|===|  |---|_|--.__| I |--|:::| |~|-==-|==|---.
      |%%|RDA|  |===| |~~|%%| L |--|   |_|~|CATS|  |___|-.
      |  |   |  |===| |==|  | L |  |:::|=| |    |BX|---|=|
      |  |   |  |   |_|__|  | Y |__|   | | |    |  |___| |
      |~~|===|--|===|~|~~|%%|~~~|--|:::|=|~|----|==|---|=|
      ^--^---'--^---^-^--^--^---'--^---^-^-^-==-^--^---^-'

`
	// ASCII art credit: https://www.asciiart.eu/books/books
}

func parse(body []byte) error {
	page, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	// TODO: can you make this a little more reliable/concrete, to ensure we've found body?
	bodyTag := page.FirstChild.NextSibling.LastChild
	if bodyTag == nil || bodyTag.Data != "body" {
		return fmt.Errorf("%w (failed to find body)", traversalError)
	}
	mainContainer := getFirstChildWithAttr(bodyTag, id, mainContainerID)
	if mainContainer == nil {
		return fmt.Errorf("%w (failed to find %s)", traversalError, mainContainerID)
	}

	table := mainContainer.LastChild.PrevSibling // Should be the table; can we make this more concrete?
	rows := getFirstChildWithAttr(table, id, marcViewID)
	if rows == nil {
		return fmt.Errorf("%w (failed to find %s)", traversalError, marcViewID)
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
