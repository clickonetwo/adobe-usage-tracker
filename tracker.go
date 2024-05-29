/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */
package tracker

import (
	"bytes"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"io"
	"net/http"
	"net/url"
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
// - the endpoint URL of the InfluxDB
// - the name of the database
// - the retention policy of the database
// - an API token authorized for writes of the database
type AdobeUsageTracker struct {
	endpoint string `json:"endpoint,omitempty"`
	database string `json:"database,omitempty"`
	policy   string `json:"policy,omitempty"`
	token    string `json:"token,omitempty"`

	ep  string
	db  string
	pol string
	tok string
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
	if m.endpoint == "" {
		return fmt.Errorf("An endpoint URL must be specified")
	}
	u, err := url.Parse(m.endpoint)
	if err != nil {
		return fmt.Errorf("%q is not a valid endpoint url: %v", m.endpoint, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("The endpoint protocol must be https, not '%s'", u.Scheme)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("The endpoint %q is missing a hostname", m.endpoint)
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("The endpoint %q cannot have a path, query, or fragment portion", m.endpoint)
	}
	m.ep = m.endpoint
	if m.database == "" {
		return fmt.Errorf("A database must be specified")
	}
	m.db = m.database
	if m.policy == "" {
		return fmt.Errorf("A retention policy must be specified")
	}
	m.pol = m.policy
	if m.token == "" {
		return fmt.Errorf("A token must be specified")
	}
	m.tok = m.token
	return nil
}

// Validate implements caddy.Validator.
func (m *AdobeUsageTracker) Validate() error {
	if m.ep == "" {
		return fmt.Errorf("An endpoint URL must be specified")
	}
	u, err := url.Parse(m.ep)
	if err != nil {
		return fmt.Errorf("%q is not a valid endpoint URL: %v", m.ep, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("The endpoint protocol must be https, not '%s'", u.Scheme)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("The endpoint %q is missing a hostname", m.ep)
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("The endpoint %q cannot have a path, query, or fragment portion", m.ep)
	}
	if m.db == "" {
		return fmt.Errorf("A database must be specified")
	}
	if m.pol == "" {
		return fmt.Errorf("A retention policy must be specified")
	}
	if m.tok == "" {
		return fmt.Errorf("A token must be specified")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m AdobeUsageTracker) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	sessions := parseLog(string(b))
	if len(sessions) > 0 {
		go sendSessions(m.ep, m.db, m.pol, m.tok, sessions)
	}
	r.Body = io.NopCloser(bytes.NewReader(b))
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *AdobeUsageTracker) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		key := d.Val()
		if !d.NextArg() {
			return d.ArgErr()
		}
		switch key {
		case "endpoint":
			m.endpoint = d.Val()
		case "database":
			m.database = d.Val()
		case "policy":
			m.policy = d.Val()
		case "token":
			m.token = d.Val()
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

// Interface guards
var (
	_ caddy.Provisioner           = (*AdobeUsageTracker)(nil)
	_ caddy.Validator             = (*AdobeUsageTracker)(nil)
	_ caddyhttp.MiddlewareHandler = (*AdobeUsageTracker)(nil)
	_ caddyfile.Unmarshaler       = (*AdobeUsageTracker)(nil)
)
