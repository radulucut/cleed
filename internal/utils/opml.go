package utils

import (
	"encoding/xml"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Title string `xml:"title,omitempty"`
}

type Body struct {
	Oultines []*Outline `xml:"outline"`
}

type Outline struct {
	Outlines []*Outline `xml:"outline"`
	XMLURL   string     `xml:"xmlUrl,attr,omitempty"`
}
