# Belbo

A toy static site generator. Not ready for production, but it fulfills some basic needs:

- front-matter handling
- layout templating
- markdown to html processing 
- custom Golang template functions

If there is something that makes it different from others is that it enables 
what I call *parablogs* (you can read more about it [here][1]).

### Usage

Clone this repo and then run `go install cmd/belbo/main.go`. The binary
should be installed in your `$GOPATH/bin` directory.

### Custom template functions

You can add custom functions to be used in your templates. See an example in `example/plugins`.

[1]: https://lessmarcos.com/posts/2020/08/parallel-blogs/
