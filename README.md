# `Let's Go` by Alex Edwards - NOTES

## Overview of architecture and design decisions

1. Need a way to execute application logic and write HTTP response headers and bodies
- We use a handler function in go that accepts an http.ResponseWriter and a *http.Request
1. Need a way to store a mapping between the URL patterns and their corresponding handlers
- We use a new ServeMux() and register handlers
1. Need a way to listen for incoming requests
- We use http.ListenAndServe() function
