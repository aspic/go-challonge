package main

import (
    "log"
    "fmt"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "net/url"
)

const (
    tournaments = "tournaments"
    version = "v1"
)

var c ChallongeClient

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
    Id string `json:"id"`
    Url string `json:"url"`
    FullUrl string `json:"full_challonge_url"`

    client *ChallongeClient
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
func (c *ChallongeClient) CreateTournament(name string, subUrl string, open bool, tournamentType string) *Tournament {
    v := url.Values{}
    v.Add("tournament[name]", name)
    v.Add("tournament[url]", subUrl)
    v.Add("tournament[open_signup]", "false")
    // v.Add("tournament[tournament_type]", tournamentType)
    url := c.buildUrl("tournaments", v)
    response := &APIResponse{}
    c.doPost(url, response)
    return response.Tournament.withClient(c)
}

/** returns tournament with the specified id */
func (c *ChallongeClient) GetTournament(id string) *Tournament {
    v := url.Values{}
    v.Add("include_participants", "1")
    url := c.buildUrl("tournaments/" + id, v)
    response := &APIResponse{}
    c.doGet(url, response)
    return response.Tournament.withClient(c)
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
func (t *Tournament) Remove(id string) *APIResponse {
    url := t.client.buildUrl("tournaments/" + t.Url + "/participants/" + id, nil)
    response := &APIResponse{}
    c.doDelete(url, response)
    return response
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
    defer r.Body.Close()
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Fatal("unable to parse response", err)
    }
    bodyString := string(body)
    log.Print("Got response: ", bodyString)
    json.Unmarshal(body, v)
}

func main() {
    c := &ChallongeClient{user: "viking1", version: version, key: "k0PG6IxBQhH8tkpTDlxaUKHLMHRfMy1oycloZgTW"}
    c.Print()
    response := c.CreateTournament("foobar3", "foobar3", true, "oo").Add("basr")
    if response.Errors != nil {
        log.Print(response.Errors)
    }
}
