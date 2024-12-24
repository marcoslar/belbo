package internal

import "github.com/BurntSushi/toml"

// Config represents a toml snippet
type Config map[string]interface{}

func NewConfig(tomlConfig string, defaultCfg *Config) (*Config, error) {
	_, err := toml.Decode(tomlConfig, &defaultCfg)
	if err != nil {
		return nil, err
	}

	return defaultCfg, nil
}

func (cfg Config) MergeWith(b Config) Config {
	result := make(Config)
	for k, v := range cfg {
		result[k] = v
	}

	for k, v := range b {
		result[k] = v
	}
	return result
}

func (cfg Config) GetString(key string) string {
	if cfg.Get(key) != nil {
		return cfg.Get(key).(string)
	}

	return ""
}

func (cfg Config) GetStringSlice(key string) []string {
	var result []string
	values := cfg.Get(key)
	switch xx := values.(type) {
	case []string:
		for _, val := range xx {
			result = append(result, val)
		}
	case []interface{}:
		for _, val := range xx {
			result = append(result, val.(string))
		}
	}

	return result
}

func (cfg Config) Get(key string) interface{} {
	if val, ok := cfg[key]; ok {
		return val
	}
	return nil
}
