package web

import "net/http"
import "io/ioutil"

import "strings"

// Execute assumes that the address provided is a HTTP endpoint and posts the payload to it.
func Execute(address string, body string) (resp string, err error) {
	buf := strings.NewReader(body)
	response, err := http.Post(address, "application/json", buf)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return string(bytes), err
}
