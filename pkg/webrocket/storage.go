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
	"github.com/nu7hatch/persival"
	"os"
	"path"
	"regexp"
)

// _vhost is an internal struct to represent stored information about
// the vhost.
type _vhost struct {
	// The vhost's name.
	Path string
	// The vhost's access token.
	AccessToken string
}

// _channel is an internal struct to represent stored information about
// the channel.
type _channel struct {
	// ID of the vhost.
	Vhost int
	// The channel's name.
	Name string
	// The channel's type.
	Kind ChannelType
}

// _permission is an internal struct to represent stored information about
// the permission.
type _permission struct {
	// ID of the vhost.
	Vhost int
	// The permission's user id.
	Uid string
	// The permission's pattern.
	Pattern *regexp.Regexp
	// The permission's token.
	Token string
}

// Initializer.
func init() {
	persival.Register(&_vhost{})
	persival.Register(&_channel{})
	persival.Register(&_permission{})
}

// storage implements an adapter for the persistence layer. At the moment
// a Persival database is used to provide data persistency. All storage's
// functions are threadsafe.
type storage struct {
	// The vhosts bucket.
	vhosts *persival.Bucket
	// The channels bucket.
	channels *persival.Bucket
	// The permissions bucket.
	permissions *persival.Bucket
	// Path to storage directory.
	dir string
}

// Internal constructor
// -----------------------------------------------------------------------------

// newStorage creates new persistence under the specified directory.
// At the moment Kyoto Cabinet is used, so it creates a 'webrocket.kch'
// database file there.
//
// dir - A path to the storage location.
//
// Returns configured storage or error if something went wrong
func newStorage(dir string, name string) (s *storage, err error) {
	if err = os.MkdirAll(dir, 0744); err != nil {
		return nil, err
	}
	s = &storage{dir: dir}
	// Initialize all the buckets...
	s.vhosts, err = persival.NewBucket(path.Join(dir, name+".vhosts.bkt"), 0)
	if err != nil {
		return nil, err
	}
	s.channels, err = persival.NewBucket(path.Join(dir, name+".channels.bkt"), 0)
	if err != nil {
		return nil, err
	}
	s.permissions, err = persival.NewBucket(path.Join(dir, name+".permissions.bkt"), 0)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Exported
// -----------------------------------------------------------------------------

// Load reads all the webrocket data from the storage and configures given
// context with loaded information.
//
// ctx - The context to be configured.
//
func (s *storage) Load(ctx *Context) {
	var vhosts = make(map[int]*Vhost)
	for k, val := range s.vhosts.All() {
		if v, ok := val.(*_vhost); ok {
			if x, err := ctx.AddVhost(v.Path); err == nil {
				x.accessToken = v.AccessToken
				x._id = k
				vhosts[k] = x
			}
		}
	}
	for k, val := range s.channels.All() {
		if ch, ok := val.(*_channel); ok {
			if v, ok := vhosts[ch.Vhost]; ok {
				x, _ := newChannel(ch.Name, ChannelType(ch.Kind))
				x._id = k
				v.channels[ch.Name] = x
			} else {
				s.channels.Delete(k)
			}
		}
	}
	for k, val := range s.permissions.All() {
		if p, ok := val.(*_permission); ok {
			if v, ok := vhosts[p.Vhost]; ok {
				x := &Permission{k, p.Uid, p.Pattern, p.Token}
				v.permissions[p.Token] = x
			}
		}
	}
}

// AddVhost creates a databse entry for the specified vhost.
//
// vhost - The vhost to be added.
//
// Returns an error if something went wrong.
func (s *storage) AddVhost(vhost *Vhost) (err error) {
	vhost._id, err = s.vhosts.Set(&_vhost{vhost.path, vhost.accessToken})
	return
}

// UpdateVhost changes information about the specified vhost.
//
// vhost - The vhost to be changed.
//
// Returns an error if something went wrong.
func (s *storage) UpdateVhost(vhost *Vhost) (err error) {
	err = s.vhosts.Update(vhost._id, &_vhost{vhost.path, vhost.accessToken})
	return
}

// DeleteVhost removes a database entry for the specified vhost and removes
// all its channels' entries as well.
//
// vhost - The vhost to be deleted.
//
// Returns an error if something went wrong.
func (s *storage) DeleteVhost(vhost *Vhost) (err error) {
	for _, channel := range vhost.Channels() {
		s.channels.Delete(channel._id)
	}
	for _, permission := range vhost.Permissions() {
		s.permissions.Delete(permission._id)
	}
	err = s.vhosts.Delete(vhost._id)
	return
}

// AddChannel create a databse entry for the specified channel.
//
// vhost   - The channel's parent vhost.
// channel - The channel to be added.
//
// Returns an error if something went wrong.
func (s *storage) AddChannel(vhost *Vhost, channel *Channel) (err error) {
	channel._id, err = s.channels.Set(&_channel{vhost._id, channel.name, channel.kind})
	return
}

// DeleteChannel removes a given channel's entry from the database.
//
// vhost   - The channel's parent vhost.
// channel - The channel to be deleted.
//
// Returns an error if something went wrong.
func (s *storage) DeleteChannel(channel *Channel) (err error) {
	err = s.channels.Delete(channel._id)
	return
}

// AddPermission create a databse entry for the specified permission.
//
// vhost      - The permission's parent vhost.
// permission - The permission to be added.
//
// Returns an error if something went wrong.
func (s *storage) AddPermission(vhost *Vhost, perm *Permission) (err error) {
	perm._id, err = s.permissions.Set(&_permission{vhost._id, perm.uid, perm.pattern, perm.token})
	return
}

// DeletePermission removes a given permission's entry from the database.
//
// vhost      - The permission's parent vhost.
// permission - The permission to be deleted.
//
// Returns an error if something went wrong.
func (s *storage) DeletePermission(permission *Permission) (err error) {
	err = s.permissions.Delete(permission._id)
	return
}

// Clear truncates all the data in the storage.
//
// Returns an error if something went wrong.
func (s *storage) Clear() (err error) {
	if err = s.vhosts.Destroy(); err != nil {
		return
	}
	if err = s.channels.Destroy(); err != nil {
		return
	}
	if err = s.permissions.Destroy(); err != nil {
		return
	}
	return nil
}

// Save writes down and synchronizes all the data.
//
// Returns an error if something went wrong.
func (s *storage) Save() error {
	// ... nothing to do after switching to Persival.
	return nil
}

// Kill closes the storage.
func (s *storage) Kill() {
	s.vhosts.Close()
	s.channels.Close()
	s.permissions.Close()
}
