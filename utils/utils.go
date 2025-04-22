package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type UnisphereHTTP interface {
	Get(path string, query string, obj interface{}) error
	Post(path string, query string, body []byte, obj interface{}) error
}

type UnisphereClient struct {
	tgt     url.URL
	hc      HTTPClient
	ctx     context.Context
	auth    string
	token   string
	cookies []*http.Cookie
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c *UnisphereClient) Get(path string, query string, obj interface{}) error {
	u := c.tgt
	u.Path = path
	u.RawQuery = query

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	req.WithContext(c.ctx)
	var co *http.Cookie
	var _ int
	for _, co = range c.cookies {
		req.AddCookie(co)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &obj)
}

func (c *UnisphereClient) Post(path string, query string, body []byte, obj interface{}) error {
	u := c.tgt
	u.Path = path
	u.RawQuery = query
	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.auth)
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	req.Header.Add("EMC-CSRF-TOKEN", c.token)
	req.WithContext(c.ctx)
	var co *http.Cookie
	var _ int
	for _, co = range c.cookies {
		req.AddCookie(co)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(b, &obj)
}
func GetToken(ctx context.Context, u url.URL, hc *http.Client, auth string) (*UnisphereClient, error) {
	u.Path = "/api/types/user/instances"
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	req.Header.Add("Authorization", "Basic "+auth)
	if err != nil {
		return nil, err
	}
	resp, err := hc.Do(req)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	cookies := resp.Cookies()
	token := resp.Header.Get("EMC-CSRF-TOKEN")

	return &UnisphereClient{u, hc, ctx, auth, token, cookies}, nil
}
