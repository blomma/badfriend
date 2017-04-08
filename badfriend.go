package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

var (
	ip          = flag.String("ip", "", "ip to lookup")
	redisServer = flag.String("redis", "", "redis server to hook into")
)

type geoipResponse struct {
	Ip          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionCode  string  `json:"region_code"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	ZipCode     string  `json:"zip_code"`
	TimeZone    string  `json:"time_zone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
	TimeStamp   int64
}

func main() {
	flag.Parse()

	addr := fmt.Sprintf("http://freegeoip.net/json/%s", *ip)
	httpResp, err := http.Get(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer httpResp.Body.Close()

	jsonData, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		log.Fatal(err)
	}

	resp := geoipResponse{}
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		log.Fatal(err)
	}

	client, err := redis.Dial("tcp", *redisServer)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	now := time.Now()
	secs := now.Unix()

	resp.TimeStamp = secs

	request, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	key := fmt.Sprintf("badfriend:%d", resp.TimeStamp)

	client.PipeAppend("MULTI")
	client.PipeAppend("SET", key, string(request))
	client.PipeAppend("SADD", "badfriends", resp.TimeStamp)
	client.PipeAppend("EXEC")

	if err := client.PipeResp().Err; err != nil {
		log.Fatal(err)
	}
	if err := client.PipeResp().Err; err != nil {
		log.Fatal(err)
	}
	if err := client.PipeResp().Err; err != nil {
		log.Fatal(err)
	}
	if err := client.PipeResp().Err; err != nil {
		log.Fatal(err)
	}
}
