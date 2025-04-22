package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/url"
	"unisphere_exporter/collector"
)

func ProbeUnisphere(ctx context.Context, target string, registry *prometheus.Registry, hc *http.Client) (bool, error) {
	tgt, err := url.Parse(target)
	if err != nil {
		return false, fmt.Errorf("url.Parse failed: %v", err)
	}

	if tgt.Scheme != "https" && tgt.Scheme != "http" {
		return false, fmt.Errorf("Unsupported scheme %q", tgt.Scheme)
	}

	// Filter anything else than scheme and hostname
	u := url.URL{
		Scheme: tgt.Scheme,
		Host:   tgt.Host,
	}
	c, err := newUnisphereClient(ctx, u, hc)
	if err != nil {
		return false, err
	}

	// TODO: Make parallel
	fmt.Println(c)
	success := collector.ProbeMetric(c, registry)

	return success, nil
}
