package main

import "fmt"

func showVhost(params []string) (err error, ok bool) {
	var path string
	var res *Response
	if path, ok = vhostParams(params); !ok {
		return
	}
	res, err = performRequest("GET", path, "vhost")
	if err != nil {
		return
	}
	if vhost, ok := maybeVhost(res.Data); ok {
		fmt.Printf("%s\n%s\n", vhost.Path, vhost.AccessToken)
	}
	return
}
