package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
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
	var ip string
	flag.StringVar(&ip, "ip", "", "IP to lookup")

	var redisHost string
	flag.StringVar(&redisHost, "redis", "", "Redis host to connect to")

	flag.Parse()

	if response, err := http.Get("http://freegeoip.net/json/" + ip); err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()

	if jsonData, err := ioutil.ReadAll(response.Body); err != nil {
		fmt.Println(err)
		return
	}

	res := Response{}
	json.Unmarshal(jsonData, &res)

	if c, err := redis.Dial("tcp", redisHost); err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	now := time.Now()
	secs := now.Unix()

	res.TimeStamp = secs

	if result1, err := json.Marshal(res); err != nil {
		fmt.Println(err)
		return
	}

	key := "badfriend:" + strconv.FormatInt(res.TimeStamp, 10)

	c.Send("MULTI")
	c.Send("SET", key, string(result1))
	c.Send("SADD", "barfriends", res.TimeStamp)
	if _, err := c.Do("EXEC"); err != nil {
		fmt.Println(err)
	}
}
