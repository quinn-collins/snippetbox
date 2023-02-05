# 'Let's Go' book by Alex Edwards
https://lets-go.alexedwards.net/ \
Notes by Quinn Collins

## Routes

| Method  | Pattern | Handler | Action | Middleware Chain |
| ------------- | ------------- | ------------- | ------------- | ------------- |
| GET | / | home | Display the home page | Dynamic |
| GET | snippet/view/:id | snippetView | Display a specific snippet | Dynamic |
| GET | /snippet/create | snippetCreate | Display a HTML form for creating a new snippet | Protected |
| POST | /snippet/create | snippetCreatePost | Create a new snippet | Protected |
| GET | /user/signup | userSignup | Display a HTML form for signing up a new user | Dynamic |
| POST | /user/signup | userSignupPost | Create a new user | Dynamic |
| GET | /user/login | userLogin | Display a HTML form for logging in a user | Dynamic |
| POST | /user/login | userLoginPost | Authenticate and login a user | Dynamic |
| POST | /user/logout | userLogoutPost | Logout the user | Protected |
| GET | /static/\*filepath | http.FileServer | Serve a specific static file | Dynamic |

## Project tree
```
.
├── README.md
├── \
├── cmd  # Application-specific code for executable applications within project
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
> Inspired by: https://peter.bourgon.org/go-best-practices-2016/#repository-structure

> TODO: Read Peter Bourgon recommended links for new best practices
> - https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1
> - https://github.com/thockin/go-build-template

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
### Managing Configuration Settings
#### Environment Variables and Command-line Flags
- `go run ./cmd/web -addr=":80"`
```
addr := flag.String("addr", ":4000", "HTTP network address")
flag.Parse()
err := http.ListenAndServe(*addr, mux)
```
- You can use environment variables while starting the application
- `go run ./cmd/web -addr=$SNIPPET_BOX_HTTP_PORT`
### Leveled Logging
- Prefix information messages with **INFO** and error messages with **ERROR**
- `infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)`
- `errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)`
- Can redirect standard out and standard error streams to different places from the start of the application
- `go run ./cmd/web >>/tmp/info.log 2>>/tmp/error.log`
- Create a new http.Server struct with our new error logger
```
srv := &http.Server{
  Addr: *addr,
  ErrorLog: errorLog,
  Handler: mux,
}
```
### Dependency Injection
- Put dependencies in a custom application struct
- Define handler functions as methods against application struct
- Initialize instance of application struct
- Pass application struct methods(handlers) into mux
- Current Dependencies:
```
type application struct {
  errorLog *log.Logger
  infoLog *log.Logger
  snippets models.SnippetModelInterface
  users models.UserModelInterface
  templateCache map[string]*template.Template
  formDecoder *form.Decoder
  sessionManager *scs.SessionManager
}
```
### Centralized Error Handling
- Move error handling into helper methods on the application struct
```
func (app *application) serverError(w http.ResponseWriter, err error) {}
func (app *application) clientError(w http.ResponseWriter, status int) {}
func (app *application) notFound(w http.ResponseWriter) {}
```
### Database-driven Response
#### Setting up the database and connection
- Installed MySql locally
- Scaffolded the database. Created database added snippets table with some data.
- Created a user to restrict the amount of access our application has while running.
- Installed a [driver](https://github.com/go-sql-driver/mysql)
#### Creating the database connection pool
- Go's `sql.Open()` function used to return a **sql.DB** object
`db, err := sql.Open("mysql", "web:pass@/snippetbox?parseTime=true")`
- A **sql.DB** object is a pool of many connections
- Go manages connections in the connection pool via the driver.
- We use defer on a `db.Close()` call to close the connection pool before `main()` function exits
#### Designing the database model (I.e. service layer or data access layer)
- Add in a struct for data and a struct for the model under **internal/models/**
- Add methods on the model for CRUD operations etc.
- Add **prepared** SQL statements to methods
- Pass models to handlers via dependency injection
- This makes for a clean separation of concerns where our database logic isn't tied to our handlers
- Models actions are mockable and testable
### Dynamic HTML Templates
- Render a Go template from the handler passing in data from the model
- Access data in the template via `.` syntax.
- Wrap data in a struct within handler so that we can pass multiple pieces of dynamic data.
- Caching templates so that we aren't parsing the files from the hard drive repeatedly
```
cache := map[string]*template.Template{}
pages, err := filepath.Glob("./ui/html/pages/*.tmpl")

