package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// readIdParam convert id parameter into int based 10 with 64 bits.
// returns an id and correspond error
func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelop map[string]any

/*
writeJSON helper for sending responses. This takes the destination http.ResponseWriter, the HTTP status code to send,
the data to encode to JSON, and a header map.

USAGE:

	data := envelop{
		"status": "available",
		"system_information": map[string]string{
		"environment": app.config.env,
		"version":     version,
		},
	}

	// sending responses to users.
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
	app.serverErrorResponse(w, r, err)
	}
*/
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelop, headers http.Header) error {
	// Encode the data to JSON
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Add any headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Set application/json header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

/*
The readJSON function decode the request body into the target destination.
Also check the vulnerabilities from Developers, Users, Expected errors, or Unexpected errors.

USAGE:

	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

// read the request body
err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}
*/
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Use http.MaxBytesReader() to limit the size of the request to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
			maxBytesError         *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unkown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	// Handle another JSON after the first JSON body
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// The readString helper returns a string value from the query string, or the provided default value
// if no matching key could be found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// The readCSV helper reads a string value from the query string and then splits it
// into a slice on the comma character. If no matching key could be found, it returns
// the provided default value
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// query-string: ?genres=drama,action
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// The readInt helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record
// an error message in the provided Validator instance
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validation.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}
