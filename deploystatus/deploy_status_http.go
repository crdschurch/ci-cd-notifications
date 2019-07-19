// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START functions_helloworld_http]

// Package helloworld provides a set of Cloud Functions samples.
package deploystatus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
)

var DEPLOYSTATUSCHANNEL = "C835A5T0U"

func sendSlackNotification(authToken string, body string) error {
	log.Printf(body)
	slackBody := []byte(body)
	req, err := http.NewRequest(http.MethodPost, "https://api.slack.com/api/chat.postMessage", bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + authToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	var raw map[string]interface{}

	if e := json.Unmarshal([]byte(buf.String()), &raw); e != nil {
		return errors.New(fmt.Sprintf("500 - Unable to parse result from slack, error: %v", e))
	}

	if raw["ok"] != true {
		return errors.New("Non-ok response returned from Slack")
	}

	return nil
}

func getQueryStringParam(w http.ResponseWriter, r *http.Request, k string) (string, bool) {
	param := r.URL.Query().Get(k)

	if len(param) < 1 {
		responseString := fmt.Sprintf("400 - Missing query string parameter for %v", k)
		log.Printf(responseString)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(responseString))
		return "", false
	}

	return param, true
}

func logAndWrite(w http.ResponseWriter, message string, statusCode int) {
	log.Printf(message)
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var body map[string]interface{}
	err := decoder.Decode(&body)
	if err != nil {
		logAndWrite(w, fmt.Sprintf("400 - Unable to parse body, error: %v", err), http.StatusBadRequest)
		return
	}

	channel, ok := getQueryStringParam(w, r, "channel")
	if !ok {
		return
	}

	secondaryChannel := r.URL.Query().Get("secondary_channel")

	site_id := body["site_id"]
	branch := body["branch"]
	context := body["context"]
	name := body["name"]
	url := body["ssl_url"]
	commit_url := body["commit_url"]
	committer := body["committer"]

	//This is hacky but it stops any non production deploys in netlify from having their status published in deploy-status
	if isNetlify(site_id) && channel == DEPLOYSTATUSCHANNEL && isProdDeploy(context, branch) {
		logAndWrite(w, fmt.Sprintf("Did not publish status to deploy-status channel because this was not a production deploy"), 200)
		return
	} else if isNetlify(site_id) && secondaryChannel != "" {
		channel = secondaryChannel
	}

	formattedBody := fmt.Sprintf(`{
		"channel": "%v",
		"blocks": [
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "*%v* deployed to *%v* by %v from *branch:* %v [<%v|see commit>] [<%v|validate>]"
				},
				"accessory": {
					"type": "image",
					"image_url": "https://cataas.com/cat/jump/says/%v%%0Adeployed!?t=or",
					"alt_text": "meow"
				}
			}
		]
	}`, channel, name, context, committer, branch, commit_url, url, committer)

	slackToken := os.Getenv("SLACK_TOKEN")
	slackErr := sendSlackNotification(slackToken, formattedBody)
	if slackErr != nil {
		responseString := fmt.Sprintf("500 - Failed to send slack notification due to error: %v", slackErr)
		logAndWrite(w, responseString, 500)
		return
	}

	logAndWrite(w, fmt.Sprintf("Successfully published deploy status: %v", formattedBody), 200)
}

func isNetlify(siteId interface{}) bool {
	return siteId != nil
}

func isProdDeploy(context, branch interface{}) bool {
	return context == "production" || branch == "master"
}