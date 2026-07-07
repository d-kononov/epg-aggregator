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
	Items  []InputDataItem
	Expire time.Time
}

// InputData - Input EPG json data
type InputData struct {
	Channels []InputDataChannel `json:"channels"`
	Items    []InputDataItem    `json:"items"`
}
type InputDataChannel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type InputDataItem struct {
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

// -- MAIL GRABBER --

type MailGrabberConfig struct {
	Interval               int                             `yaml:"parser-interval"`
	ImapAddress            string                          `yaml:"imap-address"`
	Login                  string                          `yaml:"login"`
	Password               string                          `yaml:"password"`
	DeleteMessageAfterRead bool                            `yaml:"delete-message-after-read"`
	Channels               []MailGrabberChannelConfigEntry `yaml:"email-channels"`
}
type MailGrabberChannelConfigEntry struct {
	Name                 string `yaml:"name"`
	XmltvId              int    `yaml:"xmltv-id"`
	EmailSubjectContains string `yaml:"email-subject-contains"`
	EmailFromContains    string `yaml:"from-contains"`
}

// -- MAIL-RU --

type MailRuEpgParseConfig struct {
	Enabled  bool                        `yaml:"enabled"`
	Interval int                         `yaml:"parser-interval"`
	Region   int                         `yaml:"region"`
	Depth    int                         `yaml:"parsing-depth"`
	Channels []MailRuEpgParseConfigEntry `yaml:"mail-ru-channels"`
}
type MailRuEpgParseConfigEntry struct {
	Name     string `yaml:"name"`
	XmltvId  int    `yaml:"xmltv-id"`
	MailRuId int    `yaml:"mail-ru-id"`
}
type MailRuGetEpgResp struct {
	Data MailRuGetEpgRespSchedule `json:"data"`
}
type MailRuGetEpgRespSchedule struct {
	Channel struct {
		Id        int    `json:"id"`
		Name      string `json:"name"`
		Url       string `json:"url"`
		LiveUrl   string `json:"live_url"`
		HasLive   bool   `json:"has_live"`
		IsCentral bool   `json:"is_central"`
	} `json:"channel"`
	Events []MailRuGetEpgRespEvent `json:"events"`
}
type MailRuGetEpgRespEvent struct {
	Name    string `json:"name"`
	StartTs int64  `json:"start_ts"`
	StopTs  int64  `json:"stop_ts"`
}
