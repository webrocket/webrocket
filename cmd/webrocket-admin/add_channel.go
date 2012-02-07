package main

func channelParams(params []string) (vhost, name string, ok bool) {
	if len(params) == 2 && params[0] != "" && params[1] != "" {
		ok, vhost, name = true, params[0], params[1]
	}
	return
}

func addChannel(params []string) (err error, ok bool) {
	var vhost, name string
	if vhost, name, ok = channelParams(params); !ok {
		return
	}
	_, err = performRequest("POST", "/channels", "channel", map[string]string{
		"vhost": vhost,
		"name":  name,
	})
	if err != nil {
		return
	}
	ok = true
	return
}
