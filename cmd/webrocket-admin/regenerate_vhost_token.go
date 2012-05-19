package main

import "fmt"

func regenerateVhostToken(params []string) (err error, ok bool) {
	var path string
	var res *Response
	if path, ok = vhostParams(params); !ok {
		return
	}
	res, err = performRequest("PUT", path+"/token", "vhost")
	if err != nil {
		return
	}
	if vhost, ok := maybeVhost(res.Data); ok {
		fmt.Printf("%s\n", vhost.AccessToken)
	}
	return
}
