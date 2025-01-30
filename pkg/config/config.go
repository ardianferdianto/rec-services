package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Option struct {
	Path        string
	ConfigName  string
	ConfigType  string
	EnvPrefix   string
	EnvReplacer *strings.Replacer
}

func New(options ...func(*Option)) *Option {
	opt := &Option{
		Path:        ".",
		ConfigName:  "config",
		ConfigType:  "yaml",
		EnvPrefix:   "",
		EnvReplacer: strings.NewReplacer(".", "_"),
	}
	for _, o := range options {
		o(opt)
	}
	return opt
}

func WithPath(path string) func(*Option) {
	return func(o *Option) {
		o.Path = path
	}
}

func (o *Option) Read(cfg interface{}) error {
	viper.AddConfigPath(o.Path)
	viper.SetConfigName(o.ConfigName)
	viper.SetConfigType(o.ConfigType)
	viper.SetEnvPrefix(o.EnvPrefix)
	viper.SetEnvKeyReplacer(o.EnvReplacer)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return viper.Unmarshal(cfg)
}
