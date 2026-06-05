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
	"bytes"
	"io"
	"testing"

	"github.com/molecule-man/go-brrr"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestBrotliProvisionDefaultLevel(t *testing.T) {
	b := new(Brotli)
	err := b.Provision(caddy.Context{})
	if err != nil {
		t.Fatalf("Provision() error = %v", err)
	}
	if b.Level == nil {
		t.Fatal("Provision() did not set default level")
	}
	if *b.Level != defaultBrotliLevel {
		t.Fatalf("Provision() level = %d, want %d", *b.Level, defaultBrotliLevel)
	}
}

func TestBrotliValidateLevel(t *testing.T) {
	for i, tc := range []struct {
		level   *int
		wantErr bool
	}{
		{level: nil, wantErr: false},
		{level: levelPtr(brrr.BestSpeed), wantErr: false},
		{level: levelPtr(defaultBrotliLevel), wantErr: false},
		{level: levelPtr(brrr.BestCompression), wantErr: false},
		{level: levelPtr(brrr.BestSpeed - 1), wantErr: true},
		{level: levelPtr(brrr.BestCompression + 1), wantErr: true},
	} {
		b := Brotli{Level: tc.level}
		err := b.Provision(caddy.Context{})
		if err != nil {
			t.Errorf("Test %d: Provision() error = %v", i, err)
			continue
		}
		err = b.Validate()
		if (err != nil) != tc.wantErr {
			t.Errorf("Test %d: Validate() error = %v, wantErr = %v", i, err, tc.wantErr)
		}
	}
}

func TestBrotliUnmarshalCaddyfile(t *testing.T) {
	for i, tc := range []struct {
		input     string
		wantLevel *int
		wantErr   bool
	}{
		{input: "br", wantLevel: nil, wantErr: false},
		{input: "br 0", wantLevel: levelPtr(0), wantErr: false},
		{input: "br 5", wantLevel: levelPtr(5), wantErr: false},
		{input: "br 11", wantLevel: levelPtr(11), wantErr: false},
		{input: "br 5 extra", wantLevel: nil, wantErr: true},
		{input: "br nope", wantLevel: nil, wantErr: true},
	} {
		var b Brotli
		err := b.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tc.input))
		if (err != nil) != tc.wantErr {
			t.Errorf("Test %d: UnmarshalCaddyfile() error = %v, wantErr = %v", i, err, tc.wantErr)
			continue
		}
		if tc.wantErr {
			continue
		}
		if b.Level == nil || tc.wantLevel == nil {
			if b.Level != tc.wantLevel {
				t.Errorf("Test %d: UnmarshalCaddyfile() level = %v, want %v", i, b.Level, tc.wantLevel)
			}
			continue
		}
		if *b.Level != *tc.wantLevel {
			t.Errorf("Test %d: UnmarshalCaddyfile() level = %d, want %d", i, *b.Level, *tc.wantLevel)
		}
	}
}

func TestBrotliNewEncoderRoundTrip(t *testing.T) {
	var compressed bytes.Buffer
	original := []byte("Every site on HTTPS. Every site on HTTPS. Every site on HTTPS.")

	b := Brotli{Level: levelPtr(defaultBrotliLevel)}
	err := b.Provision(caddy.Context{})
	if err != nil {
		t.Fatalf("Provision() error = %v", err)
	}
	encoder := b.NewEncoder()
	encoder.Reset(&compressed)
	_, err = encoder.Write(original)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	err = encoder.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reader := brrr.NewReader(&compressed)
	defer reader.Close()
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if !bytes.Equal(decompressed, original) {
		t.Fatalf("round trip = %q, want %q", decompressed, original)
	}
}

func levelPtr(level int) *int {
	return &level
}
