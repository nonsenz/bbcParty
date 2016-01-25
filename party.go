package main

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"github.com/nonsenz/bbcParty/scraper"
	"github.com/nonsenz/bbcParty/storer"
	"github.com/nonsenz/bbcParty/tuber"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	showBucket                 string = "shows"
	unprocessedBroadcastBucket string = "unprocessed_broadcasts"
	broadcastBucket            string = "broadcasts"
	trackBucket                string = "tracks"
	trackDelimiter             string = "<:>"
)

type Config struct {
	GoogleApiKey string
	DbFile       string
}

var (
	db      storer.Storer
	config  Config
	tub     tuber.Tuber
	nextHit string
)

func main() {

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	db = storer.NewBoltStorer(config.DbFile)
	defer db.Close()

	tub = tuber.Tuber{config.GoogleApiKey}
	nextHit = getHitHelper()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", getIndex).Methods("GET")
	router.HandleFunc("/update", update).Methods("POST")
	router.HandleFunc("/show", addShow).Methods("POST")
	router.HandleFunc("/shows", getShows).Methods("GET")
	router.HandleFunc("/track", getRandomTrack).Methods("GET")
	router.HandleFunc("/stats", getStats).Methods("GET")
	router.HandleFunc("/hit", getHit).Methods("GET")
	log.Fatal(http.ListenAndServe(":8081", router))
}

func update(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	for _, showId := range db.All(showBucket) {
		log.Printf("processing:%s\n", showId)
		processAllPages := db.Get(showId, showBucket) == ""
		broadcastIds := scraper.BroadcastIds(showId, processAllPages)

		for _, bid := range broadcastIds {
			//			fmt.Printf("processing:%s,%s\n", showId, bid)
			if db.Get(bid, broadcastBucket) == "" {
				tracks := scraper.BroadcastTracks(bid)

				for _, track := range tracks {
					trackString := strings.ToLower(track.Artist + ": " + track.Title)
					if db.Get(trackString, trackBucket) == "" {
						if err := db.Put(trackString, "done", trackBucket); err != nil {
							log.Fatal(err)
						}
						fmt.Printf("put:%s,%s,%s;\n", showId, bid, trackString)
					} else {
						fmt.Printf("skipped:%s,%s,%s;\n", showId, bid, trackString)
					}
				}

				// we processed all tracks in this broadcast so we persist it as done
				if err := db.Put(bid, "done", broadcastBucket); err != nil {
					log.Fatal(err)
				}
			}
		}

		if processAllPages {
			// we processed all show broadcasts.
			// now we persist the showId the so we dont have to do an initial full import in the future
			if err := db.Put(showId, "done", showBucket); err != nil {
				log.Fatal(err)
			}
		}
		fmt.Printf("done with show %s\n", showId)
	}

	if err := json.NewEncoder(response).Encode("done"); err != nil {
		panic(err)
	}
}

func addShow(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	type Show struct {
		Id string `json:"id"`
	}

	var show Show

	body, err := ioutil.ReadAll(io.LimitReader(request.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := request.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &show); err != nil {
		response.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(response).Encode(err); err != nil {
			panic(err)
		}
	}

	if err := db.Put(show.Id, "", showBucket); err != nil {
		log.Fatal(err)
	}

	response.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(response).Encode("added"); err != nil {
		panic(err)
	}
}

func getShows(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	shows := db.All(showBucket)

	if err := json.NewEncoder(response).Encode(shows); err != nil {
		panic(err)
	}
}

func getRandomTrack(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	track := db.Random(trackBucket)

	if err := json.NewEncoder(response).Encode(track); err != nil {
		panic(err)
	}
}

func getStats(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(response).Encode("number of tracks: " + strconv.Itoa(len(db.All(trackBucket)))); err != nil {
		panic(err)
	}

}

func getIndex(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "text/html; charset=UTF-8")

	t, _ := template.ParseFiles("templates/index.html")
	js, _ := ioutil.ReadFile("js/main.js")

	indexData := struct {
		Js template.JS
	}{
		template.JS(js),
	}

	t.Execute(response, indexData)
}

func getHit(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	hit := nextHit

	// get hit for next call async
	// this way request does not hang until youtube api call is finished
	go func() {
		nextHit = getHitHelper()
	}()

	hitData := struct {
		Id string `json:"id"`
	}{hit}

	if err := json.NewEncoder(response).Encode(hitData); err != nil {
		panic(err)
	}
}

func getHitHelper() (hit string) {
	for hit == "" {
		hit = tub.FirstHit(db.Random(trackBucket))
	}
	return hit
}