for _, page := range pages {
    name := filepath.Base(page)
    files := []string{
        "./ui/html/base.tmpl",
        "./ui/html/partials/nav.tmpl",
        page,
}

ts, err := template.ParseFiles(files...)
cache[name] = ts
}
```
- Add template cache to application struct for dependency injection
- Initialize a new template cache
- Add cache to application dependencies
- Make template render a two-stage process to avoid runtime errors within our template that return a **200 OK** to our user
```
buf := new(bytes.Buffer)
err := ts.ExecuteTemplate(buf, "base", data)
w.WriteHeader(status)
buf.WriteTo(w)
```
### Middleware
- Create middleware functions that accept **http.Handler** and return **http.Handler** by calling **next** handler forming a closure
- Middleware chain, any code before `next.ServeHTTP(w, r)` is called on the way down the chain, and after is called on the way up
- Panic recovery to send a neat error back on a panic within a request lifetime in a goroutine
- Log Requests
- We used lightweight [justinas/alice](https://github.com/justinas/alice) to compose our middleware chain
```
dynamic: sessionManager.LoadAndSave ↔ noSurf ↔ app.authenticate
protected: sessionManager.LoadAndSave ↔ noSurf ↔ app.authenticate ↔ app.requireAuthentication
standard: recoverPanic ↔ logRequest ↔ secureHeaders ↔ servemux ↔ application handler
```
- `return` before a `next.ServeHTTP(w, r)` will stop the chain from being executed.
### Advanced Routing
- Introduce a third-party router [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
- Library chosen for being focused, lightweight, and fast. Automatically handles **OPTIONS** requests and sends appropriate responses.
- Library does not support regexp route patterns and 'grouping' of routes which use specific middleware.
- We can have Clean URLS and method-based routing via this library.
- Override httprouter.NotFound with our own handler function that wraps our app.notFound(w) helper
### Processing Forms
- Set up our form using **action** and **method** attributes so our form will POST data to **/snippet/create**
- We parse the form from the handler by calling `r.ParseForm()` and retrieving data by `r.PostForm.Get("title")`
### Form Validation
- Set up a validator package that contains helper functions for validating forms and a struct for holding errors.
- Utilize the validators when receiving POST requests to validate the data coming in.
- Manage the validation errors gracefully by re-displaying the HTML form, highlighting the fields which failed and re-populating previously submitted data.
- In our handler we check for validation error, if it exists we populate a map FieldErrors[string]string. If map is not empty we re-display the template with the data we received on the last POST request utilizing Go template `{{with .Data}}` syntax.
### Stateful HTTP & Session Management
- We use [alexedwards/scs](https://github.com/alexedwards/scs) to make session management easier.
- We store the session data server-side in MySQL.
- Session data is a combination of a unique token, binary data in a BLOB type, and an expiry field.
- We add session management to a new middleware chain that is only called where POST requests are received.
- On successful POST we add a flash message to the current request context.
- We include the flash message in the templateData struct wrapper we made to automate the display of flash messages.
### Security
- We used [crypto/tls/generate_cert.go](https://go.dev/src/crypto/tls/generate_cert.go) to generate a self-signed certificate for TLS for development purposes.
- Set the `sessionManager.Cookie.Secure` value equal to `true` so that cookies are only sent when HTTPS is being used.
- Set up Go's http library to start up a HTTPS server for us.
- Changed the TLS config to use only eliptic curves that have assembly implementations to increase speed.
- Addded timeouts for **IDLE:** 1 minute, **READ:** 5 seconds, and **WRITE** 10 seconds 
- Added a READ timeout of 5 seconds to help mitigate the risk from slow-client attacks. Set an IDLE timeout of 1 minute so it does not default to 5 seconds.
- Added a WRITE timeout of 10 seconds to prevent data the handler returns from taking too long to write.
- For CSRF attacks we set SameSite attribute to lax on the session cookie so that the session cookie won't be sent by the user's browser for unsafe cross-site requests.
- For CSRF attacks we add a library [justinas/nosurf](https://github.com/justinas/nosurf) to manage a customized CSRF cookie that is added on all routes.
- We get the CSRF token from the request context, add it to template data so that it is available on every template. We add that CSRF token in a field like so:
```
<input type='hidden' name='csrf_token' value='{{.CSRFToken}}'>
```
### User Authentication
- Add a users model for accessing users database table.
- Use bcrypt to store a one-way hash of the password with a cost of 12.
- Include check for unique email in the POST so that we don't create a race condition by adding a method on UserModel.
- Once again add validation to the form for creating a user and signing in as a user.
- Authentication core takes place in a method on UserModel called Authenticate that gets details from the database and compares them with what was supplied with the request.
- Generate a new session token after authentication succeeds or logout happens.
- Add/Remove authenticatedUserID on Login/Logout
### User Authorization
- Authenticated users are authorized to see 'Home', 'Create snippet', and 'Logout'
- Unauthenticated users are authorized to see 'Home', 'Signup', and 'Login'
- We add a helper function to check to see if a user is authenticated by checking request context for `authenticatedUserId`
- We add a IsAuthenticated boolean to template data to determine what is shown on the page.
- Add new middleware chain requireAuthentication for routes that need it. Such as POSTing data to the server.
### Request Context
- Create a constant with type contextKey: string for storing our isAuthenticatedContextKey
- Add our key to the current constant in a middleware chain called authenticated. 
- Add a exists method on the usermodel to see if a user with a specific ID exists
- We retrieve the users id from their session data, check the database with our exists method, and update request context to include our context key
### Embedding
-
### Unit Testing
-
### End-to-end Testing
-
### Integration Testing
-
### Test Coverage Profiling
-

## Notes
- `go run` is a shortcut command that compiles code and creates an executable in `/tmp`
- servemux
  - Go's servemux treats the URL pattern "/" like a catch-all.
  - Supports fixed paths `/snippet/view` and subtree paths `/`, `/static/`
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
- Custom loggers created by `log.New()` are concurrency-safe
  - If multiple loggers are writing to the same destination you need to make sure underlying `Write()` method is safe for concurrent use
- Use closures for dependency injection when handlers are spread across mulitiple packages
- Can use `debug.Stack()` to get a stack trace for current goroutine
- Can use `http.StatusTexT()` to generate a human-friendly text representation of a given HTTP status code
- Error logger's `Output()` function may need frame depth set to return correct stack trace of where the error originated
- go.mod file contains exactr versions of packages used to help with reproducible builds
- go.sum file contains cryptographic checksums representing content of required packages
- dsn for database connection can include `parseTime=true` to convert SQL **TIME** and **DATE** to Go **time.Time**
- A **sql.DB** connection pool is safe for concurrent access and can be used form handlers safely
- Connection pool to database is intended to be long-lived. Don't call `sql.Open()` in a short-lived handler.
- Import paths can be prefixed with a `_` to denote that we won't be using anything in the package.
- Database connections are established lazily, as and when needed for the first time.
- Can use db.Ping() method to create a connection and check for errors.
- `errors.Is()` is best practice way to check for error equality
  - Go 1.13 added ability to wrap errors which made regular equality operators unuseable
- Errors from `DB.QueryRow()` are deferred until `Scan()` is called
- It's critical to close a **resultset* with `defer rows.Close()` to let the database connection close
- [jmoiron/sqlx](https://github.com/jmoiron/sqlx) Can be used to reduce verbosity of using the standard **database/sql** package
- Go does not handle NULL values in database records well
  - If we query a row that contains a **NULL** value and `rows.Scan()`, go won't be able to convert **NULL** into a string.
  - This can be fixed with `sql.NullString` or simply avoiding **NULL** values altogether.
- `Exec()`, `Query()`, `QueryRow()` can use any connection from `sql.DB` pool. They may not run on the same connection.
  - You can wrap multiple statements in a transaction to guarantee the same connection is used.
  - You must always call `Rollback()` or `Commit()` before a function returns otherwise the connection will stay open.
- Can use `DB.Prepare()` to create a prepared statement for reuse to eliminate the cost of re-preparing statements on database connections.
  - Prepared statements exist on database connections.
  - Tradeoff of complexity vs. performance
- **html/template** package automatically escapes any data between `{{ }}` which is helpful in preventing XSS attacks.
- When you invoke a template from within a template, data needs to be pipelined
```
{{template "main" .}}
{{block "sidebar" .}}{{end}}
```
- Methods can be called from a type passed into the template
- You can pass parameters to these methods like this:
```
<span>{{.Snippet.Created.AddDate 0 6 0}}</span>
```
- **html/template** always strips out any HTML comments including conditional comments
- Can add common dynamic data to a struct and then initialize it within a method to be used across templates within the handlers
- Custom template functions can be created with the `template.FuncMap` object and registered with the `template.Funcs() method
  - These steps must happen before you parse the templates
  - Custom template functions can return only one value and optionally error as a second value
