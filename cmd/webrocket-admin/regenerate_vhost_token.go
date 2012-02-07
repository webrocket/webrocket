package main

import "fmt"

func regenerateVhostToken(params []string) (err error, ok bool) {
	var path string
	var res *Response
	if path, ok = vhostParams(params); !ok {
		return
	}
	res, err = performRequest("PUT", "/vhost/token", "vhost", map[string]string{
		"path": path,
	})
	if err != nil {
		return
	}
	if vhost, ok := maybeVhost(res.Data); ok {
		fmt.Printf("%s\n", vhost.AccessToken)
	}
	return
}
