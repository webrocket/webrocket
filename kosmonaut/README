Kosmonaut - Go client for WebRocket backend
===========================================

This package provides REQ Client and SUB Worker implementations compatible with
the latest version of the WebRocket backend protocol.

Usage
-----
Here's trivial example of the REQ client usage:

    package main

    import "github.com/webrocket/webrocket/pkg/kosmonaut"

    func main() {
        c := kosmonaut.NewClient("wr://{token...}@127.0.0.1:8081/hello")
        c.OpenChannel("world")
        c.Broadcast("world", "greetings", map[string]interface{}{
            "from": "Chris",
        })
        // ...
    }

And an example ot the SUB worker:

    package main

    import "github.com/webrocket/webrocket/pkg/kosmonaut"

    func main() {
        w := kosmonaut.NewWorker("wr://{token...}@127.0.0.1:8081/hello")
        for msg := range w.Run() {
            if msg.Error != nil {
                // do something with the error...
            } else {
                // do something with the message...
            }
        }
    }
    
For more information and examples check the package documentation.
	
Copyright
---------
Copyright (C) 2011 by Krzysztof Kowalik <chris@nu7hat.ch> and folks at Cubox

Released under the MIT license. See COPYING file for details.
