package main

import "fmt"

func vhostParams(params []string) (path string, ok bool) {
	if len(params) == 1 && params[0] != "" {
		ok, path = true, params[0]
	}
	return
}

func addVhost(params []string) (err error, ok bool) {
	var path string
	var res *Response
	if path, ok = vhostParams(params); !ok {
		return
	}
	res, err = performRequest("POST", path, "vhost")
	if err != nil {
		return
	}
	if vhost, ok := maybeVhost(res.Data); ok {
		fmt.Printf("%s\n%s\n", vhost.Path, vhost.AccessToken)
	}
	return
}
