package main

import "time"

// Example of custom functions that can be used in templates

// BelboFuncs must be a FuncMap
var BelboFuncs = map[string]interface{}{
	"slice": func(s string, from, to int) string {
		return s[from:to]
	},

	"reverse": func(s string) string {
		result := ""
		for _, ch := range s {
			result = string(ch) + result
		}

		return result
	},

	"SimpleDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}
