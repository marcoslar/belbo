package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/lessmarcos/belbo/belbo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	RootPath       = "."
	ConfigFilename = ".belbo.toml"
)

// DefaultCfg provides the default values for a .belbo.toml config file
var DefaultCfg = belbo.Params{
	// Where the content exists
	"content_dir": "posts",

	// Where layouts and partials exist
	"templates_dir": "templates",

	// Where HTML is written out
	"output_dir": "public",

	// Where JS, CSS, images, etc. exist
	"static_dir": "static",

	// Start up a local web server?
	"local_server": true,

	"frontmatter_sep": "---",

	"root_path": ".",
}

func main() {
	log.SetFlags(0)

	// Read .belbo.toml config file
	cfgAsToml, err := ioutil.ReadFile(filepath.Join(RootPath, ConfigFilename))
	if err != nil {
		log.Println("- could not find a valid .belbo.toml config file. Using default values")
		cfgAsToml = []byte("")
	}

	if _, err := toml.Decode(string(cfgAsToml), &DefaultCfg); err != nil {
		panic(err)
	}

	belbo.SetCfg(DefaultCfg)
	pagesToProcess := belbo.PagesToProcess()

	if len(pagesToProcess) == 0 {
		fmt.Println("- belbo found nothing to process")
		os.Exit(0)
	}

	if err := os.RemoveAll(belbo.OutputDir); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(belbo.OutputDir, os.ModePerm); err != nil {
		panic(err)
	}

	for _, page := range pagesToProcess {
		log.Println("+ Processing " + page.RelativePath)
		page.AllPages = pagesToProcess
		page.ToHtml()
	}

	// Copy static dir to output directory
	staticDir := DefaultCfg.GetString("static_dir")
	if belbo.Exists(staticDir) {
		outputDir := DefaultCfg.GetString("output_dir")
		if err := belbo.CopyDirectory(
			filepath.Join(RootPath, staticDir),
			filepath.Join(RootPath, outputDir, staticDir)); err != nil {
			panic(err)
		}
	}

	// Start local web server
	if DefaultCfg.Get("local_server").(bool) {
		http.Handle("/", http.FileServer(http.Dir(DefaultCfg.GetString("output_dir"))))
		log.Println("Serving on http://localhost:4433")

		if err := http.ListenAndServe(":4433", nil); err != nil {
			log.Fatalln(err)
		}
	}
}
