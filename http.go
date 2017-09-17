// Copyright 2017 John Scherff
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

package goutil

import (
	`fmt`
	`net/http`
)

// AllowedMethodHandler restricts http requests to methods provided.
func AllowedMethodHandler(h http.Handler, methods ...string) http.Handler {

	return http.HandlerFunc(

		func(w http.ResponseWriter, r *http.Request) {

			for _, m := range methods {
				if r.Method == m {
					h.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, fmt.Sprintf(`Unsupported method %q`, r.Method),
				http.StatusMethodNotAllowed)
		},
	)
}
