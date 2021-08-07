package fast

import (
    "encoding/xml"
    "sort"
    "time"

    "github.com/rtsf-ratings/parser"
)

type xmlTime struct {
    time.Time
}

type xmlFast struct {
    XMLName xml.Name `xml:"ffft"`

    Players     []xmlPlayer     `xml:"registeredPlayers>playerInfos"`
    Tournaments []xmlTournament `xml:"tournaments>tournament"`
}

type xmlPlayer struct {
    XMLName xml.Name `xml:"playerInfos"`

    Id      int `xml:"playerId"`
    License int `xml:"noLicense"`

    NotRegister struct {
        Id        int    `xml:"id"`
        Firstname string `xml:"person>firstName"`
        Lastname  string `xml:"person>lastName"`
    } `xml:"player"`
}

type xmlTournament struct {
    XMLName xml.Name `xml:"tournament"`

    Id           int              `xml:"id,attr"`
    Name         string           `xml:"name"`
    Competitions []xmlCompetition `xml:"competition"`
}

type xmlCompetition struct {
    XMLName xml.Name `xml:"competition"`

    Id     int        `xml:"id"`
    Name   string     `xml:"type"`
    Teams  []xmlTeam  `xml:"competitionTeam"`
    Phases []xmlPhase `xml:"phase"`
}

type xmlRank struct {
    XMLName xml.Name `xml:"ranking"`

    Order    int `xml:"rank"`
    Position int `xml:"definitivePhaseOpponentRanking>relativeRank"`
    TeamId   int `xml:"definitivePhaseOpponentRanking>teamId"`
}

type xmlTeam struct {
    XMLName xml.Name `xml:"competitionTeam"`

    Id        int `xml:"id"`
    PlayerId1 int `xml:"team>player1Id"`
    PlayerId2 int `xml:"team>player2Id"`
}

type xmlPhase struct {
    XMLName xml.Name `xml:"phase"`

    Matches []xmlMatch `xml:"teamMatch"`
    Ranking []xmlRank  `xml:"phaseRanking>ranking"`
    Order   int        `xml:"phaseOrder"`
}

type xmlMatch struct {
    XMLName xml.Name `xml:"teamMatch"`

    Id      int     `xml:"id,attr"`
    TeamId1 int     `xml:"team1Id"`
    TeamId2 int     `xml:"team2Id"`
    Order   int     `xml:"matchNumber"`
    Start   xmlTime `xml:"effectiveStart"`

    Games []xmlGame `xml:"game"`
}

type xmlGame struct {
    XMLName xml.Name `xml:"game"`

    ScoreTeam1 int `xml:"scoreTeam1"`
    ScoreTeam2 int `xml:"scoreTeam2"`
    Order      int `xml:"gameNumber"`
}

func (self *xmlTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
    var v string
    d.DecodeElement(&v, &start)

    t, err := time.Parse("02/01/2006 15:04:05", v)
    if err != nil {
        return err
    }

    self.Time = t

    return nil
}

