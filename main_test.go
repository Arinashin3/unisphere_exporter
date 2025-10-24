package main

import (
	"context"
	"fmt"
	"testing"
	"time"
	"unisphere_exporter/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestName(t *testing.T) {
	ctx := context.Background()
	var attrs []attribute.KeyValue
	v := "value1"
	res, _ := resource.New(ctx, resource.WithDetectors(
		resource.StringDetector("", attribute.Key("test"), func() (string, error) {
			return v, nil
		}),
	))

	fmt.Println(mp)
	res.String()

}
