package internal

import (
	"html/template"
	"log"
	"os/exec"
	"path/filepath"
	"plugin"
)

const PluginsFilename = "plugins.so"

func LoadFuncsAsPlugins(pluginsPath string) map[string]interface{} {
	emptyResult := make(map[string]interface{})

	if pluginsPath == "" {
		return emptyResult
	}

	pluginFilepath := filepath.Join(pluginsPath, PluginsFilename)
	// Build the plugin
	// go build -buildmode=plugin -o plugins/plugins.so plugins/plugins.go
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginFilepath, "plugins/plugins.go")
	err := cmd.Run()
	if err != nil {
		log.Printf("- could not build plugins. %s", err)
	}

	plugs, err := plugin.Open(pluginFilepath)
	if err != nil {
		log.Printf("- could not load plugins. %s", err)
		return emptyResult
	}

	belboFuncsSymbol, err := plugs.Lookup("BelboFuncs")
	if err != nil {
		log.Printf("- could not find the symbol BelboFuncs (%s)", err)
		return emptyResult
	}

	var belboFuncs *template.FuncMap
	belboFuncs, ok := belboFuncsSymbol.(*template.FuncMap)
	if !ok {
		log.Printf("- BelboFuncs must be of type template.FuncMap, type is %T instead", belboFuncs)
		return emptyResult
	}

	var funcsKeys []string
	for k := range *belboFuncs {
		funcsKeys = append(funcsKeys, k)
	}

	log.Printf("+ loading custom funcs %v", funcsKeys)

	return *belboFuncs
}
