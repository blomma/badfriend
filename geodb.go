package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	geoip2 "github.com/oschwald/geoip2-golang"
)

type geoDB struct {
	sync.RWMutex
	reader      *geoip2.Reader
	stopChan    chan struct{}
	stoppedChan chan struct{}
}

func newGeoDB() (*geoDB, func()) {
	g := &geoDB{
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}

	go fetchGeo(g)

	stop := func() {
		close(g.stopChan)
		<-g.stoppedChan
	}

	return g, stop
}

func fetchGeo(g *geoDB) {
	defer close(g.stoppedChan)
	for {
		g.Lock()
		if g.reader != nil {
			g.reader.Close()
		}

		// Download new
		log.Println("Downloading GeoLite file")
		err := downloadFile("GeoLite2-City.mmdb.gz", "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Unziping GeoLite file")
		err = unpackGzipFile("GeoLite2-City.mmdb.gz", "GeoLite2-City.mmdb")
		if err != nil {
			log.Fatal(err)
		}

		reader, err := geoip2.Open("GeoLite2-City.mmdb")
		if err != nil {
			log.Fatal(err)
		}

		g.reader = reader
		g.Unlock()

		timerChan := make(chan struct{})
		go func() {
			<-time.After(360 * time.Hour)
			close(timerChan)
		}()

		select {
		case <-g.stopChan:
			return
		case <-timerChan:
			break
		}
	}
}

func unpackGzipFile(gzFilePath, dstFilePath string) error {
	gzFile, err := os.Open(gzFilePath)
	if err != nil {
		return fmt.Errorf("Failed to open file %s for unpack: %s", gzFilePath, err)
	}

	dstFile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return fmt.Errorf("Failed to create destination file %s for unpack: %s", dstFilePath, err)
	}

	ioReader, ioWriter := io.Pipe()

	go func() { // goroutine leak is possible here
		gzReader, _ := gzip.NewReader(gzFile)
		// it is important to close the writer or reading from the other end of the
		// pipe or io.copy() will never finish
		defer func() {
			gzFile.Close()
			gzReader.Close()
			ioWriter.Close()
		}()

		io.Copy(ioWriter, gzReader)
	}()

	_, err = io.Copy(dstFile, ioReader)
	if err != nil {
		return err // goroutine leak is possible here
	}
	ioReader.Close()
	dstFile.Close()

	return nil
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
