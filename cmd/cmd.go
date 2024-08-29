package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	GithubAccessToken      string
	ManifestTemplateFile   string
	K8sCommands            []string
	CloseLastIssue         bool
	CheckBeforeCreateIssue *string `mapstructure:"check-before-create-issue"`

	Verbose bool
	Debug   bool
}

func Run() error {
	ctx := context.TODO()

	cfg, err := NewConfig()
	if err != nil {
		return errors.WithStack(err)
	}

	l, err := NewLogger(cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	defer l.Sync()

	zap.ReplaceGlobals(l)
	cfgJSON, _ := json.Marshal(cfg)
	zap.L().Debug("config", zap.String("config", string(cfgJSON)))

	cmd, err := InitializeCmd(ctx, cfg)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := cmd.Execute(); err != nil {
		zap.L().Debug("error", zap.String("stack trace", fmt.Sprintf("%+v\n", err)))
		return errors.WithStack(err)
	}
	return nil
}

func NewLogger(cfg Config) (*zap.Logger, error) {
	zcfg := zap.NewProductionConfig()
	if cfg.Debug {
		zcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	if cfg.Verbose {
		zcfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	l, err := zcfg.Build()
	return l, errors.WithStack(err)
}

func NewConfig() (Config, error) {
	pflag.StringP("GithubAccessToken", "", "", "Github Access Token")
	pflag.StringP("ManifestTemplateFile", "", "./template.tpl", "k8s manifest file template path")
	pflag.StringP("K8sCommands", "", "", "k8scommands for apply")
	pflag.BoolP("CloseLastIssue", "c", false, "Close last issue")
	pflag.BoolP("verbose", "v", false, "")
	pflag.BoolP("debug", "d", false, "")
	pflag.StringP("check-before-create-issue", "", "", "")

	viper.AutomaticEnv()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return Config{}, errors.WithStack(err)
	}

	var cfg Config
	pflag.Parse()
	err = viper.Unmarshal(&cfg)
	return cfg, errors.WithStack(err)
}
