package errs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//
// GetViaHTTP performs a HTTP "GET" request and returns the body of the response
// from the server or an error.
//
// Callers can specify requests header and cookies, but cannot supply a request body.
//
// Any error returned will be a (pointer to a) CannotError or HttpError.
func GetViaHTTP(
	url, what string,
	headers http.Header,
	cookies []*http.Cookie,
) ([]byte, error) {

	quoteIt := false
	if what == "" {
		what = url
		quoteIt = true
	}

	request, err := newRequest("GET", url, headers, cookies)
	if err != nil {
		// Will be &CannotError{"parse", "URL", ...}
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err == nil {
		if IsHTTPerror(response.StatusCode) {
			response.Body.Close()
			err = HTTPerror(response)
		}
	}
	if err != nil {
		return nil, Cannot("fetch", "", what, quoteIt, "", err)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, Cannot("read", "", what, quoteIt, " from network", err)
	}

	return responseBody, nil
}

//
// MetaViaHTTP performs a HTTP "HEAD" request and returns the body of the response
// from the server or an error.
//
// Callers can specify requests header and cookies, but cannot supply a request body.
//
// Any error returned will be a (pointer to a) CannotError or HttpError.
func MetaViaHTTP(
	url, what string,
	headers http.Header, cookies []*http.Cookie,
) (*http.Response, error) {

	quoteIt := false
	if what == "" {
		what = url
		quoteIt = true
	}

	request, err := newRequest("HEAD", url, headers, cookies)
	if err != nil {
		// Will be &CannotError{"parse", "URL", ...}
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	response.Body.Close()
	if err == nil && IsHTTPerror(response.StatusCode) {
		err = HTTPerror(response)
	}
	if err != nil {
		return nil, Cannot("get HEADers for", "", what, quoteIt, "", err)
	}

	return response, nil
}

//
// newRequest does some setup for both GetViaHTTP and MetaViaHTTP.
//
// Any errors it returns will concern parsing the url, and will be (pointers to)
// CannotError values.
//
func newRequest(
	method, url string, headers http.Header, cookies []*http.Cookie,
) (*http.Request, error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, Cannot("parse", "URL", url, true, "", err)
	}

	if headers != nil {
		for key, values := range headers {
			for _, v := range values {
				req.Header.Add(key, v)
			}
		}
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	return req, nil
}

//
// IsHTTPerror decides whether a HTTP status code counts as an error. It simply
// regards anything outside the 2xx range as an error.
// 
func IsHTTPerror(code int) bool {
	return code/100 != 2
}

//
// A HttpError is a wrapper holding a HTTP status code and the status message
// returned by the server.
//
type HttpError struct {
	Code int
	Text string
}

//
// HTTPerror returns a (pointer to a) HttpError, taking the code and text from
// a http.Response value.
//
func HTTPerror(response *http.Response) error {
	return &HttpError{Code: response.StatusCode, Text: response.Status}
}

//
// Pointers to HttpError values satisfy the error interface.
//
func (e *HttpError) Error() string {
	codeAsText := fmt.Sprintf("%d ", e.Code)
	text := e.Text
	if strings.HasPrefix(text, codeAsText) {
		text = text[len(codeAsText):]
	}
	return fmt.Sprintf("HTTP error %d (%s)", e.Code, text)
}
