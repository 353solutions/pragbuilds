package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Build struct {
	ID     int
	Name   string
	Failed bool
}

func parseHTML(r io.Reader) ([]Build, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return findBuilds(doc)
}

func findBuilds(doc *html.Node) ([]Build, error) {
	nodes := findNodes(doc, isBuildsTable)
	if len(nodes) != 1 {
		return nil, fmt.Errorf("can't find builds table")
	}

	trs := findNodes(nodes[0], isBuildTr)
	var builds []Build
	for _, tr := range trs {
		b, err := nodeToBuild(tr)
		if err != nil {
			return nil, err
		}
		builds = append(builds, b)
	}

	return builds, nil
}

func isBuildsTable(node *html.Node) bool {
	if !isElem(node, "table") {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "id" && attr.Val == "bookshelf" {
			return true
		}
	}

	return false
}

func isBuildTr(node *html.Node) bool {
	if !isElem(node, "tr") {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, "build_status") {
			return true
		}
	}
	return false
}

func isElem(node *html.Node, typ string) bool {
	if node.Type != html.ElementNode {
		return false
	}
	return node.Data == typ
}

func findNodes(node *html.Node, pred func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	stack := []*html.Node{node}

	for len(stack) != 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if pred(n) {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			stack = append(stack, c)
		}
	}

	return out
}

func nodeToBuild(node *html.Node) (Build, error) {
	var b Build
	id, err := buildID(node)
	if err != nil {
		return Build{}, err
	}
	b.ID = id

	found := false
	for _, attr := range node.Attr {
		if attr.Key != "class" {
			continue
		}
		b.Failed = strings.Contains(attr.Val, "failed")
		found = true
		break
	}

	if !found {
		return Build{}, fmt.Errorf("can't find status in %#v", node)
	}

	nodes := findNodes(node, isH4)
	if len(nodes) != 1 {
		return Build{}, fmt.Errorf("can't find build name in %#v", node)
	}
	text := nodes[0].FirstChild
	if text == nil {
		return Build{}, fmt.Errorf("can't find name in %#v", node)
	}
	b.Name = strings.TrimSpace(text.Data)
	if b.Name == "" {
		return Build{}, fmt.Errorf("empty name in %#v", node)
	}
	return b, nil
}

func isH4(node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "h4"
}

func buildID(node *html.Node) (int, error) {
	for _, attr := range node.Attr {
		if attr.Key != "id" {
			continue
		}
		i := strings.LastIndex(attr.Val, "_")
		if i == -1 {
			return 0, fmt.Errorf("can't find build ID in %q", attr.Val)
		}
		id, err := strconv.Atoi(attr.Val[i+1:])
		if err != nil {
			return 0, fmt.Errorf("bad build id in %q: %s", attr.Val, err)
		}
		return id, nil
	}

	return 0, fmt.Errorf("can't find build id in %#v", node)
}
