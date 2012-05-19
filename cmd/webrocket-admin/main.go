package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	webrocket "github.com/webrocket/webrocket/engine"
	"net/http"
	"os"
)

var (
	// The address of the admin endpoint.
	Addr string
	// Cookie string.
	Cookie string
	// The command name to be executed. 
	Cmd string
	// The name of the node.
	Node string
)

// Command represents information about for the concrete command
// function.
type Command struct {
	// The name of the command.
	Name string
	// The concrete command function, which accepts list of string
	// parameters and returns error if something went wrong, or false
	// when parameters are invalid, never both.
	Fn func([]string) (error, bool)
	// Params list displayed on help message.
	Params string
	// Description displayed on help message.
	Description string
}

// List of available commands.
var Commands = []*Command{
	&Command{"list_vhosts", listVhosts, "", "Shows list of the registered vhosts"},
	&Command{"add_vhost", addVhost, "[path]", "Registers new vhost"},
	&Command{"delete_vhost", deleteVhost, "[path]", "Removes specified vhost"},
	&Command{"show_vhost", showVhost, "[path]", "Shows information about the specified vhost"},
	&Command{"clear_vhosts", clearVhosts, "", "Removes all vhosts"},
	&Command{"regenerate_vhost_token", regenerateVhostToken, "[path]", "Generates new access token for the specified vhost"},
	&Command{"list_channels", listChannels, "[vhost]", "Shows list of channels opened under given vhost"},
	&Command{"add_channel", addChannel, "[vhost] [name]", "Opens new channel under given vhost"},
	&Command{"delete_channel", deleteChannel, "[vhost] [name]", "Removes channel from the specified vhost"},
	&Command{"clear_channels", clearChannels, "[vhost]", "Removes all channel from the specified vhost"},
	&Command{"list_workers", listWorkers, "[vhost]", "Shows list of the backend workers connected to the specified vhost"},
}

// findCommands searches for the command with specified name.
//
// name - The name to be found.
//
// Returns command if found and the status.
func findCommand(name string) (*Command, bool) {
	for _, cmd := range Commands {
		if cmd.Name == name {
			return cmd, true
		}
	}
	return nil, false
}

// Response contains information of the performed operation's response.
type Response struct {
	// HTTP response code.
	Code int
	// Unmarshalled response data.
	Data interface{}
}

// urlFor generates full request URL for specified path and map
// of parameters.
//
// path - Request path
//
// Examples
//
//     urlFor("/hello")
//     // => http://host:8083/hello
//
// Returns full URL.
func urlFor(path string) string {
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}
	return "http://" + Addr + path
}

// performRequest prepares and performs request of specified method
// for given path. Afterword it decodes the information from specified
// namespace.
//
// method    - Request method ("GET", "POST", etc.).
// path      - Request path (eg. "/hello").
// namespace - The JSON namespace to decode from.
//
// Returns decoded response.
func performRequest(method, path, namespace string) (
	*Response, error) {
	c := &http.Client{}
	req, _ := http.NewRequest(method, urlFor(path), nil)
	req.Header.Set("X-WebRocket-Cookie", Cookie)
	res, err := c.Do(req)
	if err != nil {
		goto requestError
	}
	if res, err = followRedirects(res); err != nil {
		goto requestError
	}
	return decodeResponse(res, namespace)
requestError:
	return nil, errors.New("couldn't perform the operation, is server running?")
}

// followRedirects is a custom HTTP redirects handler which appends
// cookie header to the request.
//
// r - The original response.
//
// Returns response from the new location.
func followRedirects(r *http.Response) (*http.Response, error) {
	if location, err := r.Location(); err == nil && location != nil {
		c := &http.Client{}
		req, _ := http.NewRequest("GET", location.String(), nil)
		req.Header.Set("X-WebRocket-Cookie", Cookie)
		return c.Do(req)
	}
	return r, nil
}

