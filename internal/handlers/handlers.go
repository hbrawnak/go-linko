package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/service"
	"github.com/hbrawnak/go-linko/internal/utils"
	"net/http"
	"net/url"
	"os"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

// AppHandler holds all dependencies needed for HTTP handlers
type AppHandler struct {
	Service  *service.Service
	Response *utils.Response
}

// NewHandler creates a new handler instance with dependencies
func NewHandler(service *service.Service) *AppHandler {
	return &AppHandler{
		Service:  service,
		Response: &utils.Response{},
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

	if req.URL == "" {
		app.Response.ErrorJSON(w, errors.New("url is required"), http.StatusBadRequest)
		return
	}

	_, err := url.ParseRequestURI(req.URL)
	if err != nil {
		app.Response.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	code, err := app.Service.GenerateShotCode(7)

	// 1. save data in db
	u := data.URL{
		ShortCode:   code,
		OriginalURL: req.URL,
	}

	_, err = app.Service.Models.URL.Insert(u)
	if err != nil {
		app.Response.ErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// storing cache in background
	app.Service.StoreInRedisCacheBG(u.ShortCode, u.OriginalURL)

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

	// check empty
	if code == "" {
		app.Response.ErrorJSON(w, errors.New("code is required"), http.StatusBadRequest)
		return
	}

	// check base62
	if !app.Service.IsBase62(code) {
		app.Response.ErrorJSON(w, errors.New("code is invalid"), http.StatusBadRequest)
		return
	}

	// check length
	if !app.Service.IsLengthOk(code) {
		app.Response.ErrorJSON(w, errors.New("code is invalid"), http.StatusBadRequest)
		return
	}

	if originalUrl, err := app.Service.Redis.Get(code); err == nil {
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

	// storing cache in background
	app.Service.StoreInRedisCacheBG(shortenedUrl.ShortCode, shortenedUrl.OriginalURL)

	// Update hit count
	app.Service.UpdateHitCountBG(shortenedUrl.ShortCode)

	http.Redirect(w, r, shortenedUrl.OriginalURL, http.StatusFound)
}
