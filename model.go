package parser

import (
    "fmt"
    "time"
)

type Result struct {
    Players     map[int]*Player
    Tournaments []*Tournament
}

type Player struct {
    Id        int
    License   int
    Firstname string
    Lastname  string
}

type Tournament struct {
    Id           int
    Name         string
    Competitions []*Competition
}

type Competition struct {
    Id      int
    Name    string
    Type    string
    Date    time.Time
    Teams   map[int]*Team
    Matches []*Match
    Ranking []*Rank
}

type Rank struct {
    Position int

    TeamId int
    Team   *Team `json:"-"`
}

type Team struct {
    Id int

    PlayerId1 int
    Player1   *Player `json:"-"`

    PlayerId2 int
    Player2   *Player `json:"-"`
}

type Match struct {
    Id int

    TeamId1 int
    Team1   *Team `json:"-"`

    TeamId2 int
    Team2   *Team `json:"-"`

    Time time.Time

    Games []*Game
}

type Game struct {
    ScoreTeam1 int
    ScoreTeam2 int
}

func (p *Player) String() string {
    if p.License != 0 {
        return fmt.Sprintf("L:%d", p.License)
    } else {
        return fmt.Sprintf("%s %s", p.Firstname, p.Lastname)
    }
}

func (t *Team) String() string {
    if t.Player2 != nil {
        return fmt.Sprintf("%s/%s", t.Player1, t.Player2)
    } else {
        return t.Player1.String()
    }
}
