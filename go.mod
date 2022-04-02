module github.com/marcoslar/belbo

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	gopkg.in/russross/blackfriday.v2 v2.0.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
)

replace gopkg.in/russross/blackfriday.v2 => github.com/russross/blackfriday/v2 v2.0.1
