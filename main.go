package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/gookit/slog"
	"os"
	"strconv"
	"time"
)

var epgXmlFilePath = ""
var expire = 24 * time.Hour
var stalkerEpg = true

func main() {
	setOutputFilePath()

	interval := 60 * time.Second
	if os.Getenv("INTERVAL") != "" {
		intervalInt, err := strconv.Atoi(os.Getenv("INTERVAL"))
		if err == nil {
			interval = time.Duration(intervalInt) * time.Second
		}
	}
	if os.Getenv("EXPIRE") != "" {
		expireInt, err := strconv.Atoi(os.Getenv("EXPIRE"))
		if err == nil {
			interval = time.Duration(expireInt) * time.Hour
		}
	}
	if os.Getenv("STALKER") == "false" {
		stalkerEpg = false
	}

	// update epg.xml file each 60 seconds
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	go func() {
		for {
			for {
				select {
				case <-ticker.C:
					epgFileGenerate()
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}
	}()

	go runMailRuParser()

	healthCheckEndpoint := "/health"
	r := gin.New()
	r.Use(gin.LoggerWithWriter(gin.DefaultWriter, healthCheckEndpoint), gin.Recovery())
	r.GET(healthCheckEndpoint, health)

	r.POST("/", epgIncome)
	r.StaticFile("/epg.xml", epgXmlFilePath)
	r.StaticFile("/epg.xml.gz", fmt.Sprintf("%s.gz", epgXmlFilePath))

	_ = r.Run()
}

func setOutputFilePath() {
	outputFilePath := os.Getenv("OUTPUT_PATH")
	if outputFilePath != "" {
		epgXmlFilePath = outputFilePath
		return
	}
	workDir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.M{"error": err.Error()}).Fatal("failed to get workdir")
	}
	staticDir := fmt.Sprintf("%s/static", workDir)
	_, err = os.Stat(staticDir)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(staticDir, 0755)
	}
	epgXmlFilePath = fmt.Sprintf("%s/epg.xml", staticDir)
}
