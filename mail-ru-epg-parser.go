package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/gookit/slog"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	configDefaultName = "mail-ru-parser-config.yaml"
	basePath          = "https://tv.mail.ru/ajax/channel"
)

var (
	httpClient = &http.Client{Timeout: 20 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
)

func runMailRuParser() {
	parserConfig, err := loadConfiguration()
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Error("failed to load configuration file")
		return
	}
	if !parserConfig.Enabled {
		log.Info("MailRu parser module is disabled, exiting...")
		return
	}
	mskLoc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Error("failed to get Moscow location")
		return
	}

	ticker := time.NewTicker(time.Duration(parserConfig.Interval) * time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			for {
				select {
				case <-ticker.C:
					parseEpg(parserConfig, mskLoc)
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}
	}()
	parseEpg(parserConfig, mskLoc)
}

func parseEpg(parserConfig *MailRuEpgParseConfig, mskLoc *time.Location) {
	for _, channel := range parserConfig.Channels {
		var items []InputDataItem
		for i := 0; i < parserConfig.Depth; i++ {
			currentDate := time.Now()
			currentDate = currentDate.AddDate(0, 0, i)
			itemsForDate, err := getEpgForDate(currentDate, mskLoc, channel, parserConfig)
			if err != nil {
				log.WithFields(log.M{"error": err.Error()}).Errorf("failed to get epg r '%s' channel", channel.Name)
				continue
			}
			items = append(items, itemsForDate...)
			time.Sleep(3 * time.Second)
		}
		EpgList.Store(strconv.Itoa(channel.XmltvId), EpgListEntry{Name: channel.Name, Items: items, Expire: time.Now().Add(expire)})
	}
}

func getEpgForDate(currentDate time.Time, mskLoc *time.Location, channel MailRuEpgParseConfigEntry, parserConfig *MailRuEpgParseConfig) ([]InputDataItem, error) {
	log.Debugf("getting epg for channel '%s' and date '%s'", channel.Name, currentDate.Format("2006-01-02"))
	res, err := httpClient.Do(getReq(parserConfig, channel, currentDate))
	if err != nil {
		return nil, err
	}
	var epgReqBody MailRuGetEpgResp
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&epgReqBody)
	if err != nil {
		return nil, errors.New("failed to parse epg-req-body")
	}
	if strings.ToLower(epgReqBody.Status) != "ok" {
		return nil, errors.New(fmt.Sprintf("status of epg-req is not ok -> '%s'", epgReqBody.Status))
	}

	var items []InputDataItem
	for sIndex, s := range epgReqBody.Schedule {
		epgReqBody.Schedule[sIndex].Event.Concat = append(s.Event.Past, s.Event.Current...)
		addDay := false
		for eIndex, e := range epgReqBody.Schedule[sIndex].Event.Concat {
			parsedTime, err := time.Parse("15:04", e.Start)
			if err != nil {
				log.WithFields(log.M{"error": err.Error()}).Error("failed to parse start-time '%s'", e.Start)
				continue
			}

			epgDate := time.Date(
				currentDate.Year(), currentDate.Month(), currentDate.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, mskLoc)
			if (eIndex > 0 && epgDate.Before(epgReqBody.Schedule[sIndex].Event.Concat[eIndex-1].StartDate)) || addDay {
				epgDate = epgDate.AddDate(0, 0, 1)
				addDay = true
			}
			epgReqBody.Schedule[sIndex].Event.Concat[eIndex].StartDate = epgDate
		}

		for eIndex, e := range epgReqBody.Schedule[sIndex].Event.Concat {
			if item := getItem(channel.XmltvId, epgReqBody.Schedule[sIndex], e, eIndex); item != nil {
				items = append(items, *item)
			}
		}
	}
	return items, nil
}

func getItem(id int, s MailRuGetEpgRespSchedule, e MailRuGetEpgRespEvent, eIndex int) *InputDataItem {
	endTime := s.Event.Concat[eIndex].StartDate
	if eIndex+1 < len(s.Event.Concat) {
		endTime = s.Event.Concat[eIndex+1].StartDate
	}
	item := &InputDataItem{
		Title:   map[string]string{"ru": e.Name},
		Channel: strconv.Itoa(id),
		StartUt: e.StartDate.Unix(),
		StopUt:  endTime.Unix(),
	}
	if e.EpisodeTitle != "" {
		item.Subtitle = map[string]string{"ru": e.EpisodeTitle}
	}
	return item
}

func getReq(parserConfig *MailRuEpgParseConfig, channel MailRuEpgParseConfigEntry, currentDate time.Time) *http.Request {
	url := fmt.Sprintf("%s/?region_id=%d&channel_id=%d&date=%s",
		basePath, parserConfig.Region, channel.MailRuId, currentDate.Format("2006-01-02"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.2 Safari/605.1.15")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json")
	return req
}

func loadConfiguration() (*MailRuEpgParseConfig, error) {
	configPath := ""
	workdir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Error("failed to get work-dir")
	} else {
		configPath = fmt.Sprintf("%s/%s", workdir, configDefaultName)
	}

	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	parserConfig := &MailRuEpgParseConfig{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	d := yaml.NewDecoder(file)
	if err := d.Decode(&parserConfig); err != nil {
		return nil, err
	}
	return parserConfig, nil
}
