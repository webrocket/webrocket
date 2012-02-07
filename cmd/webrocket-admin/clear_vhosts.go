package main

func clearVhosts(params []string) (err error, ok bool) {
	_, err = performRequest("DELETE", "/vhosts", "", map[string]string{})
	if err != nil {
		return
	}
	return nil, true
}
