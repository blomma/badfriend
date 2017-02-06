package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	g_ip           = flag.String("ip", "", "ip to lookup")
	g_redis_server = flag.String("redis", "", "redis server to hook into")
)

type Response struct {
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

	addr := fmt.Sprintf("http://freegeoip.net/json/%s", *g_ip)
	httpResponse, err := http.Get(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer httpResponse.Body.Close()

	jsonData, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Fatal(err)
	}

	response := Response{}
	if err := json.Unmarshal(jsonData, &response); err != nil {
		log.Fatal(err)
	}

	connection, err := redis.Dial("tcp", *g_redis_server)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	now := time.Now()
	secs := now.Unix()

	response.TimeStamp = secs

	request, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}

	key := fmt.Sprintf("badfriend:%d", response.TimeStamp)

	if err := connection.Send("MULTI"); err != nil {
		log.Fatal(err)
	}
	if err := connection.Send("SET", key, string(request)); err != nil {
		log.Fatal(err)
	}
	if err := connection.Send("SADD", "badfriends", response.TimeStamp); err != nil {
		log.Fatal(err)
	}
	if _, err := connection.Do("EXEC"); err != nil {
		log.Fatal(err)
	}
}
