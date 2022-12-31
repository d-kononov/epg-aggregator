package main

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	log "github.com/gookit/slog"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	Header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<!DOCTYPE tv SYSTEM "xmltv.dtd">` + "\n"
)

func epgFileGenerate() {
	output := &XmlTv{GeneratorName: "dk-astra-epg-aggregator"}
	EpgList.Range(func(key, value any) bool {
		entry := value.(EpgListEntry)
		if entry.Expire.Before(time.Now()) {
			EpgList.Delete(key)
		} else {
			output.Channels = append(output.Channels, Channel{Id: key.(string), DisplayName: entry.Name})
			for _, item := range entry.Items {
				if len(item.Title) < 1 {
					continue
				}
				titleLangs, titleValues := JoinMapValues(item.Title)
				descLangs, descValues := JoinMapValues(item.Subtitle)

				p := Programme{
					Channel:  item.Channel,
					Start:    time.Unix(item.StartUt, 0).Format("20060102150405 +0000"),
					Stop:     time.Unix(item.StopUt, 0).Format("20060102150405 +0000"),
					Title:    &Entry{Lang: titleLangs, Value: titleValues},
					Desc:     &Entry{Lang: descLangs, Value: descValues},
					Category: &Entry{Lang: "en", Value: strings.Join(item.Category, "/")},
				}
				if !stalkerEpg {
					p.Desc = nil
					p.SubTitle = &Entry{Lang: descLangs, Value: descValues}
				}
				output.Programmes = append(output.Programmes, p)
			}
		}
		return true
	})
	sort.Slice(output.Channels[:], func(i, j int) bool {
		return output.Channels[i].Id < output.Channels[j].Id
	})
	sort.Slice(output.Programmes[:], func(i, j int) bool {
		return output.Programmes[i].Start < output.Programmes[j].Start
	})
	out, err := xml.MarshalIndent(output, " ", "  ")
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Error("failed to convert object to xml")
		return
	}
	epgBytes := []byte(Header + string(out))
	err = os.WriteFile(epgXmlFilePath, epgBytes, 0644)
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Errorf("failed to write xml to file '%s'", epgXmlFilePath)
		return
	}

	// gzipped xml
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write(epgBytes)
	_ = w.Close()
	err = os.WriteFile(fmt.Sprintf("%s.gz", epgXmlFilePath), b.Bytes(), 0644)
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Errorf("failed to write compressed xml to file '%s.gz'", epgXmlFilePath)
		return
	}
}

func JoinMapValues(m map[string]string) (string, string) {
	keys := make([]string, 0, len(m))
	values := make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	return strings.Join(keys, "/"), strings.Join(values, "/")
}
