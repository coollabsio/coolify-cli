package client

import "net/http"

type Resource struct {
	ID     int    `json:"id"`
	Uuid   string `json:"uuid"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

func (c *CoolifyClient) ListResources() ([]Resource, error) {
	request, err := http.NewRequest(http.MethodGet, c.buildPath("resources"), nil)
	if err != nil {
		return nil, err
	}

	var resources []Resource
	err = c.doRequest(request, &resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}
