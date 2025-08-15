package main

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/service"
	"log"
	"net/http"
	"net/url"
	"time"
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

func (app *Config) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// check empty
	if code == "" {
		app.errorJSON(w, errors.New("code is required"), http.StatusBadRequest)
		return
	}

	// check base62
	if !service.IsBase62(code) {
		app.errorJSON(w, errors.New("code is invalid"), http.StatusBadRequest)
		return
	}

	// check length
	if !service.IsLengthOk(code) {
		app.errorJSON(w, errors.New("code is invalid"), http.StatusBadRequest)
		return
	}

	shortenedUrl, err := app.Models.URL.GetOne(code)
	if err != nil {
		app.errorJSON(w, errors.New("no result found"), http.StatusNotFound)
		return
	}

	// Update hit count
	app.updateHitCount(code)

	http.Redirect(w, r, shortenedUrl.OriginalURL, http.StatusFound)
}

func (app *Config) updateHitCount(c string) {
	go func(c string) {
		const maxRetries = 3
		const retryDelay = 200 * time.Millisecond

		for attempt := 1; attempt <= maxRetries; attempt++ {
			err := app.Models.URL.IncrementHitCount(c)
			if err == nil {
				return
			}

			log.Printf("failed to update hit count (attempt %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
			}
		}
	}(c)
}
