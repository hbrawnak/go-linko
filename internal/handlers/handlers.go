package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"github.com/hbrawnak/go-linko/internal/service"
	"github.com/hbrawnak/go-linko/internal/utils"
	"github.com/hbrawnak/go-linko/internal/worker"
	"net/http"
	"os"
	"sync"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type StatsDataResp struct {
	Code  string `json:"code"`
	Count int64  `json:"count"`
}

type AppHandler struct {
	Service      *service.Service
	Response     *utils.Response
	URLTaskQueue chan worker.URLTask
}

func NewHandler(service *service.Service, queue chan worker.URLTask) *AppHandler {
	return &AppHandler{
		Service:      service,
		Response:     &utils.Response{},
		URLTaskQueue: queue,
	}
}

func (app *AppHandler) HandleMain(w http.ResponseWriter, r *http.Request) {
	payload := utils.JsonResponse{
		Error:   false,
		Message: "Welcome to URL Shortener API",
	}

	_ = app.Response.WriteJSON(w, http.StatusOK, payload)
}

func (app *AppHandler) HandleShorten(w http.ResponseWriter, r *http.Request) {
	var baseUrl = os.Getenv("BASE_URL")
	var req ShortenRequest

	if err := app.Response.ReadJSON(w, r, &req); err != nil {
		app.Response.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	if err := utils.ValidateOriginalURL(req.URL); err != nil {
		app.Response.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	code := app.Service.GenerateShortCode()

	// 1. save data in db
	u := data.URL{
		ShortCode:   code,
		OriginalURL: req.URL,
	}

	fields := database.CachedURL{
		URL:       req.URL,
		Persisted: "0",
	}

	// storing cache using wait group
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = app.Service.Redis.HSet(u.ShortCode, fields.ToMap())
	}()

	wg.Wait()

	// Sending to worker queue to make db operations
	app.URLTaskQueue <- worker.URLTask{
		ShortCode:   u.ShortCode,
		OriginalURL: u.OriginalURL,
	}

	shortUrlResp := map[string]string{
		"short_url": fmt.Sprintf("%s/%s", baseUrl, u.ShortCode),
		"code":      u.ShortCode,
	}

	// 2 return response
	var payload utils.JsonResponse
	payload.Error = false
	payload.Message = "URL Shortened"
	payload.Data = shortUrlResp

	_ = app.Response.WriteJSON(w, http.StatusOK, payload)
}

func (app *AppHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Validating short code
	if err := utils.ValidateShortCode(code); err != nil {
		app.Response.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	if originalUrl, err := app.Service.Redis.HGet(code, "url"); err == nil {
		// Update hit count
		app.Service.UpdateHitCountBG(code)
		http.Redirect(w, r, originalUrl, http.StatusFound)
		return
	}

	shortenedUrl, err := app.Service.Models.URL.GetOne(code)
	if err != nil {
		app.Response.ErrorJSON(w, errors.New("no result found"), http.StatusNotFound)
		return
	}

	fields := database.CachedURL{
		URL:       shortenedUrl.OriginalURL,
		Persisted: "1",
	}
	// storing cache in background
	app.Service.StoreInRedisCacheBG(shortenedUrl.ShortCode, fields.ToMap())

	// Update hit count
	app.Service.UpdateHitCountBG(shortenedUrl.ShortCode)

	http.Redirect(w, r, shortenedUrl.OriginalURL, http.StatusFound)
}

func (app *AppHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Validating short code
	if err := utils.ValidateShortCode(code); err != nil {
		app.Response.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	stats, err := app.Service.GetStats(code)
	if err != nil {
		app.Response.ErrorJSON(w, err, http.StatusNotFound)
		return
	}

	payload := utils.JsonResponse{
		Message: "Stats Data",
		Error:   false,
		Data:    stats,
	}

	_ = app.Response.WriteJSON(w, http.StatusOK, payload)
}
