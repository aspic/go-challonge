package main

import (
    "log"
    "fmt"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "net/url"
    "strconv"
)

const (
    tournaments = "tournaments"
    version = "v1"
)

var c ChallongeClient

// Avoids infinite recursion
type tournament Tournament

type ChallongeClient struct {
    baseUrl string
    key string
    version string
    user string
}

type APIResponse struct {
    Tournament Tournament `json:"tournament"`

    Errors []string `json:"errors"`
}

type Tournament struct {
    Name string `json:"name"`
    Id int `json:"id"`
    Url string `json:"url"`
    FullUrl string `json:"full_challonge_url"`
    Participants []Participant `json:"participants"`

    client *ChallongeClient
}

type Participant struct {
    Id int `json:"id"`
    Name string `json:"name"`
}

func (c *ChallongeClient) Print() {
    log.Print(c.key)
}

func (c *ChallongeClient) buildUrl(route string, v url.Values) string {
    url := fmt.Sprintf("https://%s:%s@api.challonge.com/%s/%s.json", c.user, c.key, c.version, route)
    if v != nil {
        url += "?" + v.Encode()
    }

    return url
}

/** creates a new tournament */
func (c *ChallongeClient) CreateTournament(name string, subUrl string, open bool, tournamentType string) (*Tournament, error) {
    v := url.Values{}
    v.Add("tournament[name]", name)
    v.Add("tournament[url]", subUrl)
    v.Add("tournament[open_signup]", "false")
    // v.Add("tournament[tournament_type]", tournamentType)
    url := c.buildUrl("tournaments", v)
    response := &APIResponse{}
    c.doPost(url, response)
    if len(response.Errors) > 0 {
        return nil, fmt.Errorf("unable to create tournament: %q", response.Errors[0])
    }
    return response.Tournament.withClient(c), nil
}

/** returns tournament with the specified id */
func (c *ChallongeClient) GetTournament(id string) (*Tournament, error) {
    v := url.Values{}
    v.Add("include_participants", "1")
    url := c.buildUrl("tournaments/" + id, v)
    response := &APIResponse{}
    c.doGet(url, response)
    log.Print("resp ", response)
    if len(response.Errors) > 0 {
        return nil, fmt.Errorf("unable to retrieve tournament: %q", response.Errors[0])
    }
    return response.Tournament.withClient(c), nil
}

/** adds participant to tournament */
func (t *Tournament) Add(name string) *APIResponse {
    v := url.Values{}
    v.Add("participant[name]", name)

    url := t.client.buildUrl("tournaments/" + t.Url + "/participants", v)
    response := &APIResponse{}
    c.doPost(url, response)
    return response
}

/** removes participant from tournament */
func (t *Tournament) Remove(name string) error {
    url := t.client.buildUrl("tournaments/" + t.Url + "/participants/" + strconv.Itoa(t.GetParticipantId(name)), nil)
    response := &APIResponse{}
    c.doDelete(url, response)
    if len(response.Errors) > 0 {
        return fmt.Errorf("unable to delete participant: %q", response.Errors[0])
    }
    return nil
}

func (t *Tournament) GetParticipantId(name string) int {
    log.Print("Has: ", t.Participants[0])
    for _,p := range t.Participants {
        if p.Name == name {
            return p.Id
        }
    }
    return 0
}

func (t *Tournament) withClient(c *ChallongeClient) *Tournament {
    t.client = c
    return t
}

func (c *ChallongeClient) doGet(url string, v interface{}) {
    log.Print("gets resource on url ", url)
    resp, err := http.Get(url)
    if err != nil {
        log.Fatal("unable to get resource ", err)
    }
    handleResponse(resp, v)
}

func (c *ChallongeClient) doPost(url string, v interface{}) {
    log.Print("posts resource on url ", url)
    resp, err := http.Post(url, "application/json", nil)
    if err != nil {
        log.Fatal("unable to get resource ", err)
    }
    handleResponse(resp, v)
}

func (c *ChallongeClient) doDelete(url string, v interface{}) {
    req, err := http.NewRequest("DELETE", url, nil)
    log.Print("deletes resource on url ", url)
    if err != nil {
        log.Fatal("unable to create delete request")
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Fatal("unable to delete", err)
    }
    handleResponse(resp, v)
}

func handleResponse(r *http.Response, v interface{}) {
    body, err := ioutil.ReadAll(r.Body)

    log.Print("got response ", string(body))

    if err != nil {
        log.Fatal("unable to read response", err)
    }
    json.Unmarshal(body, v)
}

func (t *Tournament) UnmarshalJSON(b []byte) (err error) {
    placeholder := tournament{}
    if err = json.Unmarshal(b, &placeholder); err == nil {
        *t = Tournament(placeholder)
        return
    }
    return
}

func main() {
    c := &ChallongeClient{user: "viking1", version: version, key: "k0PG6IxBQhH8tkpTDlxaUKHLMHRfMy1oycloZgTW"}
    c.Print()
    //tournament, err := c.CreateTournament("foobar4", "foobar4", true, "oo")
    tournament, err := c.GetTournament("foobar4")
    if err != nil {
        log.Fatal("Got error: ", err)
    } else {
        err = tournament.Remove("foo")
        if err != nil {
            log.Fatal("error when deleting ", err)
        }
    }
}
