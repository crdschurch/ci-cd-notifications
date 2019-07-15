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
package helloworld

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
)

// Handle request from Team City
func TeamCity(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Name string `json:"name"`
		UserName string `json:"teamcity.build.triggeredBy.username"`
		URL string `json:"vcsroot.Github.url"`
		Branch string `json:"teamcity.build.vcs.branch.Github"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		fmt.Fprint(w, "Hello, World 1!")
		//log.Println("This is stderr")
		fmt.Println("1")
		return
	}
	if d.Name == "" {
		fmt.Fprint(w, "Hello, World 2!")
		fmt.Println("2")
		return
	}
	fmt.Fprintf(w, "Hello, %s!", html.EscapeString(d.UserName))
	fmt.Printf("3 Hello, %v! URL=%v,Branch=%v, ", html.EscapeString(d.UserName), html.EscapeString(d.URL), html.EscapeString(d.Branch))

}

// [END functions_helloworld_http]
