package main

import (
	"encoding/xml"
	"sync"
	"time"
)

// EpgList Global variable to store EGP list
var EpgList sync.Map //map[string]EpgListEntry

type EpgListEntry struct {
	Name   string
	Items  []InputDataItems
	Expire time.Time
}

// InputData - Input EPG json data
type InputData struct {
	Channels []InputDataChannel `json:"channels"`
	Items    []InputDataItems   `json:"items"`
}
type InputDataChannel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type InputDataItems struct {
	Title   map[string]string `json:"title"`
	Channel string            `json:"channel"`
	//Desc    map[string]string `json:"desc"`
	EventId  int64             `json:"event_id"`
	StartUt  int64             `json:"start_ut"`
	Subtitle map[string]string `json:"subtitle"`
	StopUt   int64             `json:"stop_ut"`
	Category []string          `json:"category"`
}

// XmlTv - Output EPG XML object
type XmlTv struct {
	XMLName       xml.Name    `xml:"tv"`
	GeneratorName string      `xml:"generator-info-name,attr"`
	Channels      []Channel   `xml:"channel"`
	Programmes    []Programme `xml:"programme"`
}

type Channel struct {
	Id          string `xml:"id,attr"`
	DisplayName string `xml:"display-name"`
}
type Programme struct {
	Channel string `xml:"channel,attr"`
	Start   string `xml:"start,attr"`
	Stop    string `xml:"stop,attr"`

	Title    *Entry `xml:"title"`
	Desc     *Entry `xml:"desc,omitempty"`
	Category *Entry `xml:"category"`
	SubTitle *Entry `xml:"sub-title,omitempty"`
}

type Entry struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}
