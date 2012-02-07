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

package webrocket

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

// The version information.
const (
	VerMajor = 0
	VerMinor = 3
	VerPatch = 0
)

// Version returns current version of the package.
func Version() string {
	return fmt.Sprintf("%d.%d.%d", VerMajor, VerMinor, VerPatch)
}

// Status represents client response statuses.
type Status struct {
	Status string
	Code   int
}

// Status returns full status message.
func (s *Status) String() string {
	return fmt.Sprintf("%d %s", s.Code, s.Status)
}

// Map returns status as a map.
func (s *Status) Map() map[string]interface{} {
	return map[string]interface{}{
		"code":   s.Code,
		"status": s.Status,
	}
}

// DefaultNodeName discovers name of the node from the host name configured
// in the operating system. Basically the result of the 'uname -n' command
// will be returned.
func DefaultNodeName() string {
	x := exec.Command("uname", "-n")
	node, err := x.Output()
	if err != nil {
		panic("can't get node name: " + err.Error())
	}
	return strings.TrimSpace(string(node))
}

// ReadCookie reads cookie string from the node's cookie file.
//
// Example:
//
//    node := webrocket.DefaultNodeName()
//    cookie := webrocket.ReadCookie(node)
//
// Returns cookie string.
func ReadCookie(node string) string {
	cookiePath := "/var/lib/webrocket/" + node + ".cookie"
	data, err := ioutil.ReadFile(cookiePath)
	if err != nil {
		return ""
	}
	return string(data)
}
