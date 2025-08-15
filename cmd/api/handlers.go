package main

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/service"
	"net/http"
	"net/url"
	"os"
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
	var baseUrl = os.Getenv("BASE_URL")
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

	code, err := service.GenerateShotCode(7)

	// 1. save data in db
	u := data.URL{
		ShortCode:   code,
		OriginalURL: req.URL,
	}

	_, err = app.Models.URL.Insert(u)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// storing cache in background
	app.Service.StoreInRedisCacheBG(u.ShortCode, u.OriginalURL)

	shortUrlResp := map[string]string{
		"short_url": fmt.Sprintf("%s/%s", baseUrl, u.ShortCode),
		"code":      u.ShortCode,
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

	if originalUrl, err := app.Redis.Get(code); err == nil {
		// Update hit count
		app.Service.UpdateHitCountBG(code)
		http.Redirect(w, r, originalUrl, http.StatusFound)
		return
	}

	shortenedUrl, err := app.Models.URL.GetOne(code)
	if err != nil {
		app.errorJSON(w, errors.New("no result found"), http.StatusNotFound)
		return
	}

	// storing cache in background
	app.Service.StoreInRedisCacheBG(shortenedUrl.ShortCode, shortenedUrl.OriginalURL)

	// Update hit count
	app.Service.UpdateHitCountBG(shortenedUrl.ShortCode)

	http.Redirect(w, r, shortenedUrl.OriginalURL, http.StatusFound)
}
