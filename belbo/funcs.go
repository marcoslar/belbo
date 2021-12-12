package belbo

import (
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

	var belboFuncs *map[string]interface{}
	belboFuncs, ok := belboFuncsSymbol.(*map[string]interface{})
	if !ok {
		log.Printf("- BelboFuncs must be of type map[string]interface{}")
		return emptyResult
	}

	var funcsKeys []string
	for k := range *belboFuncs {
		funcsKeys = append(funcsKeys, k)
	}

	log.Printf("+ loading custom funcs %v", funcsKeys)

	return *belboFuncs
}
