package main

func deleteChannel(params []string) (err error, ok bool) {
	var vhost, name string
	if vhost, name, ok = channelParams(params); !ok {
		return
	}
	_, err = performRequest("DELETE", "/channel", "", map[string]string{
		"vhost": vhost,
		"name":  name,
	})
	return
}
