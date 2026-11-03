package main

import (
	"flag"
	"net/http"
	"time"
)

type serverOptions struct {
	configPath        string
	address           string
	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
}

func (o *serverOptions) registerFlags(flags *flag.FlagSet) {
	flags.StringVar(&o.configPath, "config", "config.toml", "Path to the TOML configuration file")
	flags.StringVar(&o.address, "addr", "127.0.0.1:8080", "HTTP listen address")

	flags.DurationVar(&o.readHeaderTimeout, "read-header-timeout", 5*time.Second, "Maximum time to read request headers")
	flags.DurationVar(&o.readTimeout, "read-timeout", 10*time.Second, "Maximum time to read a complete request")
	flags.DurationVar(&o.writeTimeout, "write-timeout", 10*time.Second, "Maximum time to write a response")
	flags.DurationVar(&o.idleTimeout, "idle-timeout", 60*time.Second, "Maximum time to wait for the next keep-alive request")
}

func (o serverOptions) newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              o.address,
		Handler:           handler,
		ReadHeaderTimeout: o.readHeaderTimeout,
		ReadTimeout:       o.readTimeout,
		WriteTimeout:      o.writeTimeout,
		IdleTimeout:       o.idleTimeout,
	}
}
