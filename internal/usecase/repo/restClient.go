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
