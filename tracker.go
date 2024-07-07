/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

// Package tracker provides the caddy adobe_usage_tracker plugin.
package tracker

import (
	"bytes"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func init() {
	caddy.RegisterModule(AdobeUsageTracker{})
	httpcaddyfile.RegisterHandlerDirective("adobe_usage_tracker", parseCaddyfile)
}

// AdobeUsageTracker implements HTTP middleware that parses
// uploaded log files from Adobe desktop applications in order to
// collect measurements about past launches. These measurements
// are then uploaded to an InfluxDB (using the v1 HTTP API).
//
// Configuration of the tracker requires four parameters:
//
// - the endpoint URL of the influx v1 upload api
// - the name of the influx v1 database
// - the retention policy of the influx v1 database
// - an API token authorized for writes of the database
//
// Note: this middleware uses the v1 HTTP write API because it's
// fully supported by both v1 and v3 databases.  When using a
// v3 database, you must specify a "dbrp" mapping from the
// database and policy names to the specific bucket you want
// uploads to go to. See the influx docs for details:
//
// https://docs.influxdata.com/influxdb/cloud-serverless/write-data/api/v1-http/
type AdobeUsageTracker struct {
	Endpoint string `json:"endpoint,omitempty"`
	Database string `json:"database,omitempty"`
	Policy   string `json:"policy,omitempty"`
	Token    string `json:"token,omitempty"`
	Header   string `json:"header,omitempty"`
	Position string `json:"position,omitempty"`

	ep  string
	db  string
	rp  string
	tok string
	hdr string
	pos string
}

// CaddyModule returns the Caddy module information.
func (AdobeUsageTracker) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.adobe_usage_tracker",
		New: func() caddy.Module { return new(AdobeUsageTracker) },
	}
}

