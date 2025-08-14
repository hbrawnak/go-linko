package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// jsonResponse defines the standard structure for all JSON responses from the server.
type jsonResponse struct {
	Error   bool   `json:"error"`          // Indicates if the response represents an error (true = error, false = success)
	Message string `json:"message"`        // A human-readable message about the response (success message or error description)
	Data    any    `json:"data,omitempty"` // Optional field for sending additional data (empty if not needed)
}

// readJSON is a helper function to read and decode JSON from an HTTP request body into a provided Go variable.
// It protects against large payloads by setting a maximum byte limit and ensures that only a single JSON object is sent.
func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	// Set a maximum size for the request body (10MB in this case) to prevent abuse
	maxBytes := 10485760 // one mb

	// Wrap the original request body to enforce the maximum size
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Create a new JSON decoder
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	// Try to decode again to ensure there's no extra unexpected data
	// If decoding doesn't immediately hit EOF (end of file), the request body had multiple JSON objects — which is invalid
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("invalid JSON: body must contain only a single JSON object")
	}

	return nil
}

// writeJSON is a helper function to send JSON responses back to the client.
// It marshals the provided data into JSON format, sets appropriate headers, and writes it to the ResponseWriter.
// Optionally, it allows setting additional custom headers through the variadic headers parameter.
func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	// Marshal the data into a JSON byte slice
	output, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// If additional headers are provided, set them in the response
	if len(headers) > 0 {
		for key, val := range headers[0] {
			w.Header()[key] = val
		}
	}

	// Set the Content-Type to application/json and the HTTP status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Write the JSON output to the response body
	_, err = w.Write(output)
	if err != nil {
		return err
	}

	return nil
}

// errorJSON is a helper function to send an error response in JSON format.
// It automatically formats the error message and lets you optionally set a custom HTTP status code.
func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	// Set default status code to 400 Bad Request
	statusCode := http.StatusBadRequest

	// If a custom status code is provided, override the default
	if len(status) > 0 {
		statusCode = status[0]
	}

	// Create the payload with error information
	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	// Write the payload as JSON to the response writer
	return app.writeJSON(w, statusCode, payload)
}
