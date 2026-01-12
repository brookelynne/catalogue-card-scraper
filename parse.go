package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

const (
	htmlTag = "html"
	bodyTag = "body"

	idAttr    = "id"
	classAttr = "class"

	mainContainerID = "main-container"
	marcViewID      = "marc_view"

	fieldClass        = "field"
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

func parse(responseBody []byte) error {
	document, err := html.Parse(bytes.NewReader(responseBody))
	if err != nil {
		return err
	}

	htmlNode := getFirstChildOfType(document, htmlTag)
	if htmlNode == nil {
		return fmt.Errorf("%w (failed to find <%s>)", traversalError, htmlTag)
	}
	body := getFirstChildOfType(htmlNode, bodyTag)
	if body == nil {
		return fmt.Errorf("%w (failed to find <%s>)", traversalError, bodyTag)
	}
	mainContainer := getFirstChildWithAttr(body, idAttr, mainContainerID)
	if mainContainer == nil {
		return fmt.Errorf("%w (failed to find %s)", traversalError, mainContainerID)
	}

	marcView := getDescendantWithAttr(mainContainer, idAttr, marcViewID)
	if marcView == nil {
		return fmt.Errorf("%w (failed to find %s)", traversalError, marcViewID)
	}

	for field := range marcView.ChildNodes() {
		if !fieldIsWanted(field) {
			continue
		}

		if v := getFirstChildWithAttr(field, classAttr, subfieldsClass); v != nil {
			fmt.Println(getSubfieldsAsString(v))
		}
	}

	return nil
}

func getFirstChildOfType(n *html.Node, tagName string) *html.Node {
	for node := range n.ChildNodes() {
		if node.Type == html.ElementNode && node.Data == tagName {
			return node
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

func getDescendantWithAttr(n *html.Node, attrName, attrValue string) *html.Node {
	if n == nil {
		return nil
	}
	// Check if current node matches
	for _, attr := range n.Attr {
		if attr.Key == attrName && attr.Val == attrValue {
			return n
		}
	}
	// Recursively check children
	for child := range n.ChildNodes() {
		if result := getDescendantWithAttr(child, attrName, attrValue); result != nil {
			return result
		}
	}
	return nil
}

func fieldIsWanted(n *html.Node) bool {
	tagIndicator := getFirstChildWithAttr(n, classAttr, tagIndicatorClass)
	if tagIndicator == nil {
		return false
	}
	tag := getFirstChildWithAttr(tagIndicator, classAttr, tagClass)
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
	n = getFirstChildWithAttr(n, classAttr, subfieldsClass)
	for child := range n.ChildNodes() {
		for _, a := range child.Attr {
			if a.Key == classAttr && a.Val == subCodeClass && child.FirstChild.Data == "5|" {
				return true
			}
		}
	}
	return false
}

func getSubfieldsAsString(n *html.Node) string {
	if n == nil {
		return ""
	}
	rtn := ""
	for n = range n.ChildNodes() {
		if n.Type == html.ElementNode {
			if n.Data == "span" && n.FirstChild.Data == "5|" {
				return rtn // We don't want any text after the 5| delimiter
			}
			continue
		}
		if text := strings.TrimSpace(n.Data); text != "" && text != "UNAUTHORIZED" {
			if rtn != "" {
				rtn += " "
			}
			rtn += text
		}
	}
	return rtn
}
