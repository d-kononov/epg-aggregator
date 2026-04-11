package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	log "github.com/gookit/slog"
	"gopkg.in/yaml.v3"
)

const (
	configGrabberDefaultName = "mail-grabber-config.yaml"
)

func runMailGrabber() {
	config, err := loadGrabberConfiguration()
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Error("failed to load configuration file")
		return
	}
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			for {
				select {
				case <-ticker.C:
					fetchMail(config)
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}
	}()
	fetchMail(config)
}

func fetchMail(config *MailGrabberConfig) {
	c, err := client.DialTLS(config.ImapAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	if err := c.Login(config.Login, config.Password); err != nil {
		log.Fatal(err)
	}

	_, err = c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	criteria := imap.NewSearchCriteria()
	//criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, 10)
	go func() {
		err = c.Fetch(seqset, []imap.FetchItem{imap.FetchRFC822}, messages)
		if err != nil {
			log.Fatal(err)
		}
	}()

	for msg := range messages {
		processMessage(config, msg)
		if config.DeleteMessageAfterRead {
			deleteMessage(c, msg)
		}
	}
}

func processMessage(config *MailGrabberConfig, msg *imap.Message) {
	if msg == nil {
		return
	}

	body := msg.GetBody(&imap.BodySectionName{})
	if body == nil {
		return
	}

	entity, err := message.Read(body)
	if err != nil {
		log.Println("read message:", err)
		return
	}
	xmltvDateLayout := "20060102150405 -0700"

	// Wrap with mail reader (gives nice helpers)
	mr := mail.NewReader(entity)
	header := mr.Header
	from, _ := header.AddressList("From")
	subject, _ := header.Subject()
	for _, channel := range config.Channels {
		var containsSubject, containsEmailFrom bool
		fromEmailAdd := ""
		if channel.EmailSubjectContains != "" {
			if strings.Contains(strings.ToLower(subject), strings.ToLower(channel.EmailSubjectContains)) {
				containsSubject = true
			}
		}
		if channel.EmailFromContains != "" && len(from) > 0 {
			if strings.Contains(strings.ToLower(from[0].Address), strings.ToLower(channel.EmailFromContains)) {
				containsEmailFrom = true
			}
			fromEmailAdd = from[0].Address
		}
		if containsSubject || containsEmailFrom {
			log.Infof("Processing message '%s' from '%s'", channel.EmailSubjectContains, fromEmailAdd)
			xmltvFiles, err := walkEntity(entity)
			if err != nil {
				log.WithFields(log.M{"error": err.Error()}).Errorf("failed to parse xmltv from message #%d", msg.SeqNum)
				continue
			}

			var items []InputDataItem
			for _, xmltvFile := range xmltvFiles {
				for _, programme := range xmltvFile.Programmes {
					startAt, err := time.Parse(xmltvDateLayout, programme.Start)
					if err != nil {
						log.WithFields(log.M{"error": err.Error()}).Errorf("failed to parse start-time '%s'", programme.Start)
						continue
					}
					stopAt, err := time.Parse(xmltvDateLayout, programme.Stop)
					if err != nil {
						log.WithFields(log.M{"error": err.Error()}).Errorf("failed to parse stop-time '%s'", programme.Start)
						continue
					}

					item := &InputDataItem{
						Title:   map[string]string{"ru": programme.Title.Value},
						Channel: strconv.Itoa(channel.XmltvId),
						StartUt: startAt.Unix(),
						StopUt:  stopAt.Unix(),
					}
					items = append(items, *item)
				}
				log.Infof("Processed '%d' entries", len(items))
				EpgList.Store(strconv.Itoa(channel.XmltvId), EpgListEntry{Name: channel.Name, Items: items, Expire: time.Now().Add(expire)})
			}
		}
	}
}

func walkEntity(e *message.Entity) ([]XmlTv, error) {
	mediaType, params, err := e.Header.ContentType()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		var files []XmlTv
		mr := e.MultipartReader()
		if mr == nil {
			return nil, nil
		}

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			xmltv, err := walkEntity(part)
			if err != nil {
				continue
			}
			files = append(files, xmltv...)
		}
		return files, nil
	}

	// Handle single part (possible attachment)
	disposition, params, _ := e.Header.ContentDisposition()
	filename := params["filename"]

	if disposition == "attachment" && filename != "" {
		if strings.HasSuffix(strings.ToLower(filename), ".xml") {
			buf := new(bytes.Buffer)
			_, _ = io.Copy(buf, e.Body)
			file, err := parseXMLTV(buf.Bytes())
			if err != nil {
				return nil, err
			}
			return []XmlTv{*file}, nil
		}
	}
	return nil, errors.New("no xmltv file found")
}

func parseXMLTV(data []byte) (*XmlTv, error) {
	var tv XmlTv
	if err := xml.Unmarshal(data, &tv); err != nil {
		return nil, err
	}
	return &tv, nil
}

func deleteMessage(c *client.Client, msg *imap.Message) {
	log.Infof("Deleting message #%d", msg.SeqNum)
	seqset := new(imap.SeqSet)
	seqset.AddNum(msg.SeqNum)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}

	err := c.Store(seqset, item, flags, nil)
	if err != nil {
		log.Println("delete error:", err)
	}
}

func loadGrabberConfiguration() (*MailGrabberConfig, error) {
	configPath := ""
	workdir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Error("failed to get work-dir")
	} else {
		configPath = fmt.Sprintf("%s/%s", workdir, configGrabberDefaultName)
	}

	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}
	config := &MailGrabberConfig{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
