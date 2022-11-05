package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func epgIncome(c *gin.Context) {
	body := &InputData{}
	if err := c.BindJSON(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	for _, channel := range body.Channels {
		var items []InputDataItems
		for _, item := range body.Items {
			if item.Channel != channel.Id {
				continue
			}
			items = append(items, item)
		}
		EpgList.Store(channel.Id, EpgListEntry{Name: channel.Name, Items: items, Expire: time.Now().Add(expire)})
	}
	log.Debugf("processed '%d' entries", len(body.Items))
}

//func sendEpgFile(c *gin.Context) {
//	c.Header("Content-Type", "application/xml")
//	c.File(epgXmlFilePath)
//}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
