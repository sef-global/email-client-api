package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"github.com/mayura-andrew/email-client/internal/validator"

)

// define an envelope type
type envelop map[string]interface{}

// retrieve the "id" URL parameter from the current request context, then convert it to,
// an integer and return it. If the operation isn't successful, return 0 and an error.

// we can then use the ByName() method to get the value of the "id" parameter from
// the slice. In our project all movies will have a unique positive integer ID, but
// the value returned by ByName() is always a string. So we try to convert it to a base 10 integer (with a bit size of 64). If the parameter couldn't be converted,
// function to return a 404 Not Found response.

// when httprouter is parsing a request, any interpolated URL parameters will be
// stored in the request context. we can use the ParamsFromContext() function to
// retrieve a slice containing these parameter names and values.

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

// writeJSON() helper for sending responses.
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelop, headers http.Header) error {
	// encode the data to JSON, returning the error if there was one.
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// append a newline to make it easier to view in terminal application.
	js = append(js, '\n')

	// loop through the header map and add each header to the http.ResponseWriter header map.
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// use http.MaxBytesReader() to limit the size of the request body to 1 MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	// initialize the json.Decoder, and call the DisallowUnknownFields() method on it before decoding. this means that if the JSON from the client now includes any field which cannot be mapped to the target
	// destination, the decoder will return an error instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// decode the request body to the destination
	err := dec.Decode(dst)

	// Decode the request body into the target destination.
	if err != nil {
		// if there is n error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// use the errors.As() function to check whether the error has the tye
		// *json.SyntaxError. if it does, then return a plain-english error message
		// which includes the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		//In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

			// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
		// JSON value is the wrong type for the target destination. if the error
		// related to a specific field, then we include, that int our error message to make it
		// easier for the client to debug.

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

			// an io.EOF error will be returned by decode() if the request body is empty. we
			// check for this with errors.Is() and return a plain-english error message instead.
		case errors.Is(err, io.EOF):
			return errors.New("body not be empty")

		// if the JSON contains field which cannot be mapped  to the target destination
		// then Decode() will now return an error message in the format "json: unknown field "<name>".
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "jsonL unknown field ")
			return fmt.Errorf("body contanins unknown key %s", fieldName)

		// if the request body exceeds 1MB in size the decode will now fail with the
		// error http: request body too large.
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// For anything else, return the error message as-is.
		default:
			return err
		}

	}
	// call Decode() again using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a signal JSON value this will
	// return an io.EOF error. so if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// the readString() helper returns a string value from the query string, or the provided default
// value if no matching key could be found.

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string. If no key exists this
	// will return the empty string "".
	s := qs.Get(key)

	// If no key exists (or the value is empty) then return the default value.

	if s == "" {
		return defaultValue
	}

	// otherwise return the string.

	return s
}

// The readCSV() helper reads string value from the query string and then splits it
// into a slice on the comma character. If no matching key could be found, it returns
// the provided default value.

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// extract the value from the query string.

	csv := qs.Get(key)

	// if no key exists (or the value is empty) then return the default value.

	if csv == "" {
		return defaultValue
	}

	// otherwise parse the value into a []string slice and return it.
	return strings.Split(csv, ",")
}

// the readInt() helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// extract the value from the query string
	s := qs.Get(key)

	// if no key exists (or the value is empty) then return the default value.
	if s == "" {
		return defaultValue
	}

	// try to convert the value to an int. if this fails, add an error message to the
	// validator instance and return the default value.

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an intger value")
		return defaultValue
	}

	// otherwise, return the converted integer value.
	return i
}

// the background() helper accepts an arbitrary function as a parameter.

func (app *application) background(fn func()) {
	// launch a background goroutine.
	go func() {
		// recover any panic.
		if err := recover(); err != nil {
			app.logger.PrintError(fmt.Errorf("%s", err), nil)
		}
	}()

	// execute the arbitrary function that we passed as the parameter.
	fn()
}
