package gounity

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
	"unisphere_exporter/gounity/api"

	"github.com/tidwall/gjson"

	"go.opentelemetry.io/otel/sdk/resource"
)

type UnisphereClient struct {
	endpoint string
	auth     string
	token    string
	res      *resource.Resource
	logined  bool

	client *http.Client
}

func NewTransport(insecure bool) *http.Transport {
	return &http.Transport{
		MaxConnsPerHost: 1,
		DialTLSContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: insecure})
		},
	}
}

func NewUnisphereClient(endpoint string, username string, password string, tr *http.Transport) *UnisphereClient {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return &UnisphereClient{
		endpoint: endpoint,
		auth:     auth,
		token:    "",
		client:   &http.Client{Transport: tr},
	}
}

func (_c *UnisphereClient) send(req *http.Request) ([]byte, error) {
	// Variables...
	var resp *http.Response
	var body []byte
	var err error

	// Set Header
	if _c.client.Jar == nil {
		_c.client.Jar, _ = cookiejar.New(nil)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-EMC-REST-CLIENT", "true")
	switch req.Method {
	case "GET":
		req.Header.Add("Authorization", "Basic "+_c.auth)
	case "POST":
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("EMC-CSRF-TOKEN", _c.token)
	}

	// Send Request
	if resp, err = _c.client.Do(req); err != nil {
		return nil, err
	}

	// Read Body
	defer resp.Body.Close()
	if body, err = io.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	// Check StatusCode
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		_c.logined = false
	case http.StatusForbidden:
		return nil, errors.New("forbidden")
	case http.StatusNotFound:
		return nil, errors.New("not found")
	case http.StatusUnprocessableEntity:
		message := gjson.GetBytes(body, "error.messages.0.en-US").String()
		return nil, errors.New(message)
	case http.StatusInternalServerError:
		message := gjson.GetBytes(body, "error.messages.0.en-US").String()
		return nil, errors.New(message)
	}

	// Renew Token
	if resp.Header.Get("EMC-CSRF-TOKEN") != "" {
		_c.token = resp.Header.Get("EMC-CSRF-TOKEN")
		_c.logined = true
	}

	return body, nil
}

func (_c *UnisphereClient) GetInstances(opt *api.UnityActionOptions) ([]gjson.Result, error) {
	var path string
	var req *http.Request
	var body []byte
	var err error
	if opt == nil {
		return nil, errors.New("option is required")
	}
	if path, err = opt.ParseRaw(); err != nil {
		return nil, err
	}

	if req, err = http.NewRequest("GET", _c.endpoint+path, nil); err != nil {
		return nil, err
	}

	if body, err = _c.send(req); err != nil {
		return nil, err
	}

	var data []gjson.Result
	data = gjson.GetBytes(body, "entries.#.content").Array()

	return data, nil
}

func (_c *UnisphereClient) PostMetricRealTimeQuery(opt *api.UnityActionOptions, paths []string, interval time.Duration) (string, error) {
	var path string
	var req *http.Request
	var body []byte
	var err error
	if opt == nil {
		return "", errors.New("option is required")
	}
	if path, err = opt.ParseRaw(); err != nil {
		return "", err
	}

	var reqData struct {
		Paths    []string `json:"paths"`
		Interval int      `json:"interval"`
	}
	reqData.Paths = paths
	reqData.Interval = int(interval / time.Second)
	reqBody, err := json.Marshal(reqData)

	if req, err = http.NewRequest("POST", _c.endpoint+path, bytes.NewBuffer(reqBody)); err != nil {
		return "", err
	}

	if body, err = _c.send(req); err != nil {
		return "", err
	}

	var qid string
	qid = gjson.GetBytes(body, "content.id").String()
	fmt.Println(qid)

	return qid, nil
}
