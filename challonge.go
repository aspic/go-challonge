package challonge

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
    API_VERSION = "v1"
    tournaments = "tournaments"
    STATE_OPEN = "open"
    STATE_ALL = "all"
)

var debug = false

var c Client

type tournament Tournament

type Client struct {
    baseUrl string
    key string
    version string
    user string
}

type APIResponse struct {
    Tournament Tournament `json:"tournament"`
    Participant Participant `json:"participant"`

    Errors []string `json:"errors"`
}

type Tournament struct {
    Name string `json:"name"`
    Id int `json:"id"`
    Url string `json:"url"`
    FullUrl string `json:"full_challonge_url"`
    Participants []ParticipantItem `json:"participants"`
    Matches []MatchItem `json:"matches"`
    State string `json:"state"`
    ParticipantsCount int `json:"participants_count"`

    client *Client
}

type Participant struct {
    Id int `json:"id"`
    Name string `json:"name"`
    Wins int
    Losses int
}

type Match struct {
    Id int `json:"id"`
    Identifier string `json:"identifier"`
    State string `json:"state"`
    PlayerOneId int `json:"player1_id"`
    PlayerTwoId int `json:"player2_id"`
    WinnerId int `json:"winner_id"`

    PlayerOne *Participant
    PlayerTwo *Participant
    Winner *Participant
}

/** items to flatten json structure */
type TournamentItem struct {
    Tournament Tournament `json:"tournament"`
}

type ParticipantItem struct {
    Participant Participant `json:"participant"`
}

type MatchItem struct {
    Match Match `json:"match"`
}

func (c *Client) Print() {
    log.Print(c.key)
}

func New(user string, key string) *Client {
    return &Client{user: user, version: API_VERSION, key: key}
}

func (c *Client) Debug() {
    debug = true
}

func (c *Client) buildUrl(route string, v url.Values) string {
    url := fmt.Sprintf("https://%s:%s@api.challonge.com/%s/%s.json", c.user, c.key, c.version, route)
    if v != nil {
        url += "?" + v.Encode()
    }

    return url
}

func params(p map[string]string) *url.Values {
    values := url.Values{}
    for k,v := range(p) {
        values.Add(k, v)
    }
    return &values
}

/** creates a new tournament */
func (c *Client) CreateTournament(name string, subUrl string, open bool, tournamentType string) (*Tournament, error) {
    v := *params(map[string]string{
        "tournament[name]": name,
        "tournament[url]": subUrl,
        "tournament[open_signup]": "false",
    })
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
func (c *Client) GetTournament(id string) (*Tournament, error) {
    v := *params(map[string]string{
        "include_participants": "1",
        "include_matches": "1",
    })
    url := c.buildUrl("tournaments/" + id, v)
    response := &APIResponse{}
    c.doGet(url, response)
    if len(response.Errors) > 0 {
        return nil, fmt.Errorf("unable to retrieve tournament: %q", response.Errors[0])
    }
    return response.Tournament.withClient(c), nil
}

func (c *Client) getTournaments(state string) (*[]Tournament, error) {
    v := *params(map[string]string{
        "state": state,
    })
    url := c.buildUrl("tournaments", v)
    items := make([]TournamentItem, 0)
    c.doGet(url, &items)
    if len(items) == 0 {
        return nil, fmt.Errorf("unable to retrieve tournaments")
    }
    tours := make([]Tournament, 0)
    for _, item := range(items) {
        resolved, err := c.GetTournament(item.Tournament.Name)
        if err != nil {
            return nil, fmt.Errorf("unable to resolve tournament: %q", err)
        }
        tours = append(tours, *resolved)
    }
    return &tours, nil
}

/** returns all ended tournaments */
func (c *Client) GetEndedTournaments() (*[]Tournament, error) {
    return c.getTournaments("ended")
}

/** adds participant to tournament */
func (t *Tournament) AddParticipant(name string) (*Participant, error) {
    v := *params(map[string]string{
        "participant[name]": name,
    })
    url := t.client.buildUrl("tournaments/" + t.Url + "/participants", v)
    response := &APIResponse{}
    c.doPost(url, response)
    if len(response.Errors) > 0 {
        return nil, fmt.Errorf("unable to add participant: %q", response.Errors[0])
    }
    t.Participants = append(t.Participants, ParticipantItem{response.Participant})
    return &response.Participant, nil
}

/** removes participant from tournament */
func (t *Tournament) RemoveParticipant(name string) error {
    p := t.GetParticipantByName(name)
    if p.Id == 0 {
        return fmt.Errorf("participant with name %q not found in tournament", name)
    }
    return t.RemoveParticipantById(p.Id)
}

/** removes participant by id */
func (t *Tournament) RemoveParticipantById(id int) error {
    url := t.client.buildUrl("tournaments/" + t.Url + "/participants/" + strconv.Itoa(id), nil)
    response := &APIResponse{}
    c.doDelete(url, response)
    if len(response.Errors) > 0 {
        return fmt.Errorf("unable to delete participant: %q", response.Errors[0])
    }
    return nil
}

/** returns a participant id based on name */
func (t *Tournament) GetParticipantByName(name string) *Participant {
    for _,item := range t.Participants {
        p := item.Participant
        if p.Name == name {
            return &p
        }
    }
    return nil
}
func (t *Tournament) GetParticipant(id int) *Participant {
    for _,item := range t.Participants {
        p := item.Participant
        if p.Id == id {
            return &p
        }
    }
    return nil
}

func (t *Tournament) GetParticipants() []Participant {
    participants := make([]Participant, 0)
    for _,item := range(t.Participants) {
        participants = append(participants, item.Participant)
    }
    return participants
}

/** returns all matches for tournament */
func (t *Tournament) GetMatches() []Match {
    return t.getMatches(STATE_ALL)
}

/** returns all open matches */
func (t *Tournament) GetOpenMatches() []Match {
    return t.getMatches(STATE_OPEN)
}

/** resolves and returns matches for tournament */
func (t *Tournament) getMatches(state string) []Match {
    matches := make([]Match, 0)

    for _,item := range t.Matches {
        m := item.Match
        m.ResolveParticipants(t)
        if state == STATE_ALL {
            matches = append(matches, m)
        } else if m.State == state {
            matches = append(matches, m)
        }
    }
    return matches
}


/** returns match with resolved participants */
func (t *Tournament) GetMatch(id int) *Match {
    for _,item := range t.Matches {
        m := item.Match
        if m.Id == id {
            m.ResolveParticipants(t)
            return &m
        }
    }
    return nil
}

func (m *Match) ResolveParticipants(t *Tournament) {
    m.PlayerOne = t.GetParticipant(m.PlayerOneId)
    m.PlayerTwo = t.GetParticipant(m.PlayerTwoId)
    m.Winner = t.GetParticipant(m.WinnerId)
}


func (t *Tournament) withClient(c *Client) *Tournament {
    t.client = c
    return t
}

func (c *Client) doGet(url string, v interface{}) {
    if debug {
        log.Print("gets resource on url ", url)
    }
    resp, err := http.Get(url)
    if err != nil {
        log.Fatal("unable to get resource ", err)
    }
    handleResponse(resp, v)
}

func (c *Client) doPost(url string, v interface{}) {
    if debug {
        log.Print("posts resource on url ", url)
    }
    resp, err := http.Post(url, "application/json", nil)
    if err != nil {
        log.Fatal("unable to get resource ", err)
    }
    handleResponse(resp, v)
}

func (c *Client) doDelete(url string, v interface{}) {
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
    if debug {
        log.Print("got response ", string(body))
    }

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
