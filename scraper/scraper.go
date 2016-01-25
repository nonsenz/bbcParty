package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
	"strings"
)

type Track struct {
	Artist, Title, Label string
}

func BroadcastIds(showId string, all bool) []string {

	continueUntilPage := 1
	showUrl := "https://www.bbc.co.uk/programmes/" + showId + "/episodes/guide?page="
	var broadcastIds []string

	showDoc, err := goquery.NewDocument(showUrl + strconv.Itoa(continueUntilPage))
	if err != nil {
		log.Fatal(err)
	}

	if all {
		maxPage, _ := strconv.Atoi(showDoc.Find(".pagination__page--last a").Text())
		if maxPage > 0 {
			continueUntilPage = maxPage
		}
	}

	for pageCount := 1; pageCount <= continueUntilPage; pageCount++ {
		if pageCount > 1 {
			showDoc, err = goquery.NewDocument(showUrl + strconv.Itoa(pageCount))
			if err != nil {
				log.Fatal(err)
			}
		}

		broadcastIds = append(broadcastIds, showDoc.Find(".programme__titles a").Map(func(i int, s *goquery.Selection) string {
			broadcastLink, _ := s.Attr("href")
			return strings.Split(broadcastLink, "/")[2]
		})...)
	}

	return broadcastIds
}

func BroadcastTracks(broadcastId string) []Track {
	broadcastUrl := "https://www.bbc.co.uk/programmes/" + broadcastId + "/segments.inc"

	broadcastDoc, err := goquery.NewDocument(broadcastUrl)
	if err != nil {
		log.Fatal(err)
	}

	var tracks []Track

	broadcastDoc.Find(".segment__track").Each(func(i int, s *goquery.Selection) {
		tracks = append(tracks, Track{s.Find(".artist").Text(), s.Find("p span").Text(), s.Find("ul span").Text()})
	})

	return tracks
}
