package main

import (
	"errors"
	"fmt"
	"sort"
)

func listVhosts(params []string) (err error, ok bool) {
	var vhosts []interface{}
	var res *Response
	res, err = performRequest("GET", "/vhosts", "vhosts", map[string]string{})
	if err != nil {
		return
	}
	if vhosts, ok = res.Data.([]interface{}); !ok {
		err = errors.New("couldn't list vhosts, invalid response")
		return
	}
	paths, i := make([]string, len(vhosts)), 0
	for _, x := range vhosts {
		if vhost, ok := maybeVhost(x); ok {
			paths[i] = vhost.Path
			i += 1
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%s\n", path)
	}
	return
}
