package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/marcoslar/belbo/internal"
)

const (
	RootPath       = "."
	ConfigFilename = ".belbo.toml"
	Version        = "v0.2.1"
)

// DefaultCfg provides default values for a .belbo.toml config file
var DefaultCfg = internal.Config{
	// Where the content exists
	"content_dir": []string{"posts", "logs"},

	// Where layouts and partials exist
	"templates_dir": "templates",

	// Template filenames
	"templates": []string{},

	// Where Html is written out
	"build_dir": "dist",

	// Where JS, CSS, images, etc. exist
	"static_dir": "static",

	// Start up a local web server?
	"local_server": true,
	"server_port":  "4433",

	"frontmatter_sep": "---",

	"root_path": RootPath,
}

func main() {
	tomlConfig, err := os.ReadFile(filepath.Join(RootPath, ConfigFilename))

	if err != nil {
		log.Println("- could not find a .belbo.toml config file")
		tomlConfig = []byte("")
	}

	DefaultCfg, err := internal.NewConfig(string(tomlConfig), &DefaultCfg)
	if err != nil {
		log.Fatalln("- could not create a valid config file.", err)
	}

	belboFuncs := internal.LoadFuncsAsPlugins(DefaultCfg.GetString("plugins_dir"))

	siteGenerator := internal.NewBelbo(DefaultCfg, os.DirFS(RootPath), belboFuncs)
	siteGenerator.BuildPages()

	if len(siteGenerator.Pages) == 0 {
		log.Println("- belbo found nothing to process")
		os.Exit(0)
	}

	if err := os.RemoveAll(siteGenerator.BuildDir); err != nil {
		log.Fatalf("- could not remove %s directory. %s", siteGenerator.BuildDir, err)
	}

	if err := os.MkdirAll(siteGenerator.BuildDir, os.ModePerm); err != nil {
		log.Fatalf("- could not create %s directory. %s", siteGenerator.BuildDir, err)
	}

	for _, p := range siteGenerator.Pages {
		p.AllPages = siteGenerator.Pages
		log.Println("+ processing " + p.RelativePath)

		if siteGenerator.IsPlainFile(p.RelativePath) {
			x := filepath.Join(DefaultCfg.GetString("output_dir"), p.RelativePath)

			f, err := os.Create(x)
			if err != nil {
				log.Fatalf("- could not create %s. %s (%s)", p.RelativePath, err, p.BuildDir)
			}
			defer f.Close()
			f.Write([]byte(p.RawContent))
			continue
		}

		dirPath := filepath.Join(p.BuildDir...)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			log.Fatalf("- could not create %s directory. %s", dirPath, err)
		}

		f, err := os.Create(filepath.Join(append(p.BuildDir, "index.html")...))
		if err != nil {
			log.Fatalf("- could not create index.html. %s", err)
		}

		defer f.Close()
		f.Write([]byte(p.Html))
	}

	// Copy static dir to output directory
	staticDir := DefaultCfg.GetString("static_dir")
	if internal.Exists(staticDir) {
		outputDir := DefaultCfg.GetString("output_dir")
		if err := internal.CopyDirectory(
			filepath.Join(RootPath, staticDir),
			filepath.Join(RootPath, outputDir, staticDir)); err != nil {
			panic(err)
		}
	}

	// Start local web server
	if DefaultCfg.Get("local_server").(bool) {
		port := DefaultCfg.GetString("server_port")
		http.Handle("/", http.FileServer(http.Dir(DefaultCfg.GetString("output_dir"))))
		log.Println("+ Belbo " + Version + ". Serving on http://localhost:" + port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalln(err)
		}
	}
}
