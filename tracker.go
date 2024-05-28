/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */
package tracker

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(AdobeUsageTracker{})
	httpcaddyfile.RegisterHandlerDirective("adobe_usage_tracker", parseCaddyfile)
}

// AdobeUsageTracker implements an HTTP handler that writes the
// visitor's IP address to a file or stream.
type AdobeUsageTracker struct {
	// The file or stream to write to. Can be "stdout"
	// or "stderr".
	Output string `json:"output,omitempty"`

	w io.Writer
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
	switch m.Output {
	case "stdout":
		m.w = os.Stdout
	case "stderr":
		m.w = os.Stderr
	default:
		return fmt.Errorf("an output stream is required")
	}
	return nil
}

// Validate implements caddy.Validator.
func (m *AdobeUsageTracker) Validate() error {
	if m.w == nil {
		return fmt.Errorf("no writer")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m AdobeUsageTracker) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	msg := fmt.Sprintf("Log upload from %s with %d bytes.\n", r.RemoteAddr, r.ContentLength)
	_, err := m.w.Write([]byte(msg))
	if err != nil {
		return err
	}

	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *AdobeUsageTracker) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	// require an argument
	if !d.NextArg() {
		return d.ArgErr()
	}

	// store the argument
	m.Output = d.Val()
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