- `{{.Created | humanDate}}` and `{{humanDate .Created}}` are equivalent
- Middleware design patterns:
```
func myMiddleware(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        // TODO: Execute our middleware logic here...
        next.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}
```
Or as a different pattern:
```
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // TODO: Execute our middleware logic here...
        next.ServeHTTP(w, r)
    })
}
```
- If we spin up another goroutine within our handlers we'll have to account for panics not being recovered by our middleware chain
- Struct fields must be exported in order to be read by html/template package when rendering a template
- In a template you can access a Go map[string]string] just by chaining the key name on. `{{.Form.FieldErrors.title}}`


## Commands Covered
`go run .`\
`go test .`\
`go build .`\
`go run main.go`\
`go run snippetbox.qcollins.net`\
`go run ./cmd/web`\
`go run ./cmd/web -addr=":80"`\
`go run ./cmd/web -help`\
`go run ./cmd/web >>/tmp/info.log 2>>/tmp/error.log`\
`go get github.com/go-sql-driver/mysql@v1`\
`go get github.com/go-sql-driver/mysql`\
`go get github.com/go-sql-driver/mysql@v1.0.3`\
`go mod verify`\
`go mod download`\
`go get -u github.com/foo/bar`\
`go get -u github.com/foo/bar@v2.0.0`\
`go get github.com/foo/bar@none`\
`go mod tidy -v`
