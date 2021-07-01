// Copyright 2017, 2021 Tamás Gulácsi. All rights reserved.
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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
	"gopkg.in/h2non/filetype.v1"

	"github.com/go-kit/kit/log"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/tgulacsi/go/globalctx"
	"github.com/tgulacsi/go/loghlp/kitloghlp"
	"github.com/tgulacsi/go/term"
	"github.com/tgulacsi/mantis-soap"
)

var logger = kitloghlp.New(os.Stderr)

func main() {
	if err := Main(); err != nil {
		logger.Log("error", err)
		os.Exit(1)
	}
}

func Main() error {
	var cl mantis.Client

	toInts := func(args []string) ([]int, error) {
		var firstErr error
		ints := make([]int, 0, len(args))
		for _, a := range args {
			i, err := strconv.Atoi(a)
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("%q: %w", a, err)
				}
				continue
			}
			ints = append(ints, i)
		}
		return ints, firstErr
	}

	existCmd := &ffcli.Command{Name: "exist", ShortUsage: "check the existence of issues",
		Exec: func(ctx context.Context, args []string) error {
			issueIDs, err := toInts(args)
			if err != nil {
				return err
			}
			answer := make(map[string]interface{}, len(issueIDs))
			for _, i := range issueIDs {
				exists, err := cl.IssueExists(ctx, i)
				if err != nil {
					return err
				}
				answer[strconv.Itoa(i)] = exists
			}
			return E(answer)
		},
	}
	getIssuesCmd := &ffcli.Command{Name: "get", ShortUsage: "get",
		Exec: func(ctx context.Context, args []string) error {
			issueIDs, err := toInts(args)
			if err != nil {
				return err
			}
			answer := make(map[string]interface{}, len(issueIDs))
			for _, i := range issueIDs {
				issue, err := cl.IssueGet(ctx, i)
				if err != nil {
					return err
				}
				answer[strconv.Itoa(i)] = issue
			}
			return E(answer)
		},
	}

	getMonitorsCmd := &ffcli.Command{Name: "monitors", ShortUsage: "get monitors",
		Exec: func(ctx context.Context, args []string) error {
			issueIDs, err := toInts(args)
			if err != nil {
				return err
			}
			answer := make(map[string]interface{}, len(issueIDs))
			for _, i := range issueIDs {
				issue, err := cl.IssueGet(ctx, i)
				if err != nil {
					return err
				}
				answer[strconv.Itoa(i)] = issue.Monitors
			}
			return E(answer)
		},
	}

	addAttachmentCmd := &ffcli.Command{Name: "attach", ShortUsage: "attach a file to the issue",
		Exec: func(ctx context.Context, args []string) error {
			issueID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			fn := args[1]

			issue, err := cl.IssueGet(ctx, issueID)
			if err != nil {
				return err
			}
			for _, at := range issue.Attachments {
				if at.FileName == fn && at.Size > 0 {
					logger.Log("msg", "Attachment already there.", "file", fn)
					return nil
				}
			}
			fh, err := os.Open(fn)
			if err != nil {
				return err
			}
			defer fh.Close()

			t, err := filetype.MatchReader(fh)
			if err != nil {
				return err
			}
			if _, err = fh.Seek(0, 0); err != nil {
				return err
			}
			if _, err := cl.IssueAttachmentAdd(ctx, issueID, filepath.Base(fn), t.MIME.Value, fh); err != nil {
				return fmt.Errorf("add attachment %q: %w", fn, err)
			}
			return nil
		},
	}

	issueListAttachmentsCmd := &ffcli.Command{Name: "attachments", ShortUsage: "list attachments",
		Exec: func(ctx context.Context, args []string) error {
			issueID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			issue, err := cl.IssueGet(ctx, issueID)
			if err != nil {
				return err
			}
			return E(issue.Attachments)
		},
	}
	issueDownloadAttachmentCmd := &ffcli.Command{Name: "download", ShortUsage: "download attachments of the issue",
		Exec: func(ctx context.Context, args []string) error {
			issueID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			issue, err := cl.IssueGet(ctx, issueID)
			if err != nil {
				return err
			}
			for _, att := range issue.Attachments {
				return E(att.DownloadURL)
			}

			return nil
		},
	}

	addMonitorsCmd := &ffcli.Command{Name: "add", ShortUsage: "add monitor",
		Exec: func(ctx context.Context, args []string) error {
			issueID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return addMonitors(ctx, cl, issueID, args[1:])
		},
	}

	issueCmd := &ffcli.Command{Name: "issue", ShortUsage: "do sth on issues",
		Subcommands: []*ffcli.Command{
			existCmd, getIssuesCmd,
			getMonitorsCmd, addMonitorsCmd,
			addAttachmentCmd, issueListAttachmentsCmd, issueDownloadAttachmentCmd,
		},
	}

	addNoteCmd := &ffcli.Command{
		Name: "add", ShortUsage: "add a note to an issue",
		Exec: func(ctx context.Context, args []string) error {
			issueID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			args = args[1:]
			noteID, err := cl.IssueNoteAdd(ctx, issueID, mantis.IssueNoteData{
				Reporter: cl.User,
				Text:     strings.Join(args, " "),
			})
			if err != nil {
				return err
			}
			logger.Log("msg", "added", "note", noteID)
			return nil
		},
	}
	noteCmd := &ffcli.Command{Name: "note", ShortUsage: "do sth with notes",
		Subcommands: []*ffcli.Command{addNoteCmd},
	}

	listProjectsCmd := &ffcli.Command{Name: "list", ShortUsage: "list projects",
		Exec: func(ctx context.Context, args []string) error {
			projects, err := cl.ProjectsGetUserAccessible(ctx)
			if err != nil {
				return err
			}
			return E(projects)
		},
	}

	pVersionsListCmd := &ffcli.Command{Name: "list", ShortUsage: "list project versions",
		Exec: func(ctx context.Context, args []string) error {
			projectID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			versions, err := cl.ProjectVersionsList(ctx, projectID)
			enc := json.NewEncoder(os.Stdout)
			for _, v := range versions {
				enc.Encode(v)
			}
			return err
		},
	}

	fs := flag.NewFlagSet("project-version-add", flag.ContinueOnError)
	pVersionsAddProjectID := fs.Int("project", 0, "project id")
	pVersionsAddDescription := fs.String("description", "", "version description")
	pVersionsAddReleased := fs.Bool("released", false, "released?")
	pVersionsAddObsolete := fs.Bool("obsolete", false, "obsolete?")
	pVersionsAddCmd := &ffcli.Command{Name: "add", ShortUsage: "add project version", FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			id, err := cl.ProjectVersionAdd(ctx, *pVersionsAddProjectID, args[0], *pVersionsAddDescription, *pVersionsAddReleased, *pVersionsAddObsolete, nil)
			fmt.Println(id)
			return err
		},
	}

	pVersionsDeleteCmd := &ffcli.Command{Name: "delete", ShortUsage: "delete project version",
		Exec: func(ctx context.Context, args []string) error {
			projectID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return cl.ProjectVersionDelete(ctx, projectID)
		},
	}

	projectVersionsCmd := &ffcli.Command{Name: "versions", ShortUsage: "do sth with versions",
		Subcommands: []*ffcli.Command{pVersionsListCmd, pVersionsAddCmd, pVersionsDeleteCmd},
	}

	projectsCmd := &ffcli.Command{Name: "project", ShortUsage: "do sth with projects",
		Subcommands: []*ffcli.Command{listProjectsCmd, projectVersionsCmd},
	}

	usersAccessLevel := fs.Int("access-level", 10, "access level threshold")
	listUsersCmd := &ffcli.Command{Name: "list", ShortUsage: "list users", FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			projectID := 1
			if len(args) != 0 {
				var err error
				if projectID, err = strconv.Atoi(args[0]); err != nil {
					return err
				}
			}
			users, err := cl.ProjectGetUsers(ctx, projectID, *usersAccessLevel)
			if err != nil {
				return err
			}
			return E(users)
		},
	}
	usersCmd := &ffcli.Command{Name: "user", ShortUsage: "do sth with users",
		Subcommands: []*ffcli.Command{listUsersCmd},
	}

	fs = flag.NewFlagSet("mantiscli", flag.ContinueOnError)
	appVerbose := fs.Bool("v", false, "verbose logging")
	URL := fs.String("mantis", "", "Mantis URL")
	username := fs.String("user", os.Getenv("USER"), "Mantis user name")
	passwordEnv := fs.String("password-env", "MC_PASSWORD", "Environment variable's name for the password")
	configFile := fs.String("config", os.ExpandEnv("/home/$USER/.config/mantiscli.json"), "config file with the stored password")

	app := ffcli.Command{Name: "mantiscli", ShortUsage: "Mantis Command-Line Interface", FlagSet: fs,
		Subcommands: []*ffcli.Command{issueCmd, noteCmd, projectsCmd, usersCmd},
	}

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
			logger.Log("msg", "load config", "file", *configFile, "error", err)
		} else {
			passw = conf.Passwd[*username]
		}
	}

	u := *URL
	if passw == "" {
		fmt.Printf("Password for %q at %q: ", *username, u)
		if b, err := terminal.ReadPassword(0); err != nil {
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
		cl.Logger = log.With(logger, "lib", "mantis-soap")
	}
	if *configFile != "" {
		Log := log.With(logger, "file", configFile).Log
		os.MkdirAll(filepath.Dir(*configFile), 0700)
		fh, err := os.OpenFile(*configFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			Log("msg", "create", "error", err)
		} else {
			if err = json.NewEncoder(fh).Encode(conf); err != nil {
				Log("msg", "encode", "config", conf, "error", err)
			} else if closeErr := fh.Close(); closeErr != nil {
				Log("msg", "close", "error", err)
			}
		}
	}

	args := os.Args[1:]
	enc := term.GetTTYEncoding()
	for i, a := range args {
		var err error
		if args[i], err = enc.NewDecoder().String(a); err != nil {
			logger.Log("msg", "Error decoding", "raw", a, "encoding", enc, "error", err)
			args[i] = a
		}
	}
	//logger.Log("args", args)

	return app.Run(ctx)
}

