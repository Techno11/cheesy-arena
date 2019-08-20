// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for testing the field LEDs and PLC.

package web

import (
	//"fmt"
	//"github.com/Techno11/cheesy-arena/field"
	//"github.com/Techno11/cheesy-arena/led"
	"github.com/Techno11/cheesy-arena/model"
	"github.com/Techno11/cheesy-arena/websocket"
	//"github.com/mitchellh/mapstructure"
	//"io"
	//"log"
	"net/http"
)

// Shows the LED/PLC test page.
func (web *Web) ledPlcGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_led_plc.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	plc := web.arena.Plc
	data := struct {
		*model.EventSettings
		InputNames    []string
		RegisterNames []string
		CoilNames     []string
	}{web.arena.EventSettings, plc.GetInputNames(), plc.GetRegisterNames(), plc.GetCoilNames()}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for sending realtime updates to the LED/PLC test page.
func (web *Web) ledPlcWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	ws.HandleNotifiers(web.arena.Plc.IoChangeNotifier)
}
