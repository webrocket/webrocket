// Copyright (C) 2011 by Krzysztof Kowalik <chris@nu7hat.ch>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package engine

import (
	"net/http"
	"strings"
)

// admnHandlerFunc is a shorthand for the admin interface's handler function.
type adminHandlerFunc func(*Context, http.ResponseWriter, *http.Request) (int, error)

// AdminServeMux is a simple wrapper for map of admin handler functions,
// which allows to effectively search for the handlers.
type AdminServeMux map[string]adminHandlerFunc

// Exported
// -----------------------------------------------------------------------------

// Match looks for the handler matching given request method and path.
//
// method - A HTTP request method.
// path   - A path to be found.
//
// Returns matching handler function and status.
func (mux AdminServeMux) Match(method, path string) (adminHandlerFunc, bool) {
	path = strings.Split(path, "?")[0]
	handler, ok := mux[method+" "+path]
	return handler, ok
}
