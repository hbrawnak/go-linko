package main

import (
	"errors"
	"fmt"
	"github.com/hbrawnak/go-linko/data"
	"github.com/hbrawnak/go-linko/internal/service"
	"net/http"
	"net/url"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

func (app *Config) HandleMain(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Welcome to URL Shortener API",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	if err := app.readJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		app.errorJSON(w, errors.New("url is required"), http.StatusBadRequest)
		return
	}

	_, err := url.ParseRequestURI(req.URL)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	shortCode, err := service.GenerateShotCode(7)

	// 1. save data in db
	u := data.URL{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
	}

	_, err = app.Models.URL.Insert(u)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	shortUrlResp := map[string]string{
		"short_url": fmt.Sprintf("http://localhost:8080/%s", shortCode),
	}

	// 2 return response
	var payload jsonResponse
	payload.Error = false
	payload.Message = "URL Shortened"
	payload.Data = shortUrlResp

	_ = app.writeJSON(w, http.StatusOK, payload)
}
