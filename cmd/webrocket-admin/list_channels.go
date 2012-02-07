package main

import (
	"errors"
	"fmt"
	"sort"
)

func listChannels(params []string) (err error, ok bool) {
	var vhost string
	var entries []interface{}
	var res *Response
	if vhost, ok = vhostParams(params); !ok {
		return
	}
	res, err = performRequest("GET", "/channels", "channels", map[string]string{
		"vhost": vhost,
	})
	if err != nil {
		return
	}
	if entries, ok = res.Data.([]interface{}); !ok {
		err = errors.New("couldn't list channels, invalid response")
		return
	}
	names, channels := make([]string, len(entries)), make([]*Channel, len(entries))
	for i, x := range entries {
		if channel, ok := maybeChannel(x); ok {
			names[i] = channel.Name
			channels[i] = channel
		}
	}
	sort.Strings(names)
	for i, name := range names {
		fmt.Printf("%s\t(%d subscribers)\n", name, channels[i].SubscribersSize)
	}
	return
}
