package main

import (
	"errors"
	"fmt"
	"sort"
)

func listWorkers(params []string) (err error, ok bool) {
	var vhost string
	var entries []interface{}
	var res *Response
	if vhost, ok = vhostParams(params); !ok {
		return
	}
	res, err = performRequest("GET", "/workers", "workers", map[string]string{
		"vhost": vhost,
	})
	if err != nil {
		return
	}
	if entries, ok = res.Data.([]interface{}); !ok {
		err = errors.New("couldn't list workers, invalid response")
		return
	}
	ids := make([]string, len(entries))
	for i, x := range entries {
		if worker, ok := maybeWorker(x); ok {
			ids[i] = worker.Id
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Printf("%s\n", id)
	}
	return
}
