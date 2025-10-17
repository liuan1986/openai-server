package proxy

import (
	"bytes"
	"io"
	"net/http"
)

type Client struct {
	targetURL string
	apiKey    string
	client    *http.Client
}

func New(targetURL, apiKey string) *Client {
	return &Client{
		targetURL: targetURL,
		apiKey:    apiKey,
		client:    &http.Client{},
	}
}

func (c *Client) Forward(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(body))

	outbound, err := http.NewRequest(req.Method, c.targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	outbound.Header = req.Header.Clone()
	outbound.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(outbound)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func CopyResponse(w http.ResponseWriter, resp *http.Response) {
	defer resp.Body.Close()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