// decodeResponse takes a HTTP response objects and decodes information
// from it depending on the response code and requested namespace.
//
// r         - The response to be decoded.
// namespace - Requested namespace.
//
// Returns decoded response object or an error if something went wrong.
func decodeResponse(r *http.Response, namespace string) (*Response, error) {
	var data map[string]interface{}
	var res = &Response{Code: r.StatusCode}

	if namespace == "" && res.Code < 400 {
		// Everything's fine, no data attached to this response.
		return res, nil
	}
	// Decode data from the response...
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&data); err != nil {
		goto decodeError
	}
	if idata, ok := data[namespace]; ok && res.Code < 400 {
		// Everything's fine, data attached is valid.
		res.Data = idata
		return res, nil
	}
	if imsg, ok := data["error"]; ok && res.Code >= 400 {
		if msg, ok := imsg.(string); ok {
			// Valid error response received...
			err := errors.New(msg)
			return nil, err
		}
	}
decodeError:
	return nil, errors.New("couldn't perform the operation, invalid response!")
}

// Vhost represents the information about single vhost.
type Vhost struct {
	// The path to the vhost.
	Path string
	// Single access token.
	AccessToken string
}

// maybeVhosts takes an interface value and converts it to vhost information
// record if possible.
//
// x - The object to be converted to vhost.
//
// Returns decoded vhost and status, true if everything was fine.
func maybeVhost(x interface{}) (v *Vhost, ok bool) {
	var data map[string]interface{}
	v = &Vhost{}
	if data, ok = x.(map[string]interface{}); !ok {
		return nil, false
	}
	if v.Path, ok = data["path"].(string); !ok {
		return nil, false
	}
	if v.AccessToken, ok = data["accessToken"].(string); !ok {
		return nil, false
	}
	ok = true
	return
}

// Channel represents information of the single channel.
type Channel struct {
	// The channel's name.
	Name string
	// Number of the active subscribers.
	SubscribersSize int
}

// maybeChannel takes an interface and converts it to the channel information
// record if possible.
//
// x - The object to be converted to channel.
//
// Returns decoded channel and status, true if everything was fine.
func maybeChannel(x interface{}) (ch *Channel, ok bool) {
	var data map[string]interface{}
	var subscribers map[string]interface{}
	var ssize float64
	ch = &Channel{}
	if data, ok = x.(map[string]interface{}); !ok {
		return nil, false
	}
	if ch.Name, ok = data["name"].(string); !ok {
		return nil, false
	}
	if subscribers, ok = data["subscribers"].(map[string]interface{}); !ok {
		if ssize, ok = subscribers["size"].(float64); !ok {
			ch.SubscribersSize = int(ssize)
		}
	}
	ok = true
	return
}

// Worker represents information of the single worker connection.
type Worker struct {
	// The worker's unique identifier.
	Id string
}

// maybeWorker takes an interface and converts it to the worker information
// record if possible.
//
// x - The object to be converted to worker.
//
// Returns decoded worker and status, true if everything was fine.
func maybeWorker(x interface{}) (w *Worker, ok bool) {
	var data map[string]interface{}
	w = &Worker{}
	if data, ok = x.(map[string]interface{}); !ok {
		return nil, false
	}
	if w.Id, ok = data["id"].(string); !ok {
		return nil, false
	}
	ok = true
	return
}

func init() {
	flag.StringVar(&Addr, "admin-addr", "127.0.0.1:8082", "Address of the server's admin interface")
	flag.StringVar(&Cookie, "cookie", "", "Cookie string generated by the server")
	flag.StringVar(&Node, "node", "", "Name of the node")
	flag.Parse()

	Cmd = flag.Arg(0)

	if Node == "" {
		Node = webrocket.DefaultNodeName()
	}
	if Cookie == "" {
		Cookie = webrocket.ReadCookie(Node)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] command [args ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nAvailable commands\n")
	for _, cmd := range Commands {
		var params string
		if cmd.Params != "" {
			params = " " + cmd.Params
		}
		fmt.Fprintf(os.Stderr, "  %s%s: %s\n", cmd.Name, params, cmd.Description)
	}
	fmt.Fprintf(os.Stderr, "\nAvailable options\n")
	flag.PrintDefaults()
}

func main() {
	var ok bool
	var err error
	var cmd *Command

	cmd, ok = findCommand(Cmd)
	if !ok {
		goto usage
	}
	err, ok = cmd.Fn(flag.Args()[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERR: %v\033[0m\n", err)
		os.Exit(1)
	}
	if !ok {
		goto usage
	}
	return
usage:
	usage()
	os.Exit(1)
}
