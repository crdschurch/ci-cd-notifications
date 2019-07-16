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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"bytes"
	"time"
	"errors"
)

var slackChannelMap = map[string]string {
	"devy_mcopsface": "https://hooks.slack.com/services/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",
	"deploy-status": "https://hooks.slack.com/services/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",
}

type SlackRequestBody struct {
    Text string `json:"text"`
}

func sendSlackNotification(webhookUrl string, msg string) error {

    slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
    req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
    if err != nil {
        return err
    }

    req.Header.Add("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(resp.Body)
    if buf.String() != "ok" {
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

	site_id := body["site_id"]
	branch := body["branch"]
	context := body["context"]
	name := body["name"]
	deploy_url := body["deploy_url"]
	commit_url := body["commit_url"]
	committer := body["committer"]
	review_url := body["review_url"]

	//This is hacky but it stops any non production deploys from having their status published in deploy-status
	if site_id != nil && channel == "deploy-status" && (context != "production" || branch != "master") {
		logAndWrite(w, fmt.Sprintf("Did not publish status to deploy-status channel because this was not a production deploy"), 200)
		return
	}

	webhookUrl := slackChannelMap[channel]
	if len(webhookUrl) < 1 {
		responseString := fmt.Sprintf("400 - No slack webhook found for channel: %v. You must add an entry to the slack channels map.", channel)
		logAndWrite(w, responseString, http.StatusBadRequest)
		return
	}

	formattedSting := fmt.Sprintf("*%v* deployed to *%v*.\n*committer:* %v\n*commit_url:* %v\n*review_url:* %v\n*deploy_url:* %v\n" +
	"*branch:* %v", name, context, committer, commit_url, review_url, deploy_url, branch)
	slackErr := sendSlackNotification(webhookUrl, formattedSting)
    if slackErr != nil {
		responseString := fmt.Sprintf("500 - Failed to send slack notification due to error: %v", slackErr)
		logAndWrite(w, responseString, 500)
		return
    }
	
	logAndWrite(w, fmt.Sprintf("Successfully published deploy status: %v", formattedSting), 200)
}

// func main() {
//     http.HandleFunc("/", Handler)
//     http.ListenAndServe(":8080", nil)
// }

// [END functions_helloworld_http]
