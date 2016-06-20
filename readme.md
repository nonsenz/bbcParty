# bbcParty

small work in progress project just to play around with golang.
tries to play random tracks from given bbc radio shows on youtube by scraping the show playlists.

## examples

### add show
    curl -H "Content-Type: application/json" -X POST -d '{"id":"b0072l4x"}' http://localhost:8081/show

### update
    curl -H "Content-Type: application/json" -X POST http://localhost:8081/update

## todo

- fix api
- use [jsonapi](http://jsonapi.org/) for the json responses
- add tests ;-)