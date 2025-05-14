package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CoolifyClient struct {
	baseUrl    string
	token      string
	httpClient *http.Client
}

func New(baseUrl, token string) *CoolifyClient {
	return &CoolifyClient{
		baseUrl:    baseUrl,
		token:      token,
		httpClient: http.DefaultClient,
	}
}

type JustMessage struct {
	Message string `json:"message"`
}

func (c *CoolifyClient) buildPath(pathFromApiBase string) string {
	return fmt.Sprintf("%s/api/v1/%s", c.baseUrl, pathFromApiBase)
}

func (c *CoolifyClient) doRequest(request *http.Request, mapToStruct interface{}) error {
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Accept", "application/json")

	// Basic request and error handling
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// A successful request should not be 400+
	if response.StatusCode >= http.StatusBadRequest {
		// Get the body
		rawError, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		// It is probably a simple JSON message response
		message := &JustMessage{}
		err = json.Unmarshal(rawError, message)
		if err != nil {
			// It wasn't the error JSON message we were expecting, blast out the raw error
			return fmt.Errorf("%s: %s", response.Status, rawError)
		}
		return fmt.Errorf("%s: %s", response.Status, message.Message)
	}

	// Marshal to JSON
	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rawBody, mapToStruct)
	if err != nil {
		return err
	}
	return nil
}
