# Filesurf

A simple binary webserver written in Go.
I just wanted to write a bare-bones application that serves a local directory for download, and allows upload of files to the server.
Kinda like a binary, stripped-down version of the `python http.server` or `updog`.

## Synopsis
```
$ ./filesurf -h

Usage of ./filesurf:
  -cert string
        Path to TLS certificate file (required if -s)
  -d string
        Base directory to serve. (default ".")
  -key string
        Path to TLS key file (required if -s)
  -p string
        Port to listen on. (default "8090")
  -s    Enable HTTPS.
```

Example usage for uploading a file via HTTPS, with curl:

```
curl -X POST -k https://192.168.1.42/upload -F "content=@./myfile.txt"
```

Use the `-k` flag for "insecure" mode to accept the self-generated cert.

## What works
- Directory listing
- File download
- File upload via POST form object
- HTTPS
- Self-generating HTTPS certificates

## What is still missing
- parameterized file size limit
- some pretty-printing of directory listing for downloads
    - however, this would prolly require some more in-depth work
- perhaps a built-in upload frontend?
    - I don't want to have a seperate html/js file
    - basic idea was to serve all as a single binary
    - might include it as some hard-coded web page content, idk
