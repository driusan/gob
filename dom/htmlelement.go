package dom

import (
	"fmt"
	"strings"
	"unicode"

	"net/url"

	"golang.org/x/mobile/event/key"
	"golang.org/x/net/html"
)

type Element struct {
	*html.Node

	Parent, FirstChild, NextSibling *Element

	// for Input elements
	Value string
}

func (e Element) GetTextContent() string {
	if e.Node.Type == html.TextNode {
		return e.Node.Data
	}
	var ret string
	for c := e.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			ret += c.Data
		case html.ElementNode:
			ret += c.GetTextContent()
		}
	}
	return ret
}

func (e Element) GetAttribute(attr string) string {
	for _, attrField := range e.Attr {
		if attrField.Key == attr {
			return attrField.Val
		}
	}
	return ""

}

func (e Element) IsLink() bool {
	if e.Type != html.ElementNode {
		return false
	}
	return e.Data == "a" && e.GetAttribute("href") != ""
}

func (e Element) OnClick() {
}

func (e *Element) SendKey(evt key.Event) {
	if e.Type != html.ElementNode || e.Data != "input" {
		// This isn't necessarily a bad thing, but currently the only
		// way for this function to get called is for input elements,
		// so panic to make sure we don't forget to update the code if
		// we change it
		panic("Sent key to non-input (not implemented)")
	}
	switch evt.Code {
	case key.CodeDeleteBackspace:
		if len(e.Value) > 0 {
			// FIXME: This will only work with ASCII, but the
			// strings package doesn't have any kind of "LastRune"
			// function
			e.Value = e.Value[:len(e.Value)-1]
		}
		return
	case key.CodeReturnEnter:
		form := e.getContainingForm()
		if form != nil {
			form.Submit()
		}
		return
	}

	// FIXME: This should be more intelligent about control
	// characters
	if evt.Rune >= 0 && unicode.IsPrint(evt.Rune) {
		e.Value += fmt.Sprintf("%c", evt.Rune)
	}
}

func (e *Element) getContainingForm() *Element {
	for cur := e; cur != nil; cur = cur.Parent {
		if cur.Type == html.ElementNode && strings.ToLower(cur.Data) == "form" {
			return cur
		}
	}
	return nil
}

func (e *Element) Submit() error {
	if e.Type != html.ElementNode || strings.ToLower(e.Data) != "form" {
		return fmt.Errorf("Not a form")
	}

	values := make(url.Values)
	e.addInputValues(values)
	fmt.Printf("%v", values)
	panic("Submit not implemented")
}

func (e *Element) addInputValues(values url.Values) {
	for cur := e.FirstChild; cur != nil; cur = cur.NextSibling {
		if cur.Type == html.ElementNode && strings.ToLower(cur.Data) == "input" {
			if name := cur.GetAttribute("name"); name != "" {
				values.Add(name, cur.Value)
			}
		} else if cur.Type == html.ElementNode {
			cur.addInputValues(values)
		}
	}
}
