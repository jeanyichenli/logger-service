package main

import (
	"log-service/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLogs(w http.ResponseWriter, r *http.Request) {
	// read json
	var payload JSONPayload
	_ = app.readJson(w, r, &payload)

	event := data.LogEntry{
		Name: payload.Name,
		Data: payload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJson(w, http.StatusAccepted, resp)
}
