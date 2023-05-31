package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Id     string `json:"clientid"` // capital letter indicates public fields
	Secret string `json:"secret"`
}

type ServerResponse struct {
	Token      string `json:"access_token"`
	Scope      string `json:"scope"`
	Token_type string `json:"bearer"`
	Expires_in string `json:"expires_in"`
}

func errorCheck(e error, s string) {
	if e != nil {
		fmt.Println(s, e.Error())
		os.Exit(-1)
	}
}

func main() {
	// read clientID and secret
	in_conf_byte, err := ioutil.ReadFile("indeed-config.json")
	errorCheck(err, "Unable to open file: ")

	// assign clientID and secret
	indeed_config := Config{}
	err = json.Unmarshal(in_conf_byte, &indeed_config)
	errorCheck(err, "Unable to parse config json: ")

	// Setup url and form field
	u := "https://apis.indeed.com/oauth/v2/tokens"
	formData := url.Values{
		"grant_type":    {"client_credentials"},
		"scope":         {"employer_access"},
		"client_id":     {indeed_config.Id},
		"client_secret": {indeed_config.Secret},
	}

	// Setup client request
	client := http.Client{}
	req, err := http.NewRequest("POST", u, strings.NewReader(formData.Encode()))
	errorCheck(err, "Unable to Post to URL")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	// Get server response
	resp, err := client.Do(req)
	errorCheck(err, "Response error: ")
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	errorCheck(err, "Error Reading Response: ")

	var dat map[string]interface{}
	err = json.Unmarshal(body, &dat)
	errorCheck(err, "Unable to parse response json: ")
	fmt.Println(dat["access_token"])

}
