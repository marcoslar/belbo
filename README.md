# Belbo

A toy static site generator. Not ready for production, but it fulfills some basic needs:

- front-matter handling
- layout templating
- markdown to html processing
- custom Golang template functions

If there is something that makes it different from others is that it enables
what I call *parablogs* (you can read more about it [here][1]).

### Installation

```
go install github.com/marcoslar/belbo/cmd/belbo@latest
```

### Custom template functions

You can add custom functions to be used in your templates. See an example in `example/plugins`.

[1]: https://www.marcoslar.com/posts/2020/08/parallel-blogs/
