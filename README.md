# Overview

[![Build Status](https://travis-ci.org/catcombo/go-staticfiles.svg)](https://travis-ci.org/catcombo/go-staticfiles)
[![GoDoc](https://godoc.org/github.com/catcombo/go-staticfiles?status.svg)](https://godoc.org/github.com/catcombo/go-staticfiles)

`staticfiles` collect files from a different directories (including subdirectories),
compute hash of each file, append hash sum to the filenames and copy files
to the output directory. This approach allows to use aggressive caching on CDN and
HTTP headers for static files and to implement so called
[cache hierarchy strategy](https://developers.google.com/web/fundamentals/performance/optimizing-content-efficiency/http-caching#invalidating_and_updating_cached_responses).
If you ever worked with [Django](https://www.djangoproject.com/) you'll find it
very similar to the `staticfiles` application.


# Installation

`go get -u github.com/catcombo/go-staticfiles/...`


# Usage

There are two ways to collect files:

1. Using command line tool

    Run `collectstatic --output web/staticfiles --input assets/static --input media/`.

    Init storage in your code:
    ```go
    storage := staticfiles.NewStorage("web/staticfiles")
    err := storage.LoadManifest()
    ```
   
    **Pros**: Run separately from the main application and doesn't influence it start up time.
    Also can be run on a docker container build stage to automate the process.

    **Cons**: You may forget to run the command if you do not configure it to run automatically.

2. Collect files every time the program starts

    ```go
    storage := staticfiles.NewStorage("web/staticfiles")
    storage.AddInputDir("assets/static")
    storage.AddInputDir("media")
    
    err := storage.CollectStatic()
    ```

    **Pros**: Collecting files runs automatically every time the program starts.

    **Cons**: Collecting files need a time. Thus, the application is running but is not
    accept incoming connections until collecting is finished.


For use in templates, define a static files prefix and register the template function
for resolving files in the storage at the original relative path:
```go
staticFilesPrefix := "/static/"
staticFilesRoot := "output/dir"

storage := NewStorage(staticFilesRoot)
err := storage.LoadManifest()

funcs := template.FuncMap{
    "static": func(relPath string) string {
        return staticFilesPrefix + storage.Resolve(relPath)
    },
}
tmpl := template.Must(
    template.New("").Funcs(funcs).ParseFiles("templates/main.tpl")
)
```

Now you can call `static` function in templates like this `{{static "css/style.css"}}`.
The call will be automatically converted to `/static/css/style.d41d8cd98f00b204e9800998ecf8427e.css` (hash may vary).


# Post-processing

`staticfiles` by default post-process `.css` files to fix files references.

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
`func(*Storage, *StaticFile) error` which must be registered by call `storage.RegisterRule(CustomRule)` 
See `postprocess.go` as an example of `.css` post-processing implementation.


# Disable static directory listing

Its often require to disable directory listing when serving static files via `http.FileServer`.
`staticfiles` comes with `staticfiles.FileSystem` which implements this feature. Usage:

```go
fs := staticfiles.FileSystem(staticFilesRoot, true)
h := http.StripPrefix(staticFilesPrefix, http.FileServer(fs))
```
