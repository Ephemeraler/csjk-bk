package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-openapi/errors"
)

var DefaultHTTPCode = http.StatusBadRequest

type apiError struct {
	Count    int32       `json:"count"`
	Previous string      `json:"previous"`
	Next     string      `json:"next"`
	Result   interface{} `json:"results"`
	Detail   string      `json:"detail"`
}

func (a *apiError) Error() string {
	return a.Detail
}

func errorAsJSON(err *apiError) []byte {
	//nolint:errchkjson
	b, _ := json.Marshal(err)
	return b
}

func New(message string, args ...interface{}) *apiError {
	if len(args) > 0 {
		return &apiError{
			Count:    -1,
			Previous: "",
			Next:     "",
			Result:   nil,
			Detail:   fmt.Sprintf(message, args...),
		}
	}
	return &apiError{
		Count:    -1,
		Previous: "",
		Next:     "",
		Result:   nil,
		Detail:   message,
	}
}

func flattenComposite(errs *errors.CompositeError) *errors.CompositeError {
	var res []error
	for _, er := range errs.Errors {
		switch e := er.(type) {
		case *errors.CompositeError:
			if e != nil && len(e.Errors) > 0 {
				flat := flattenComposite(e)
				if len(flat.Errors) > 0 {
					res = append(res, flat.Errors...)
				}
			}
		default:
			if e != nil {
				res = append(res, e)
			}
		}
	}
	return errors.CompositeValidationError(res...)
}

func ServeError(rw http.ResponseWriter, r *http.Request, err error) {
	// rw.WriteHeader(http.StatusBadRequest)
	// rw.Write(errorAsJSON(New(err.Error())))

	rw.Header().Set("Content-Type", "application/json")
	switch e := err.(type) {
	case *errors.CompositeError:
		er := flattenComposite(e)
		// strips composite errors to first element only
		if len(er.Errors) > 0 {
			ServeError(rw, r, er.Errors[0])
		} else {
			// guard against empty CompositeError (invalid construct)
			ServeError(rw, r, nil)
		}
	case *errors.MethodNotAllowedError:
		rw.Header().Add("Allow", strings.Join(e.Allowed, ","))
		rw.WriteHeader(asHTTPCode(int(e.Code())))
		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(New(e.Error())))
		}
	case errors.Error:
		fmt.Println(e.Code(), e.Error())
		value := reflect.ValueOf(e)
		if value.Kind() == reflect.Ptr && value.IsNil() {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write(errorAsJSON(New("Unknown error")))
			return
		}
		rw.WriteHeader(asHTTPCode(int(e.Code())))
		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(New(e.Error())))
		}
	case nil:
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(errorAsJSON(New("Unknown error")))
	default:
		rw.WriteHeader(http.StatusInternalServerError)
		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(New(err.Error())))
		}
	}
}

const maximumValidHTTPCode = 600

func asHTTPCode(input int) int {
	if input >= maximumValidHTTPCode {
		return DefaultHTTPCode
	}
	return input
}
