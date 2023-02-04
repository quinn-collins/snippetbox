# 'Let's Go' by Alex Edwards
# Notes

## Overview of architecture and design decisions

### Decisions
1. Need: A way to execute application logic and write HTTP response headers and bodies
- Used: A handler function in go that accepts an http.ResponseWriter and a *http.Request
2. Need: A way to store a mapping between the URL patterns and their corresponding handlers
- Used: A new ServeMux() and registered handlers via mux.HandleFunc(path, handlerFunction)
3. Need: A way to listen for incoming requests
- Used: http.ListenAndServe(port, servemux)
4. Need: A way to make `/` behave like a fixed path and return NOT FOUND if path does not match
- Used: conditional to check if path does not equal `/` that returns http.NotFound(w, r)

### Notes
- `go run` is a shortcut command that compiles code and creates an executable in `/tmp`
- servemux
  - Go's servemux treats the URL pattern "/" like a catch-all.
  - Supports fixes paths `/snippet/view` and subtree paths `/` `/static/`
  - Fixed paths are only matched when path matches exactly
  - Subtree paths are matched when the start of a request path matches
  - Longer URL patterns take precedence over shorter ones
  - URL paths are automatically sanitized I.e. directory manipulation with `..` or `////`
  - Automatic redirect to matching subtree path
  - Does not support routing based on request method
  - Does not support clean URLs with variables
  - Does not support regexp patterns
- http.ResponseWriter
  - Can only call w.WriteHeader() once per response
  - Can not call w.WriteHeader() after status code has been written
  - w.Write() will send a `200 OK` if w.WriteHeader() is not called explicitly
  - Can let the user know what request methods are allowed with `w.Header().Set("Allow", "POST")
  - Can use http.Error(w, string, statusCode) to send a non-200 and plain-text response body
    - Note we are passing http.ResponseWriter to a function that sends a response on our behalf
- Go will attempt to resolve named ports by checking /etc/services when starting the server


## Commands Covered
`go run .`
`go run main.go`
`go run snippetbox.qcollins.net`