// Provision implements caddy.Provisioner.
func (m *AdobeUsageTracker) Provision(caddy.Context) error {
	if m.Endpoint == "" {
		return fmt.Errorf("an endpoint URL must be specified")
	}
	u, err := url.Parse(m.Endpoint)
	if err != nil {
		return fmt.Errorf("%q is not a valid endpoint url: %v", m.Endpoint, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("endpoint protocol must be https, not '%s'", u.Scheme)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("endpoint %q is missing a hostname", m.Endpoint)
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("endpoint %q cannot have a path, query, or fragment portion", m.Endpoint)
	}
	m.ep = m.Endpoint
	if m.Database == "" {
		return fmt.Errorf("database must be specified")
	}
	m.db = m.Database
	if m.Policy == "" {
		return fmt.Errorf("A retention policy must be specified")
	}
	m.rp = m.Policy
	if m.Token == "" {
		return fmt.Errorf("A token must be specified")
	}
	m.tok = m.Token
	m.hdr = m.Header
	switch strings.ToLower(m.Position) {
	case "first":
		m.pos = "first"
	case "last":
		m.pos = "last"
	default:
		return fmt.Errorf("Position must be \"first\" or \"last\", found %q", m.Position)
	}
	return nil
}

// Validate implements caddy.Validator.
func (m *AdobeUsageTracker) Validate() error {
	if m.ep == "" {
		return fmt.Errorf("endpoint URL must be specified")
	}
	u, err := url.Parse(m.ep)
	if err != nil {
		return fmt.Errorf("%q is not a valid endpoint URL: %v", m.ep, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("endpoint protocol must be https, not '%s'", u.Scheme)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("endpoint %q is missing a hostname", m.ep)
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("endpoint %q cannot have a path, query, or fragment portion", m.ep)
	}
	if m.db == "" {
		return fmt.Errorf("database must be specified")
	}
	if m.rp == "" {
		return fmt.Errorf("retention policy must be specified")
	}
	if m.tok == "" {
		return fmt.Errorf("token must be specified")
	}
	if m.pos != "first" && m.pos != "last" {
		return fmt.Errorf("position must be \"first\" or \"last\"")
	}
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *AdobeUsageTracker) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	// set default values
	m.Header = "X-Forwarded-For"
	m.Position = "first"
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		key := d.Val()
		if !d.NextArg() {
			return d.ArgErr()
		}
		switch key {
		case "endpoint":
			m.Endpoint = d.Val()
		case "database":
			m.Database = d.Val()
		case "policy":
			m.Policy = d.Val()
		case "token":
			m.Token = d.Val()
		case "header":
			m.Header = d.Val()
		case "position":
			m.Position = d.Val()
		default:
			return d.ArgErr()
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new AdobeUsageTracker.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m AdobeUsageTracker
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// ServeHTTP implements caddyhttp.MiddlewareHandler. It extracts
// measurements from any logs uploaded in the request, sends them
// to the influxDB endpoint, and then passes the request intact
// onto the next handler.
func (m AdobeUsageTracker) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	logger := caddy.Log()
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	remoteAddr := m.parseRemoteAddr(r, logger)
	sessions := parseLog(string(buf), remoteAddr)
	userAgent, err := url.QueryUnescape(r.UserAgent())
	if err != nil {
		userAgent = r.UserAgent()
	}
	logger.Info("AdobeUsageTracker: incoming request summary",
		zap.String("remote-address", remoteAddr),
		zap.String("user-agent", userAgent),
		zap.Int("content-length", len(buf)),
		zap.Int("session-count", len(sessions)),
	)
	logger.Debug("AdobeUsageTracker: uploading sessions", zap.Objects("sessions", sessions))
	if len(sessions) == 0 {
		logger.Info("AdobeUsageTracker: no sessions to upload")
	} else {
		err = sendSessions(m.ep, m.db, m.rp, m.tok, sessions, logger)
		if err != nil {
			logger.Error("AdobeUsageTracker: failed to send sessions", zap.Error(err))
		} else {
			logger.Info("AdobeUsageTracker: sent sessions successfully")
		}
	}
	r.Body = io.NopCloser(bytes.NewReader(buf))
	return next.ServeHTTP(w, r)
}

// / parseRemoteAddr consults the request headers and determines
// / the remote address of the actual client doing the upload
func (m *AdobeUsageTracker) parseRemoteAddr(r *http.Request, l *zap.Logger) string {
	remoteHost, remotePort, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost, remotePort, _ = strings.Cut(r.RemoteAddr, ":")
	}
	if m.hdr == "" {
		l.Debug("AdobeUsageTracker: per config, ignoring headers",
			zap.String("remote-address", remoteHost),
			zap.String("remote-port", remotePort))
		return remoteHost
	}
	header := r.Header.Get(m.hdr)
	if header == "" {
		l.Warn("AdobeUsageTracker: header not found",
			zap.String("header_name", m.hdr),
			zap.String("remote-address", remoteHost),
			zap.String("remote-port", remotePort))
		return remoteHost
	}
	l.Debug("AdobeUsageTracker: found header",
		zap.String("header_name", m.hdr),
		zap.String("header-value", header))
	parts := strings.Split(header, ",")
	address := strings.Trim(parts[0], " ")
	if len(parts) > 1 && m.pos == "last" {
		last := strings.Trim(parts[len(parts)-1], " ")
		if last == "" && len(parts) > 2 {
			last = strings.Trim(parts[len(parts)-2], " ")
		}
		if last != "" {
			address = last
		}
	}
	if address == "" {
		l.Warn("AdobeUsageTracker: no address found in header",
			zap.String("header_name", m.hdr),
			zap.String("header_value", header),
			zap.String("remote-address", remoteHost),
			zap.String("remote-port", remotePort))
		return remoteHost
	}
	return address
}

// Interface guards
var (
	_ caddy.Provisioner           = (*AdobeUsageTracker)(nil)
	_ caddy.Validator             = (*AdobeUsageTracker)(nil)
	_ caddyhttp.MiddlewareHandler = (*AdobeUsageTracker)(nil)
	_ caddyfile.Unmarshaler       = (*AdobeUsageTracker)(nil)
)
