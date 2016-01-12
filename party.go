package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nonsenz/bbcScraper"
	"log"
	"net/http"
	"strings"
	"io/ioutil"
	"io"
)

const showBucket string = "shows"
const unprocessedBroadcastBucket string = "unprocessed_broadcasts"
const broadcastBucket string = "broadcasts"
const trackBucket string = "tracks"

var storer bbcScraper.Storer

func main() {

	storer = bbcScraper.Storer(bbcScraper.NewBoltStorer("tracks.db"))
	defer storer.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/update", update).Methods("POST")
	router.HandleFunc("/show", addShow).Methods("POST")
	router.HandleFunc("/shows", getShows).Methods("GET")
	router.HandleFunc("/track", getRandomTrack).Methods("GET")
	log.Fatal(http.ListenAndServe(":8081", router))

}

func update(response http.ResponseWriter, request *http.Request) {

	shows := [...]string{
		"b01fm4ss", // gilles
		"b00rwkjd", // freakier zone
		"b0072l4x", // freakzone
	}

	for _, showId := range shows {
		fmt.Printf("processing:%s\n", showId)
		processAllPages := storer.Get(showId, showBucket) == ""
		broadcastIds := bbcScraper.BroadcastIds(showId, processAllPages)

		for _, bid := range broadcastIds {
			//			fmt.Printf("processing:%s,%s\n", showId, bid)
			if storer.Get(bid, broadcastBucket) == "" {
				tracks := bbcScraper.BroadcastTracks(bid)

				for _, track := range tracks {
					trackString := strings.ToLower(track.Artist + ": " + track.Title)
					if storer.Get(trackString, trackBucket) == "" {
						if err := storer.Put(trackString, "done", trackBucket); err != nil {
							log.Fatal(err)
						}
						fmt.Printf("put:%s,%s,%s;\n", showId, bid, trackString)
					} else {
						fmt.Printf("skipped:%s,%s,%s;\n", showId, bid, trackString)
					}
				}

				// we processed all tracks in this broadcast so we persist it as done
				if err := storer.Put(bid, "done", broadcastBucket); err != nil {
					log.Fatal(err)
				}
			}
		}

		if processAllPages {
			// we processed all show broadcasts.
			// now we persist the showId the so we dont have to do an initial full import in the future
			if err := storer.Put(showId, "done", showBucket); err != nil {
				log.Fatal(err)
			}
		}
		fmt.Printf("done with show %s\n", showId)
	}

	if err := json.NewEncoder(response).Encode("done"); err != nil {
		panic(err)
	}

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")
	response.WriteHeader(http.StatusOK)
}

func addShow(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	type Show struct {
		Id  string	`json:"id"`
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

	if err := storer.Put(show.Id, "", showBucket); err != nil {
		log.Fatal(err)
	}

	response.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(response).Encode("added"); err != nil {
		panic(err)
	}
}

func getShows(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	shows := storer.All(showBucket)

	if err := json.NewEncoder(response).Encode(shows); err != nil {
		panic(err)
	}
}

func getRandomTrack(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	track := storer.Random(trackBucket)

	if err := json.NewEncoder(response).Encode(track); err != nil {
		panic(err)
	}
}
