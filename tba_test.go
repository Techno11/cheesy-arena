// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"bytes"
	"encoding/json"
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublishTeams(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	eventSettings.TbaEventCode = "my_event_code"
	eventSettings.TbaSecretId = "my_secret_id"
	eventSettings.TbaSecret = "my_secret"
	db.CreateTeam(&Team{Id: 254})
	db.CreateTeam(&Team{Id: 1114})

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "event/my_event_code")
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "[\"frc254\",\"frc1114\"]", reader.String())
		assert.Equal(t, "my_secret_id", r.Header["X-Tba-Auth-Id"][0])
		assert.Equal(t, "f5c022fde6d1186ea0719fe28ab6cc63", r.Header["X-Tba-Auth-Sig"][0])
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishTeams())
}

func TestPublishMatches(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	match1 := Match{Type: "qualification", DisplayName: "2", Time: time.Unix(600, 0), Red1: 7, Red2: 8, Red3: 9,
		Blue1: 10, Blue2: 11, Blue3: 12, Status: "complete"}
	match2 := Match{Type: "elimination", DisplayName: "SF2-2", ElimRound: 2, ElimGroup: 2, ElimInstance: 2}
	db.CreateMatch(&match1)
	db.CreateMatch(&match2)
	matchResult1 := buildTestMatchResult(match1.Id, 1)
	db.CreateMatchResult(matchResult1)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		var matches []*TbaMatch
		json.Unmarshal(body, &matches)
		assert.Equal(t, 2, len(matches))
		assert.Equal(t, "qm", matches[0].CompLevel)
		assert.Equal(t, "sf", matches[1].CompLevel)
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishMatches())
}

func TestPublishRankings(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	db.CreateRanking(game.TestRanking2())
	db.CreateRanking(game.TestRanking1())

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		var response TbaRankings
		json.Unmarshal(body, &response)
		assert.Equal(t, 2, len(response.Rankings))
		assert.Equal(t, "frc254", response.Rankings[0].TeamKey)
		assert.Equal(t, "frc1114", response.Rankings[1].TeamKey)
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishRankings())
}

func TestPublishAlliances(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	buildTestAlliances(db)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader bytes.Buffer
		reader.ReadFrom(r.Body)
		assert.Equal(t, "[[\"frc254\",\"frc469\",\"frc2848\",\"frc74\"],[\"frc1718\",\"frc2451\"]]",
			reader.String())
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.Nil(t, PublishAlliances())
}

func TestPublishingErrors(t *testing.T) {
	clearDb()
	defer clearDb()
	var err error
	db, err = OpenDatabase(testDbPath)
	assert.Nil(t, err)
	defer db.Close()
	eventSettings, _ = db.GetEventSettings()
	buildTestAlliances(db)

	// Mock the TBA server.
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oh noes", 500)
	}))
	defer tbaServer.Close()
	tbaBaseUrl = tbaServer.URL

	assert.NotNil(t, PublishTeams())
	assert.NotNil(t, PublishMatches())
	assert.NotNil(t, PublishRankings())
	assert.NotNil(t, PublishAlliances())
}
