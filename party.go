package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"log"
	"strings"
	"time"
	"github.com/nonsenz/bbcScraper"
)

const showBucket string = "shows"
const unprocessedBroadcastBucket string = "unprocessed_broadcasts"
const broadcastBucket string = "broadcasts"
const trackBucket string = "tracks"

func main() {

	shows := [...]string{
		"b01fm4ss", // gilles
//		"b00rwkjd", // freakier zone
//		"b0072l4x",	// frakzone
	}

//	shows := [...]string{"b01fm4ss"}

	db := initDb()
	defer db.Close()

	for _, showId := range shows {
		if err := db.Update(func(tx *bolt.Tx) error {
			showBucket := tx.Bucket([]byte(showBucket))
			bb := tx.Bucket([]byte(broadcastBucket))
			tb := tx.Bucket([]byte(trackBucket))

			// if showId is not in bucket we need an initial import of all pages
			processAllPages := showBucket.Get([]byte(showId)) == nil
			processAllPages = false
			broadcastIds := bbcScraper.BroadcastIds(showId, processAllPages)
			fmt.Println(broadcastIds)

			for _, bid := range broadcastIds {
				if bb.Get([]byte(bid)) == nil {
					tracks := bbcScraper.BroadcastTracks(bid)

					for _, track := range tracks {
						if tb.Get([]byte(track)) == nil {
							fmt.Printf("T")
							if err := tb.Put([]byte(track), []byte("")); err != nil {
								log.Fatal(err)
							}
						}
					}

					// we processed all tracks in this broadcast so we persist it as done
					fmt.Printf("B")
					if err := bb.Put([]byte(bid), []byte("")); err != nil {
						log.Fatal(err)
					}
				}
			}

			if processAllPages {
				// we processed all show broadcasts.
				// now we persist the showId the so we dont have to do an initial full import in the future
				fmt.Printf("S")
				if err := showBucket.Put([]byte(showId), []byte("")); err != nil {
					log.Fatal(err)
				}
			}

			return nil
		}); err != nil {
			log.Fatal(err)
		}
	}


//	ids := bbcScraper.BroadcastIds("b01fm4ss", true)
//	tracks := bbcScraper.BroadcastTracks(ids[2])
//	fmt.Printf("%s\n", ids)

//	Scrape("b01fm4ss")	// gilles
//	Scrape("b00rwkjd") // freakier zone
//	Scrape("b0072l4x") // freak zone
}

func Scrape(showId string) {

	db := initDb()
	defer db.Close()

	showUrl := "http://www.bbc.co.uk/programmes/" + showId + "/episodes/guide"

	showDoc, err := goquery.NewDocument(showUrl)
	if err != nil {
		log.Fatal(err)
	}

	showDoc.Find(".programme__titles a").Each(func(i int, s *goquery.Selection) {
		broadcastLink, _ := s.Attr("href")
		broadcastId := strings.Split(broadcastLink, "/")[2]
		persistBroadcastIdToProcess(broadcastId, db)
	})

	if err := db.Update(func(tx *bolt.Tx) error {

		ubb := tx.Bucket([]byte(unprocessedBroadcastBucket))
		c := ubb.Cursor()

		for broadcastId, _ := c.First(); broadcastId != nil; broadcastId, _ = c.Next() {
			broadcastUrl := "http://www.bbc.co.uk/programmes/" + string(broadcastId) + "/segments.inc"
			fmt.Printf("processing broadcast: %s\n", broadcastId)

			broadcastDoc, err := goquery.NewDocument(broadcastUrl)
			if err != nil {
				log.Fatal(err)
			}

			broadcastDoc.Find(".segment__track").Each(func(i int, s *goquery.Selection) {
				track := strings.ToLower(s.Find(".artist").Text() + " " + s.Find("p span").Text())

				tb := tx.Bucket([]byte(trackBucket))
				trackExists := tb.Get([]byte(track)) != nil

				if ! trackExists {
					fmt.Printf("creating new track: %s\n", track)
					if err := tb.Put([]byte(track), []byte("")); err != nil {
						log.Fatal(err)
					}
				}

				bb := tx.Bucket([]byte(broadcastBucket))
				if err := bb.Put([]byte(broadcastId), []byte("")); err != nil {
					log.Fatal(err)
				}

				ubb.Delete(broadcastId)
			})

		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("done\n")


}

func persistBroadcastIdToProcess(broadcastId string, db *bolt.DB) {
	if err := db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(broadcastBucket))
		broadcastExists := bb.Get([]byte(broadcastId)) != nil

		if ! broadcastExists {
			ubb := tx.Bucket([]byte(unprocessedBroadcastBucket))
			broadcastExists = ubb.Get([]byte(broadcastId)) != nil

			if ! broadcastExists {
				fmt.Printf("creating broadcast %s in db\n", broadcastId)
				if err := ubb.Put([]byte(broadcastId), []byte("")); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func initDb() *bolt.DB {
	db, err := bolt.Open("tracks.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte(broadcastBucket))
		_, err = tx.CreateBucketIfNotExists([]byte(unprocessedBroadcastBucket))
		_, err = tx.CreateBucketIfNotExists([]byte(trackBucket))
		_, err = tx.CreateBucketIfNotExists([]byte(showBucket))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return db
}
