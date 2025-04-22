package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"log"
	"net/http"
	"net/url"
	"os/user"
	"runtime"
	"time"
	"unisphere_exporter/client"
	"unisphere_exporter/collector"
	"unisphere_exporter/utils"
)

type Auth struct {
	User     string
	Password string
}

var authMap map[string]Auth

func newUnisphereClient(ctx context.Context, tgt url.URL, hc *http.Client) (utils.UnisphereHTTP, error) {
	auth, ok := authMap[tgt.String()]
	if !ok {
		return nil, fmt.Errorf("No API authentication registered for %q", tgt.String())
	}

	if auth.User != "" && auth.Password != "" {
		_, err := newUnispherePasswordClient(ctx, tgt, hc, auth.User, auth.Password)

		fmt.Println()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, fmt.Errorf("Invalid authentication data for %q", tgt.String())
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	roots, err := x509.SystemCertPool()
	tc := &tls.Config{RootCAs: roots}
	tr := &http.Transport{TLSClientConfig: tc}
	params := r.URL.Query()
	target := params.Get("target")
	if target == "" {
		http.Error(w, "Target parameter missing or empty", http.StatusBadRequest)
		return
	}
	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Whether or not the probe succeeded",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_duration_seconds",
		Help: "How many seconds the probe took to complete",
	})
	timeoutSeconds := 30
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	start := time.Now()
	success, err := ProbeUnisphere(ctx, target, registry, &http.Client{Transport: tr})
	if err != nil {
		log.Printf("Probe request rejected; error is: %v", err)
		http.Error(w, fmt.Sprintf("probe: %v", err), http.StatusBadRequest)
		return
	}
	duration := time.Since(start).Seconds()
	probeDurationGauge.Set(duration)
	if success {
		probeSuccessGauge.Set(1)
		log.Printf("Probe of %q succeeded, took %.3f seconds", target, duration)
	} else {
		// probeSuccessGauge default is 0
		log.Printf("Probe of %q failed, took %.3f seconds", target, duration)
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	var (
		configFile = kingpin.Flag("config.file", "file containing the authentication map to use when connecting to a Unisphere device").Default("config.yml").String()
		listen     = kingpin.Flag("web.listen-address", "Addresses on which to expose metrics and web interface.").Default(":9182").String()
		maxProcs   = kingpin.Flag("runtime.gomaxprocs", "The target number of CPUs Go will run on (GOMAXPROCS)").Envar("GOMAXPROCS").Default("1").Int()
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("unisphere_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)
	logger.Info("Starting unisphere_exporter", "version", version.Info())
	logger.Info("Build contex", "build_context", version.BuildContext())

	if u, err := user.Current(); err == nil && u.Uid == "0" {
		logger.Warn("Unisphere Exporter is running as root user. This exporter is designed to run as unprivileged user, root is not required.")
	}
	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	client.SetModule(configFile, logger)

	logger.Info("Unisphere exporter running.", "listen_port", *listen)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		reg := prometheus.NewRegistry()
		collector.Probe(w, r, logger, reg)
	})

	http.ListenAndServe(*listen, nil)

	//roots, err := x509.SystemCertPool()
	//if err != nil {
	//	log.Fatalf("Unable to fetch system CA store: %v", err)
	//}
	//if *extraCAs != "" {
	//	certs, err := os.ReadFile(*extraCAs)
	//	if err != nil {
	//		log.Fatalf("Failed to read extra CA file: %v", err)
	//	}
	//
	//	if ok := roots.AppendCertsFromPEM(certs); !ok {
	//		log.Fatalf("Failed to append certs from PEM, unknown error")
	//	}
	//}
	//tc := &tls.Config{RootCAs: roots}
	//var insecure bool
	//insecure = true
	////if insecure {
	////	tc.InsecureSkipVerify = true
	////}
	////tr := &http.Transport{TLSClientConfig: tc}
	//
	//http.Handle("/metrics", promhttp.Handler())
	//http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
	//	probeHandler(w, r, tr)
	//})
	//http.ListenAndServe(*listen, nil)
	//log.Printf("Unisphere exporter running, listening on %q", *listen)
}
