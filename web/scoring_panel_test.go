// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Techno11/cheesy-arena/field"
	"github.com/Techno11/cheesy-arena/game"
	"github.com/Techno11/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScoringPanel(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/panels/scoring/invalidalliance")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid alliance")
	recorder = web.getHttpResponse("/panels/scoring/red")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Scoring Panel - Untitled Event - Cheesy Arena")
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)
	redConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red/websocket", nil)
	assert.Nil(t, err)
	defer redConn.Close()
	redWs := websocket.NewTestWebsocket(redConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))
	blueConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blue/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := websocket.NewTestWebsocket(blueConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))

	// Should get a few status updates right after connection.
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "matchTime")
	readWebsocketType(t, blueWs, "realtimeScore")

	// Send some autonomous period scoring commands.
	web.arena.MatchState = field.AutoPeriod
	redWs.Write("1", nil)
	redWs.Write("3", nil)
	redWs.Write("w", nil)
	redWs.Write("X", nil)
	redWs.Write("x", nil)
	redWs.Write("z", nil)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, [3]bool{true, false, true}, web.arena.RedRealtimeScore.CurrentScore.ExitedInitiationLine)
	assert.Equal(t, [2]int{1, 0}, web.arena.RedRealtimeScore.CurrentScore.AutoCellsBottom)
	assert.Equal(t, [2]int{0, 0}, web.arena.RedRealtimeScore.CurrentScore.AutoCellsOuter)
	assert.Equal(t, [2]int{1, 0}, web.arena.RedRealtimeScore.CurrentScore.AutoCellsInner)

	// Send some teleoperated period scoring commands.
	web.arena.MatchState = field.TeleopPeriod
	blueWs.Write("f", nil)
	blueWs.Write("F", nil)
	blueWs.Write("o", nil)
	blueWs.Write("5", nil)
	blueWs.Write("5", nil)
	blueWs.Write("L", nil)
	blueWs.Write("k", nil)
	for i := 0; i < 6; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, [4]int{2, 0, 0, 0}, web.arena.BlueRealtimeScore.CurrentScore.TeleopCellsOuter)
	assert.Equal(t, [3]game.EndgameStatus{game.EndgameNone, game.EndgameHang, game.EndgameNone},
		web.arena.BlueRealtimeScore.CurrentScore.EndgameStatuses)
	assert.Equal(t, true, web.arena.BlueRealtimeScore.CurrentScore.RungIsLevel)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
	web.arena.MatchState = field.PostMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // Allow some time for the commands to be processed.
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))

	// Load another match to reset the results.
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, field.NewRealtimeScore(), web.arena.BlueRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
}
