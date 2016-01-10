package main

import (
	"fmt"
	"github.com/nonsenz/bbcScraper"
	"log"
	"strings"
	//	"github.com/gorilla/mux"
)

const showBucket string = "shows"
const unprocessedBroadcastBucket string = "unprocessed_broadcasts"
const broadcastBucket string = "broadcasts"
const trackBucket string = "tracks"

func main() {

	s := bbcScraper.Storer(bbcScraper.NewBoltStorer("tracks.db"))
	defer s.Close()

	//	router := mux.NewRouter().StrictSlash(true)
	//	router.HandleFunc("/", Index)
	//	log.Fatal(http.ListenAndServe(":8080", router))

	shows := [...]string{
		"b01fm4ss", // gilles
		//		"b00rwkjd", // freakier zone
		//		"b0072l4x",	// frakzone
	}

	for _, showId := range shows {
		processAllPages := s.Get(showId, showBucket) == ""
		processAllPages = false
		fmt.Printf("processall: %b", s.Get(showId, showBucket))
		broadcastIds := bbcScraper.BroadcastIds(showId, processAllPages)

		for _, bid := range broadcastIds {
			if s.Get(bid, broadcastBucket) == "" {
				tracks := bbcScraper.BroadcastTracks(bid)

				for _, track := range tracks {
					trackString := strings.ToLower(track.Artist + ": " + track.Title)
					if s.Get(trackString, trackBucket) == "" {
						if err := s.Put(trackString, "done", trackBucket); err != nil {
							log.Fatal(err)
						}
					}
				}

				// we processed all tracks in this broadcast so we persist it as done
				if err := s.Put(bid, "done", broadcastBucket); err != nil {
					log.Fatal(err)
				}
			}
		}

		if processAllPages {
			// we processed all show broadcasts.
			// now we persist the showId the so we dont have to do an initial full import in the future
			if err := s.Put(showId, "done", showBucket); err != nil {
				log.Fatal(err)
			}
		}
	}

	fmt.Println("done")
}