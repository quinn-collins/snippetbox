# 'Let's Go' by Alex Edwards
# Notes

## Project tree
> Inspired by: https://peter.bourgon.org/go-best-practices-2016/#repository-structure

> TODO: READ his recommended links: 
> https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1
> https://github.com/thockin/go-build-template
```
.
├── README.md
├── \
├── cmd
│   └── web
│       ├── context.go
│       ├── handlers.go
│       ├── handlers_test.go
│       ├── helpers.go
│       ├── main.go
│       ├── middleware.go
│       ├── middleware_test.go
│       ├── routes.go
│       ├── templates.go
│       ├── templates_test.go
│       └── testutils_test.go
├── go.mod
├── go.sum
├── internal
│   ├── assert
│   │   └── assert.go
│   ├── models
│   │   ├── errors.go
│   │   ├── mocks
│   │   │   ├── snippets.go
│   │   │   └── users.go
│   │   ├── snippets.go
│   │   ├── testdata
│   │   │   ├── setup.sql
│   │   │   └── teardown.sql
│   │   ├── testutils_test.go
│   │   ├── users.go
│   │   └── users_test.go
│   └── validator
│       └── validator.go
├── tls
│   ├── cert.pem
│   └── key.pem
└── ui
    ├── efs.go
    ├── html
    │   ├── base.tmpl.html
    │   ├── pages
    │   │   ├── create.tmpl.html
    │   │   ├── home.tmpl.html
    │   │   ├── login.tmpl.html
    │   │   ├── signup.tmpl.html
    │   │   └── view.tmpl.html
    │   └── partials
    │       └── nav.tmpl.html
    └── static
        ├── css
        │   ├── index.html
        │   └── main.css
        ├── img
        │   ├── favicon.ico
        │   ├── index.html
        │   └── logo.png
        ├── index.html
        └── js
            ├── index.html
            └── main.js
```

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
  - It's rare to use w.WriteHeader() and w.Write() methods directly
- Go will attempt to resolve named ports by checking /etc/services when starting the server
- net/http constants can be used for common HTTP status codes
- When sending a response to the user Go will automatically set `Date` `Content-Length` and `Content-Type`
  - Go attempts to set `Content-Type` by sniffing response bodies with http.DetectContentType()
  - `Content-Type: application/octet-stream` is the fallback when Go cannot detect the type
- Can use `r.URL.Query().Get()` to retrieve request URL query strings
- Can use `strconv.Atoi()` to parse strings to integers
- `fmt.Fprintf()` takes an `io.Writer` interface which `http.ResponseWriter` satisfies


## Commands Covered
`go run .`
`go run main.go`
`go run snippetbox.qcollins.net`
