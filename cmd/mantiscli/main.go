// Copyright 2017, 2022 Tamás Gulácsi. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go get github.com/hooklift/gowsdl/cmd/gowsdl
//go:generate wget -O mantis.wsdl.raw -q "https://www.unosoft.hu/mantis/kobe/api/soap/mantisconnect.php?wsdl"
//go:generate iconv -f ISO-8859-2 -t UTF-8 mantis.wsdl.raw -o mantis.wsdl
//go:generate sh -c "sed -i -e '1{s/ISO-8859-1/UTF-8/}' mantis.wsdl"
//go:generate rm -f mantis.wsdl.raw
//go:generate mkdir -p mantisconnect
//go:generate mv mantis.wsdl mantisconnect/
//go:generate gowsdl -o mantisconnect.go -p mantisconnect mantisconnect/mantis.wsdl

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/term"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"

	"github.com/tgulacsi/go/globalctx"
	tterm "github.com/tgulacsi/go/term"
	"github.com/tgulacsi/mantis-soap"
	mantiscmd "github.com/tgulacsi/mantis-soap/cmd"
)

var (
	zl     = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	logger = zerologr.New(&zl)
)

func main() {
	if err := Main(); err != nil {
		logger.Error(err, "Main")
		os.Exit(1)
	}
}

func Main() error {
	var cl mantis.Client

	app := mantiscmd.App(&cl)
	appVerbose := app.FlagSet.Bool("v", false, "verbose logging")
	URL := app.FlagSet.String("mantis", "", "Mantis URL")
	username := app.FlagSet.String("user", os.Getenv("USER"), "Mantis user name")
	passwordEnv := app.FlagSet.String("password-env", "MC_PASSWORD", "Environment variable's name for the password")
	configFile := app.FlagSet.String("config", os.ExpandEnv("/home/$USER/.config/mantiscli.json"), "config file with the stored password")

	if err := app.Parse(os.Args[1:]); err != nil {
		return err
	}

	ctx, cancel := globalctx.Wrap(context.Background())
	defer cancel()

	passw := os.Getenv(*passwordEnv)
	var conf Config
	if passw == "" && *configFile != "" {
		var err error
		if conf, err = loadConfig(*configFile); err != nil {
			logger.Error(err, "load config", "file", *configFile)
		} else {
			passw = conf.Passwd[*username]
		}
	}

	u := *URL
	if passw == "" {
		fmt.Printf("Password for %q at %q: ", *username, u)
		if b, err := term.ReadPassword(0); err != nil {
			return fmt.Errorf("read password: %w", err)
		} else {
			passw = string(b)
			if conf.Passwd == nil {
				conf.Passwd = map[string]string{*username: passw}
			} else {
				conf.Passwd[*username] = passw
			}
		}
		fmt.Printf("\n")
	}
	var err error
	if cl, err = mantis.New(ctx, u, *username, passw); err != nil {
		cancel()
		return err
	}
	if *appVerbose {
		cl.Logger = logger.WithName("mantis-soap")
		mantis.SetLogger(cl.Logger)
	}
	if *configFile != "" {
		logger := logger.WithValues("file", configFile)
		_ = os.MkdirAll(filepath.Dir(*configFile), 0700)
		fh, err := os.OpenFile(*configFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			logger.Error(err, "create")
		} else {
			if err = json.NewEncoder(fh).Encode(conf); err != nil {
				logger.Error(err, "encode", "config", conf)
			} else if closeErr := fh.Close(); closeErr != nil {
				logger.Error(err, "close")
			}
		}
	}

	args := os.Args[1:]
	enc := tterm.GetTTYEncoding()
	for i, a := range args {
		var err error
		if args[i], err = enc.NewDecoder().String(a); err != nil {
			logger.Error(err, "Error decoding", "raw", a, "encoding", enc)
			args[i] = a
		}
	}
	//logger.Info("main", "args", args)

	return app.Run(ctx)
}

type Config struct {
	Passwd map[string]string
}

func loadConfig(file string) (Config, error) {
	var conf Config
	fh, err := os.Open(file)
	if err != nil {
		return conf, err
	}
	defer fh.Close()
	return conf, json.NewDecoder(fh).Decode(&conf)
}

// vim: set fileencoding=utf-8 noet:
