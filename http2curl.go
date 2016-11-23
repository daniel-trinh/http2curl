package http2curl

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// CurlCommand contains exec.Command compatible slice + helpers
type CurlCommand struct {
	slice []string
}

// append appends a string to the CurlCommand
func (c *CurlCommand) append(newSlice ...string) {
	c.slice = append(c.slice, newSlice...)
}

// String returns a ready to copy/paste command
func (c *CurlCommand) String() string {
	return strings.Join(c.slice, " ")
}

// nopCloser is used to create a new io.ReadCloser for req.Body
type nopCloser struct {
	io.Reader
}

func bashEscape(str string) string {
	return `'` + strings.Replace(str, `'`, `'\''`, -1) + `'`
}

func (nopCloser) Close() error { return nil }

// GetCurlCommand returns a CurlCommand corresponding to an http.Request
func GetCurlCommand(req *http.Request, c *http.Client) (*CurlCommand, error) {
	command := CurlCommand{}

	command.append("curl")

	command.append("-X", bashEscape(req.Method))

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = nopCloser{bytes.NewBuffer(body)}
		bodyEscaped := bashEscape(string(body))
		command.append("-d", bodyEscaped)
	}

	var keys []string

	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		command.append("-H", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " "))))
	}

	cookies := make([]*http.Cookie, 0)

	// Cookies from request
	for _, cookie := range req.Cookies() {
		cookies = append(cookies, cookie)
	}
	// Cookies from jar

	if c != nil && c.Jar != nil {
		fmt.Println(c.Jar.Cookies(req.URL))
		cookies = append(cookies, c.Jar.Cookies(req.URL)...)
	}

	command.append("-H", bashEscape(cookieStringFromCookies(cookies)))
	command.append(bashEscape(req.URL.String()))

	return &command, nil
}

func cookieStringFromCookies(cookies []*http.Cookie) string {
	cookieStrings := []string{}
	for _, v := range cookies {
		cookieStrings = append(cookieStrings, v.Name+"="+v.Value)
	}
	cookieString := "Cookie: "+strings.Join(cookieStrings, "; ")
	return cookieString
}
