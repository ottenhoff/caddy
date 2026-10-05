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

package reverseproxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLongsightLegacyCookieSelection(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cookies    []*http.Cookie
		unhealthy  bool
		want       int
		wantCookie bool
	}{
		{
			name:    "JSESSIONID routes to legacy node",
			cookies: []*http.Cookie{{Name: "JSESSIONID", Value: "session.sakai01"}},
			want:    1,
		},
		{
			name:    "SAKAIID takes precedence",
			cookies: []*http.Cookie{{Name: "SAKAIID", Value: "session.sakai02"}, {Name: "JSESSIONID", Value: "session.sakai01"}},
			want:    2,
		},
		{
			name:       "unknown legacy node uses fallback",
			cookies:    []*http.Cookie{{Name: "JSESSIONID", Value: "session.sakai99"}},
			wantCookie: true,
		},
		{
			name:       "unhealthy legacy node uses fallback",
			cookies:    []*http.Cookie{{Name: "JSESSIONID", Value: "session.sakai01"}},
			unhealthy:  true,
			wantCookie: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := CookieHashSelection{Name: "lb", fallback: FirstSelection{}}
			pool := testPool()
			pool[0].Dial = "10.0.0.1:8080"
			pool[1].Dial = "10.0.65.179:8080"
			pool[2].Dial = "10.0.67.12:8080"
			pool[1].setHealthy(!tc.unhealthy)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, cookie := range tc.cookies {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			if got := policy.Select(pool, req, w); got != pool[tc.want] {
				t.Fatalf("Select() = %v, want %v", got, pool[tc.want])
			}
			if got := len(w.Result().Cookies()) > 0; got != tc.wantCookie {
				t.Errorf("response sets cookie = %v, want %v", got, tc.wantCookie)
			}
		})
	}
}