func addMonitors(ctx context.Context, cl mantis.Client, issueID int, plusMonitors []string) error {
	issue, err := cl.IssueGet(ctx, issueID)
	if err != nil {
		return err
	}
	exists := make(map[string]struct{}, len(issue.Monitors)+len(plusMonitors))

	users, err := cl.ProjectGetUsers(ctx, 1, 0)
	if err != nil {
		return err
	}
	uM := make(map[string]mantis.AccountData, len(users))
	for _, u := range users {
		uM[u.Name] = u
	}

	for _, m := range issue.Monitors {
		exists[m.Name] = struct{}{}
	}
	var n int
	for _, p := range plusMonitors {
		if _, ok := exists[p]; ok {
			continue
		}
		u, ok := uM[p]
		if !ok {
			logger.Log("error", "unknown user "+p)
			continue
		}
		issue.Monitors = append(issue.Monitors, u)
		exists[p] = struct{}{}
		n++
	}
	E(issue.Monitors)

	if n == 0 {
		return nil
	}
	issue.CustomFields, issue.Attachments, issue.Notes = nil, nil, nil
	_, err = cl.IssueUpdate(ctx, issueID, issue)
	return err
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

func E(answer interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ") // Go1.7
	if err := enc.Encode(answer); err != nil {
		logger.Log("msg", "ERROR encoding answer", "error", err)
		return err
	}
	return nil
}

// vim: set fileencoding=utf-8 noet:
