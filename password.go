package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type unispherePasswordClient struct {
	tgt url.URL
	hc  HTTPClient
	ctx context.Context
	usr string
	pw  string
}

func (c *unispherePasswordClient) newGetRequest(url string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(c.ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	encoder := base64.StdEncoding.EncodeToString([]byte(c.usr + ":" + c.pw))
	r.Header.Add("Authorization", "Basic "+encoder)
	r.Header.Add("X-EMC-REST-CLIENT", "true")
	r.Header.Add("Content-Type", "application/json")

	return r, nil
}

func (c *unispherePasswordClient) newPostRequest(url string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(c.ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	encoder := base64.StdEncoding.EncodeToString([]byte(c.usr + ":" + c.pw))
	r.Header.Add("Authorization", "Basic "+encoder)
	r.Header.Add("X-EMC-REST-CLIENT", "true")
	r.Header.Add("Content-Type", "application/json")

	return r, nil
}

func (c *unispherePasswordClient) Get(path string, query string, obj interface{}) error {
	u := c.tgt
	u.Path = path
	u.RawQuery = query

	req, err := c.newPostRequest(u.String())
	if err != nil {
		return err
	}

	req = req.WithContext(c.ctx)
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Response code was %d, expected 200", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, obj)
}

func (c *unispherePasswordClient) String() string {
	return c.tgt.String()
}

func newUnispherePasswordClient(ctx context.Context, tgt url.URL, hc HTTPClient, usr string, pw string) (*unispherePasswordClient, error) {
	return &unispherePasswordClient{tgt, hc, ctx, usr, pw}, nil
}
