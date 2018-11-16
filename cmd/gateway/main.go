package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/logging"
	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	gatewayConfig "github.com/skygeario/skygear-server/pkg/gateway/config"
	pqStore "github.com/skygeario/skygear-server/pkg/gateway/db/pq"
	"github.com/skygeario/skygear-server/pkg/gateway/middleware"
	"github.com/skygeario/skygear-server/pkg/gateway/provider"
)

var config gatewayConfig.Configuration

func init() {
	// logging initialization
	logging.SetModule("gateway")

	logger := logging.LoggerEntry("gateway")
	if err := config.ReadFromEnv(); err != nil {
		logger.WithError(err).Panic(
			"Fail to load config for starting gateway server")
	}

	logger.WithField("config", config).Debug("Gateway config")
}

func main() {
	logger := logging.LoggerEntry("gateway")

	// create gateway store
	store, connErr := pqStore.NewGatewayStore(
		context.Background(),
		config.DB.ConnectionStr,
	)
	if connErr != nil {
		logger.WithError(connErr).Panic("Fail to create db conn")
	}
	defer store.Close()

	r := mux.NewRouter()
	r.HandleFunc("/healthz", HealthCheckHandler)

	proxy := NewReverseProxy()
	gr := r.PathPrefix("/{gear}").Subrouter()

	// RecoverMiddleware must come first
	gr.Use(coreMiddleware.RecoverMiddleware{}.Handle)
	// TODO:
	// Currently both config and authz middleware both query store to get
	// app, see how to reduce query to optimize the performance
	gr.Use(coreMiddleware.TenantConfigurationMiddleware{
		ConfigurationProvider: provider.GatewayTenantConfigurationProvider{
			Store: store,
		},
	}.Handle)
	gr.Use(middleware.TenantAuthzMiddleware{
		Store: store,
	}.Handle)

	gr.HandleFunc("/{rest:.*}", rewriteHandler(proxy))

	srv := &http.Server{
		Addr: config.HTTP.Host,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	logger.Info("Start gateway server")
	if err := srv.ListenAndServe(); err != nil {
		logger.Errorf("Fail to start gateway server %v", err)
	}
}

// NewReverseProxy takes an incoming request and sends it to coresponding
// gear server
func NewReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		path := req.URL.Path
		req.URL = config.Router.GetRouterMap()[req.Header.Get("X-Skygear-Gear")]
		req.URL.Path = path
	}
	return &httputil.ReverseProxy{Director: director}
}

func rewriteHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Skygear-Gear", mux.Vars(r)["gear"])
		r.URL.Path = "/" + mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}

// HealthCheckHandler is basic handler for server health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "OK")
}