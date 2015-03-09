# go-challonge

Golang API client for retrieving (and potentially updating) tournaments in  [Challonge](http://challonge.com/).

## Install

    $ go get github.com/aspic/go-challonge

## Usage

### Load and initialize

    package main
    import "github.com/aspic/go-challonge"
    
    fun main() {
        client := challonge.New("challonge-user", "challonge-key")
    }


### Tournaments

Retrieve a tournament, including all matches and participating players:

    t, err := client.GetTournament("tournament-name")
    
    if err != nil {
        // invalid tournament name, unable to reach host etc.
        log.Fatal("unable to retrieve tournament ", err)
    }
    
### Matches

Get a list of all open matches:

    matches := t.GetOpenMatches()
    
Get a specific match:

    match := t.GetMatch(id)
