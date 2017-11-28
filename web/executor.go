package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Execute assumes that the address provided is a HTTP endpoint and posts the payload to it.
func Execute(address string, body string) (resp string, err error) {
	buf := strings.NewReader(body)
	response, err := http.Post(address, "application/json", buf)
	if err != nil {
		return "", err
	}
	if response.StatusCode < 200 || response.StatusCode > 300 {
		err = fmt.Errorf("received status code: %v", response.StatusCode)
	}
	bytes, readErr := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err == nil {
		err = readErr
	}
	return string(bytes), err
}
