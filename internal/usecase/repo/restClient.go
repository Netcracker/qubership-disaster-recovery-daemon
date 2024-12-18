// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"fmt"
	"io"
	"net/http"
)

func NewRestClient(url string, httpClient http.Client) *RestClient {
	return &RestClient{
		url:        url,
		httpClient: httpClient,
	}
}

type RestClient struct {
	url        string
	httpClient http.Client
}

func (rc RestClient) SendRequest(method string, path string, body io.Reader) (statusCode int, responseBody []byte, err error) {
	requestUrl := fmt.Sprintf("%s%s", rc.url, path)
	request, err := http.NewRequest(method, requestUrl, body)
	if err != nil {
		return
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")
	response, err := rc.httpClient.Do(request)
	if err != nil {
		return
	}
	statusCode = response.StatusCode
	responseBody, err = io.ReadAll(response.Body)
	return
}
