# Overview

[![Build Status](https://travis-ci.org/catcombo/go-staticfiles.svg)](https://travis-ci.org/catcombo/go-staticfiles)
[![Code Coverage](http://gocover.io/_badge/github.com/catcombo/go-staticfiles)](http://gocover.io/github.com/catcombo/go-staticfiles)
[![GoDoc](https://godoc.org/github.com/catcombo/go-staticfiles?status.svg)](https://godoc.org/github.com/catcombo/go-staticfiles)

`staticfiles` is an asset manager for a web applications written in Go. It collects asset files (CSS, JS, images, etc.)
from a different locations (including subdirectories), appends hash sum of each file to its name and copies files
to the target directory to be served by `http.FileServer`.

This approach allows to serve files without having to clear a CDN or browser cache every time the files was changed.
This also allows to use aggressive caching on CDN and HTTP headers to implement so called
[cache hierarchy strategy](https://developers.google.com/web/fundamentals/performance/optimizing-content-efficiency/http-caching#invalidating_and_updating_cached_responses).
If you ever worked with [Django](https://www.djangoproject.com/) you'll find it very similar
to the `staticfiles` application.


# Installation

`go get -u github.com/catcombo/go-staticfiles/...`


# Usage

There are two ways to collect files:

1. Using command line tool

    Run `collectstatic --output web/staticfiles --input assets/static --input media/`

    Init storage in your code:
    ```go
    storage, err := staticfiles.NewStorage("web/staticfiles")
    ```
   
    **Pros**: Run separately from the main application and doesn't influence it startup time.
    It can be run on a docker container build stage, for example.

    **Cons**: You may forget to run the command if you didn't schedule it's start.

2. Collect files every time the program starts

    ```go
    storage, err := staticfiles.NewStorage("web/staticfiles")
    storage.AddInputDir("assets/static")
    storage.AddInputDir("media")
    
    err := storage.CollectStatic()
    ```

    **Pros**: Collecting files runs automatically every time the program starts.

    **Cons**: Collecting files need a time. Thus, the application is running but is not
    accept incoming connections until copying and processing is finished.


To use in templates, define a static files prefix and register a template function
to resolve storage file path from its original relative file path:
```go
staticFilesPrefix := "/static/"
staticFilesRoot := "output/dir"

storage, err := NewStorage(staticFilesRoot)

funcs := template.FuncMap{
    "static": func(relPath string) string {
        return staticFilesPrefix + storage.Resolve(relPath)
    },
}
tmpl, err := template.New("").Funcs(funcs).ParseFiles("templates/page.html")
```

Now you can call `static` function in templates like this `{{static "css/style.css"}}`.
The generated output will be `/static/css/style.d41d8cd98f00b204e9800998ecf8427e.css` (hash may vary).


# Serve static files

To serve static files from the storage output directory pass `storage` as an argument to the `http.FileServer`.
```go
storage.OutputDirList = false    // Disable directories listing, optional
handler := http.StripPrefix(staticFilesPrefix, http.FileServer(storage))
http.Handle(staticFilesPrefix, handler)
```

It's often required to change assets during development. `staticfiles` uses cached versions of the original files
and to refresh files you need to run `collectstatic` every time you change a file. Enable development mode
by set `storage.Enabled = false` will force `storage` to read original files instead of cached versions.
Don't forget to enable storage back in production.


# Post-processing

`staticfiles` post-process `.css` files to fix files references.

Sample input file `css/style.css`
```css
@import "import.css";

div {
    background: url("../img/pix.png");
}
```

Output file `css/style.d41d8cd98f00b204e9800998ecf8427e.css` (hashes may vary)
```css
@import "import.5f15d96d5cdb4d0d5eb6901181826a04.css";

div {
    background: url("../img/pix.3eaf17869bb51bf27bd7c91bc9853973.png");
}
```


# Writing custom post-processing rules

You can add custom rule to post-process files. A rule is a simple function with a signature
`func(*Storage, *StaticFile) error` which must be registered with `storage.RegisterRule(CustomRule)` 
See `postprocess.go` as an example of `.css` post-processing implementation.
