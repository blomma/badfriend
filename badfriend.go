package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

// Version is the version number or commit hash
// These variables should be set by the linker when compiling
var (
	Version = "0.0.0"
)

// Options
var (
	flagRedis   = flag.String("redis", "", "redis server to hook into")
	flagVersion = flag.Bool("version", false, "Show the version number and information")
)

// Global variables
var (
	db *geoDB
)

type geoResponse struct {
	IP          string
	City        string
	Continent   string
	Country     string
	Latitude    float64
	Longitude   float64
	CountryCode string
	PostalCode  string
	TimeZone    string
	TimeStamp   int64
}

func badFriendHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	ipFromForm := r.Form.Get("ip")
	ip := net.ParseIP(ipFromForm)
	db.RLock()
	record, err := db.reader.City(ip)
	if err != nil {
		log.Fatal(err)
	}
	db.RUnlock()

	reponse := &geoResponse{
		IP:          ip.String(),
		City:        record.City.Names["en"],
		Continent:   record.Continent.Names["en"],
		Country:     record.Country.Names["en"],
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		CountryCode: record.Country.IsoCode,
		PostalCode:  record.Postal.Code,
		TimeZone:    record.Location.TimeZone,
		TimeStamp:   time.Now().Unix(),
	}

	client, err := redis.Dial("tcp", *flagRedis)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	request, err := json.Marshal(reponse)
	if err != nil {
		log.Fatal(err)
	}

	key := fmt.Sprintf("badfriend:%d", reponse.TimeStamp)

	client.PipeAppend("MULTI")
	client.PipeAppend("SET", key, string(request))
	client.PipeAppend("SADD", "badfriends", reponse.TimeStamp)
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

	// fmt.Printf("City name: %v\n", record.City.Names["en"])
	// fmt.Printf("Continent name: %v\n", record.Continent.Names["en"])
	// fmt.Printf("Country name: %v\n", record.Country.Names["en"])
	// fmt.Printf("Coordinates: %v, %v\n", record.Location.Latitude, record.Location.Longitude)
	// fmt.Printf("Postal: %v \n", record.Postal.Code)
	// fmt.Printf("ISO country code: %v\n", record.Country.IsoCode)
	// fmt.Printf("Time zone: %v\n", record.Location.TimeZone)
}

func logHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		log.Println(fmt.Sprintf("%q", x))
		defer log.Println("<------")
		next.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	geoDB, stopGeoDB := newGeoDB()
	db = geoDB

	mux := http.NewServeMux()
	mux.HandleFunc("/badfriend", badFriendHandler)

	srv := &http.Server{
		Handler:      logHandler(mux),
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stopGeoDB()
		if db.reader != nil {
			db.reader.Close()
		}
		os.Exit(1)
	}()

	log.Fatal(srv.ListenAndServe())
}
