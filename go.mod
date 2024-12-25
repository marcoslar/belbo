module github.com/marcoslar/belbo

go 1.23

require (
	github.com/BurntSushi/toml v0.3.1
	gopkg.in/russross/blackfriday.v2 v2.0.1
)

require (
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
)

replace gopkg.in/russross/blackfriday.v2 => github.com/russross/blackfriday/v2 v2.0.1
