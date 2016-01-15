package tuber

import (
	"log"
	"net/http"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/googleapi/transport"
	"fmt"
)

type Tuber struct {
	ApiKey string
}

func (t *Tuber) FirstHit(searchTerm string) string {

	fmt.Println(searchTerm)

	client := &http.Client{
		Transport: &transport.APIKey{Key: t.ApiKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	call := service.Search.List("id").Q(searchTerm).MaxResults(10)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
	}

	fmt.Println(len(response.Items))

	// Iterate through each item and add it to the correct list.
	for k, item := range response.Items {
		fmt.Println(k, item.Id.Kind)
		if item.Id.Kind == "youtube#video" {
			return item.Id.VideoId
		}
	}

	return ""
}