func (fastParser *FastParser) ParseXML(content []byte) (*parser.Result, error) {
    var xmlfast xmlFast

    err := xml.Unmarshal(content, &xmlfast)
    if err != nil {
        return nil, err
    }

    var fast parser.Result

    players := make(map[int]*parser.Player, len(xmlfast.Players))

    for _, xmlplayer := range xmlfast.Players {
        if xmlplayer.Id == 0 && xmlplayer.NotRegister.Id == 0 {
            continue
        }

        player := &parser.Player{
            Id:        xmlplayer.Id,
            License:   xmlplayer.License,
            Firstname: xmlplayer.NotRegister.Firstname,
            Lastname:  xmlplayer.NotRegister.Lastname,
        }

        if xmlplayer.Id != 0 {
            players[xmlplayer.Id] = player
        } else {
            players[xmlplayer.NotRegister.Id] = player
        }
    }

    fast.Players = players

    for _, xmltornament := range xmlfast.Tournaments {
        tournament := &parser.Tournament{
            Id:   xmltornament.Id,
            Name: xmltornament.Name,
        }

        fast.Tournaments = append(fast.Tournaments, tournament)

        for _, xmlcompetition := range xmltornament.Competitions {
            if len(xmlcompetition.Phases) == 0 {
                continue
            }

            competition := &parser.Competition{
                Id:   xmlcompetition.Id,
                Name: xmlcompetition.Name,
            }

            tournament.Competitions = append(tournament.Competitions, competition)

            teams := make(map[int]*parser.Team, len(xmlcompetition.Teams))

            for _, xmlteam := range xmlcompetition.Teams {
                if xmlteam.PlayerId1 == 0 {
                    continue
                }

                team := &parser.Team{}
                team.Id = xmlteam.Id
                team.PlayerId1 = xmlteam.PlayerId1
                team.Player1 = players[xmlteam.PlayerId1]

                if team.Player1 == nil {
                    continue
                }

                if xmlteam.PlayerId2 > 0 {
                    team.PlayerId2 = xmlteam.PlayerId2
                    team.Player2 = players[xmlteam.PlayerId2]
                }

                teams[xmlteam.Id] = team
            }

            competition.Teams = teams

            xmlphases := make([]xmlPhase, len(xmlcompetition.Phases))
            copy(xmlphases, xmlcompetition.Phases)

            sort.SliceStable(xmlphases, func(i, j int) bool {
                return xmlphases[i].Order < xmlphases[j].Order
            })

            for _, xmlphase := range xmlphases {
                xmlmatches := make([]xmlMatch, len(xmlphase.Matches))
                copy(xmlmatches, xmlphase.Matches)

                sort.SliceStable(xmlmatches, func(i, j int) bool {
                    return xmlmatches[i].Start.Time.Before(xmlmatches[j].Start.Time)
                })

                for _, xmlmatch := range xmlmatches {
                    if xmlmatch.TeamId1 == 0 || xmlmatch.TeamId2 == 0 ||
                        len(xmlmatch.Games) == 0 {

                        continue
                    }

                    match := &parser.Match{}
                    match.Id = xmlmatch.Id
                    match.TeamId1 = xmlmatch.TeamId1
                    match.Team1 = teams[xmlmatch.TeamId1]
                    match.TeamId2 = xmlmatch.TeamId2
                    match.Team2 = teams[xmlmatch.TeamId2]
                    match.Time = xmlmatch.Start.Time

                    xmlgames := make([]xmlGame, len(xmlmatch.Games))
                    copy(xmlgames, xmlmatch.Games)

                    sort.SliceStable(xmlgames, func(i, j int) bool {
                        return xmlgames[i].Order < xmlgames[j].Order
                    })

                    for _, xmlgame := range xmlgames {
                        if xmlgame.ScoreTeam1 == 0 && xmlgame.ScoreTeam2 == 0 {
                            continue
                        }

                        if xmlgame.ScoreTeam1 < 0 || xmlgame.ScoreTeam2 < 0 {
                            continue
                        }

                        game := &parser.Game{
                            ScoreTeam1: xmlgame.ScoreTeam1,
                            ScoreTeam2: xmlgame.ScoreTeam2,
                        }

                        match.Games = append(match.Games, game)
                    }

                    if len(match.Games) > 0 {
                        competition.Matches = append(competition.Matches, match)
                    }
                }
            }

            xmlranking := make([]xmlRank, len(xmlphases[len(xmlphases)-1].Ranking))
            copy(xmlranking, xmlphases[len(xmlphases)-1].Ranking)

            sort.SliceStable(xmlranking, func(i, j int) bool {
                return xmlranking[i].Order < xmlranking[j].Order
            })

            for _, xmlrank := range xmlranking {
                if xmlrank.TeamId == 0 {
                    continue
                }

                team := teams[xmlrank.TeamId]
                if team == nil {
                    continue
                }

                rank := &parser.Rank{
                    Team:     team,
                    TeamId:   xmlrank.TeamId,
                    Position: xmlrank.Position,
                }

                competition.Ranking = append(competition.Ranking, rank)
            }
        }
    }

    return &fast, nil
}
