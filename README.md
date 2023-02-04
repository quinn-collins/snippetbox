# 'Let's Go' book by Alex Edwards - Notes by Quinn Collins

## Project tree
> Inspired by: https://peter.bourgon.org/go-best-practices-2016/#repository-structure

> TODO: READ his recommended links for new best practices \
> - https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1 \
> - https://github.com/thockin/go-build-template \
```
.
├── README.md
├── \
├── cmd  # Appliation-specific code for executable applications within project
│   └── web # Executable application
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
├── internal # Ancillary non-application-specific code, potentially re-usable code across applications
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
└── ui # User-interface assets used by the web application
    ├── efs.go
    ├── html # HTML templates
    │   ├── base.tmpl.html # Master template for shared content
    │   ├── pages
    │   │   ├── create.tmpl.html
    │   │   ├── home.tmpl.html
    │   │   ├── login.tmpl.html
    │   │   ├── signup.tmpl.html
    │   │   └── view.tmpl.html
    │   └── partials # HTML templates to be reused in different pages or layouts
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

## Architecture Decisions
### Routing Requests
- Go functions that accept `http.ResponseWriter` & `*http.Request` passed to `http.HandlerFunc()`
- Chain handlers together via `ServeHTTP()` interface
- Handlers managed by a Go `servemux` (HTTP request multiplexer) AKA a router
- ServeMux is created and we create a mapping between url and handler via `mux.HandleFunc(path, handlerFunction)`
- Listen for incoming requests via `http.ListenAndServer(port, mux)`
### Serving Content
- Parse Go templates with `ts, err := template.ParseFiles(files)`
- Write template content to respones body with `ts.ExecuteTemplate(w, template, nil)`
- Static content served with `http.FileServer`
- Pass the fileserver into mux to create a route at `/static/`
- Use static content in templates by adding links in the `head` of the HTML document+
### Configuration
-
### Error Handling
-

## Notes
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
- `internal` directory
  - Any packages under `internal` can only be imported by code inside the parent of `internal`
- `html/template`
  - `ParseFiles()` must either be relative to current working directory or an absolute path
- Go HTML Templates
  - `{{define "base"}}...{{end}}` defines a distinct named template called base
  - `{{template "title" .}}` actions denote that we want to invoke other named templates i.e. `title`
  - `.` represents dynamic data to be passed to the invoked template
  - `{{block}}...{{end}}` can be used instead of `{{template}}` if you want to include default content i.e. a sidebar
- `net/http` `fileserver`
  - All request paths are sanitized by running them through `path.Clean()`
  - Supports [Range Requests](https://benramsey.com/blog/2008/05/206-partial-content-and-range-requests/)
  - `Last-Modified` and `If-Modified-Since` headers are transparently supported
  - `Content-Type` is automatically set from the file extension using `mime.TypeByExtension()` function
  - You can add custom extensions and content types using `mime.AddExtensionType()`
  - `http.FileServer` will most likely serve files from RAM after inital application run
  - `http.ServeFile()` can be used to serve a single file form within a handler but does not automatically sanitize the file path
  - [Disable FileServer Directory Listings](https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings)
- `http.Handler` interface
  - Handler is any object that satisfies the `http.Handler` interface
    - i.e - has a `ServeHTTP(ResponseWriter, *Request)` function
  - Functions can be passed to `http.HandlerFunc()` to make them satisfy the interface
  - `servemux` also satisfies the `http.Handler` interface so that we may chain handlers
  - Common Go idiom is to chain `ServeHTTP()` handlers which is how we can think of this app
- All incoming HTTP requests are served in their own goroutine
  - Code called in or by your handlers will most likely be running concurrently
  - Be aware of race conditions when accessing shared resources from handlers


## Commands Covered
`go run .`
`go run main.go`
`go run snippetbox.qcollins.net`
`go run ./cmd/web`
