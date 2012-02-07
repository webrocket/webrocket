package main

func clearChannels(params []string) (err error, ok bool) {
	var vhost string
	if vhost, ok = vhostParams(params); !ok {
		return
	}
	_, err = performRequest("DELETE", "/channels", "", map[string]string{
		"vhost": vhost,
	})
	return
}
