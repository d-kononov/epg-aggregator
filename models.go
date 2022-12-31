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
	Status        string                     `json:"status"`
	CurrentTs     int                        `json:"current_ts"`
	CurrentOffset int                        `json:"current_offset"`
	Schedule      []MailRuGetEpgRespSchedule `json:"schedule"`
}
type MailRuGetEpgRespSchedule struct {
	Channel struct {
		Name      string      `json:"name"`
		PicBig    interface{} `json:"pic_big"`
		PicSmall  interface{} `json:"pic_small"`
		IsPromo   int         `json:"is_promo"`
		PicUrl    interface{} `json:"pic_url"`
		PicUrl128 interface{} `json:"pic_url_128"`
		Url       string      `json:"url"`
		Slug      string      `json:"slug"`
		Id        string      `json:"id"`
		PicUrl64  interface{} `json:"pic_url_64"`
	} `json:"channel"`
	Event struct {
		Current []MailRuGetEpgRespEvent `json:"current"`
		Past    []MailRuGetEpgRespEvent `json:"past"`
		Concat  []MailRuGetEpgRespEvent
	} `json:"event"`
}
type MailRuGetEpgRespEvent struct {
	ChannelId    string `json:"channel_id"`
	Name         string `json:"name"`
	CategoryId   int    `json:"category_id"`
	EpisodeTitle string `json:"episode_title"`
	Url          string `json:"url"`
	Id           string `json:"id"`
	Start        string `json:"start"`
	EpisodeNum   int    `json:"episode_num"`
	StartDate    time.Time
}
