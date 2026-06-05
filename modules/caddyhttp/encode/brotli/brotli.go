// Copyright 2015 Matthew Holt and The Caddy Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package caddybrotli

import (
	"fmt"
	"strconv"

	"github.com/molecule-man/go-brrr"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/encode"
)

func init() {
	caddy.RegisterModule(Brotli{})
}

// Brotli can create Brotli encoders.
type Brotli struct {
	// The compression level. Accepted values are 0 through 11.
	Level *int `json:"level,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (Brotli) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.encoders.br",
		New: func() caddy.Module { return new(Brotli) },
	}
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens.
func (b *Brotli) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume option name
	args := d.RemainingArgs()
	switch len(args) {
	case 0:
	case 1:
		level, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		b.Level = &level
	default:
		return d.ArgErr()
	}
	return nil
}

// Provision provisions b's configuration.
func (b *Brotli) Provision(ctx caddy.Context) error {
	if b.Level == nil {
		level := defaultBrotliLevel
		b.Level = &level
	}
	return nil
}

// Validate validates b's configuration.
func (b Brotli) Validate() error {
	if *b.Level < brrr.BestSpeed {
		return fmt.Errorf("quality too low; must be >= %d", brrr.BestSpeed)
	}
	if *b.Level > brrr.BestCompression {
		return fmt.Errorf("quality too high; must be <= %d", brrr.BestCompression)
	}
	return nil
}

// AcceptEncoding returns the name of the encoding as
// used in the Accept-Encoding request headers.
func (Brotli) AcceptEncoding() string { return "br" }

// NewEncoder returns a new Brotli writer.
func (b Brotli) NewEncoder() encode.Encoder {
	writer, _ := brrr.NewWriter(nil, *b.Level)
	return writer
}

var defaultBrotliLevel = 5

// Interface guards
var (
	_ encode.Encoding       = (*Brotli)(nil)
	_ caddy.Provisioner     = (*Brotli)(nil)
	_ caddy.Validator       = (*Brotli)(nil)
	_ caddyfile.Unmarshaler = (*Brotli)(nil)
)
