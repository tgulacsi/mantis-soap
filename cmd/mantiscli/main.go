// Copyright 2017 Tamás Gulácsi. All rights reserved.

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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
	"gopkg.in/h2non/filetype.v1"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"

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
	app := kingpin.New("mantiscli", "Mantis Command-Line Interface")
	appVerbose := app.Flag("verbose", "verbose logging").Short('v').Bool()

	URL := app.Flag("mantis", "Mantis URL").URL()
	username := app.Flag("user", "Mantis user name").Short('u').Default(os.Getenv("USER")).String()
	passwordEnv := app.Flag("password-env", "Environment variable's name for the password").
		Default("MC_PASSWORD").
		String()
	configFile := app.Flag("config", "config file with the stored password").
		Default(os.ExpandEnv("/home/$USER/.config/mantiscli.json")).
		String()

	issueCmd := app.Command("issue", "do sth on issues").Alias("issues")
	existsCmd := issueCmd.Command("exists", "check the existence of issues")
	existsIssueIDs := existsCmd.Arg("issueid", "Issue IDs to check").Ints()

	getIssuesCmd := issueCmd.Command("get", "get")
	getIssueIDs := getIssuesCmd.Arg("issueid", "Issue IDs to check").Ints()
	getMonitorsCmd := issueCmd.Command("monitors", "get monitors").Alias("list_monitors")
	getMonitorsIssueIDs := getMonitorsCmd.Arg("issueid", "Issue IDs to get monitors of").Ints()
	addMonitorsCmd := issueCmd.Command("addmonitor", "add monitor").Alias("add_monitor")
	addMonitorsIssueID := addMonitorsCmd.Arg("issueid", "Issue ID to add monitor to").Int()
	plusMonitors := addMonitorsCmd.Arg("plus_monitors", "names of plus monitors").Strings()
	addAttachmentCmd := issueCmd.Command("attach", "attach a file to the issue")
	addAttachmentIssueID := addAttachmentCmd.Arg("issueid", "Issue ID to attach the file to").Int()
	addAttachmentFile := addAttachmentCmd.Arg("file", "file to attach").ExistingFile()
	issueListAttachmentsCmd := issueCmd.Command("attachments", "list attachments")
	issueListAttachmentsIssueID := issueListAttachmentsCmd.Arg("issueid", "issue to list attachments of").Int()
	issueDownloadAttachmentCmd := issueCmd.Command("download", "download attachemnts of the issue")
	issueDownloadAttachmentIssueID := issueDownloadAttachmentCmd.Arg("issueid", "issue to download attachments of").Int()

	attachmentCmd := app.Command("attachment", "do sth with attachments")
	attachmentAddCmd := attachmentCmd.Command("add", "add attachment")
	attachmentAddIssueID := attachmentAddCmd.Arg("issueid", "issue ID to attach to").Int()
	attachmentAddFile := attachmentAddCmd.Arg("file", "file to attach").ExistingFile()
	attachmentListCmd := attachmentCmd.Command("list", "list attachments").Default()
	attachmentListIssueID := attachmentListCmd.Arg("issueid", "issueID to list attachments of").Int()
	attachmentDownloadCmd := attachmentCmd.Command("download", "download attachments")
	attachmentDownloadIssueID := attachmentDownloadCmd.Arg("issueid", "issueID to download attachments of").Int()

	monitorsCmd := app.Command("monitor", "do sth with monitors").Alias("monitors")
	listMonitorsCmd := monitorsCmd.Command("list", "list monitors").Default()
	listMonitorsIssueID := listMonitorsCmd.Arg("issueid", "Issue ID to list monitor of").Int()
	addMonitors2 := monitorsCmd.Command("add", "add monitor")
	addMonitors2IssueID := addMonitors2.Arg("issueid", "Issue ID to add monitor to").Int()
	plusMonitors2 := addMonitors2.Arg("plus_monitors", "names of plus monitors").Strings()

	noteCmd := app.Command("note", "do sth with notes").Alias("notes")
	addNoteCmd := noteCmd.Command("add", "add a note to an issue").Default()
	addNoteIssueID := addNoteCmd.Arg("issueid", "Issue ID to add the note to").Int()
	addNoteText := addNoteCmd.Arg("text", "text to add as note").Strings()

	projectsCmd := app.Command("project", "do sth with projects").Alias("projects")
	listProjectsCmd := projectsCmd.Command("list", "list projects").Default()

	projectVersionsCmd := projectsCmd.Command("versions", "do sth with versions").Alias("version")
	pVersionsListCmd := projectVersionsCmd.Command("list", "list project versions")
	pVersionsListProjectID := pVersionsListCmd.Arg("project-id", "project ID").Required().Int()

	pVersionsAddCmd := projectVersionsCmd.Command("add", "add project version")
	pVersionsAddName := pVersionsAddCmd.Arg("name", "version name").String()
	pVersionsAddProjectID := pVersionsAddCmd.Flag("project", "project id").Required().Int()
	pVersionsAddDescription := pVersionsAddCmd.Flag("description", "version description").String()
	pVersionsAddReleased := pVersionsAddCmd.Flag("released", "released?").Bool()
	pVersionsAddObsolete := pVersionsAddCmd.Flag("obsolete", "obsolete?").Bool()

	pVersionsDeleteCmd := projectVersionsCmd.Command("delete", "delete project version")
	pVersionsDeleteVersionID := pVersionsDeleteCmd.Arg("version-id", "version id").Int()

	usersCmd := app.Command("users", "do sth with users").Alias("user")
	listUsersCmd := usersCmd.Command("list", "list users").Default()
	usersProjectID := listUsersCmd.Arg("project", "project ID").Default("1").Int()
	usersAccessLevel := listUsersCmd.Flag("access-level", "access level threshold").Default("10").Int()

	var (
		cl     mantis.Client
		ctx    context.Context
		cancel context.CancelFunc
	)

	app.Action(kingpin.Action(func(aCtx *kingpin.ParseContext) error {
		if aCtx == nil || aCtx.SelectedCommand == nil {
			return nil
		}
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

		var u string
		if u2 := *URL; u2 != nil {
			u = u2.String()
		}
		if passw == "" {
			fmt.Printf("Password for %q at %q: ", *username, u)
			if b, err := terminal.ReadPassword(0); err != nil {
				return errors.Wrap(err, "read password")
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
		ctx, cancel = C(30)
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
		return nil
	}))

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

	switch cmd := kingpin.MustParse(app.Parse(args)); cmd {
	case existsCmd.FullCommand():
		answer := make(map[string]interface{}, len(*existsIssueIDs))
		for _, i := range *existsIssueIDs {
			exists, err := cl.IssueExists(ctx, i)
			if err != nil {
				return err
			}
			answer[strconv.Itoa(i)] = exists
		}
		E(answer)

	case getMonitorsCmd.FullCommand(), listMonitorsCmd.FullCommand():
		ids := *getMonitorsIssueIDs
		if cmd == listMonitorsCmd.FullCommand() {
			ids = []int{*listMonitorsIssueID}
		}
		answer := make(map[string]interface{}, len(ids))
		for _, i := range ids {
			issue, err := cl.IssueGet(ctx, i)
			if err != nil {
				return err
			}
			answer[strconv.Itoa(i)] = issue.Monitors
		}
		E(answer)

	case getIssuesCmd.FullCommand():
		answer := make(map[string]interface{}, len(*getIssueIDs))
		for _, i := range *getIssueIDs {
			issue, err := cl.IssueGet(ctx, i)
			if err != nil {
				return err
			}
			answer[strconv.Itoa(i)] = issue
		}
		E(answer)

	case addNoteCmd.FullCommand():
		ctx, cancel := C(10)
		defer cancel()
		noteID, err := cl.IssueNoteAdd(ctx, *addNoteIssueID, mantis.IssueNoteData{
			Reporter: cl.User,
			Text:     strings.Join(*addNoteText, " "),
		})
		if err != nil {
			return err
		}
		logger.Log("msg", "added", "note", noteID)

	case listProjectsCmd.FullCommand():
		ctx, cancel := C(10)
		defer cancel()
		projects, err := cl.ProjectsGetUserAccessible(ctx)
		if err != nil {
			return err
		}
		E(projects)

	case listUsersCmd.FullCommand():
		ctx, cancel := C(10)
		defer cancel()
		users, err := cl.ProjectGetUsers(ctx, *usersProjectID, *usersAccessLevel)
		if err != nil {
			return err
		}
		E(users)

	case addMonitorsCmd.FullCommand(), addMonitors2.FullCommand():
		issueID, plus := *addMonitorsIssueID, *plusMonitors
		if cmd == addMonitors2.FullCommand() {
			issueID, plus = *addMonitors2IssueID, *plusMonitors2
		}
		ctx, cancel := C(10)
		err := addMonitors(ctx, cl, issueID, plus)
		cancel()
		return err

	case addAttachmentCmd.FullCommand(), attachmentAddCmd.FullCommand():
		issueID, fn := *addAttachmentIssueID, *addAttachmentFile
		if issueID == 0 {
			issueID, fn = *attachmentAddIssueID, *attachmentAddFile
		}
		ctx, cancel := C(30)
		defer cancel()
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
			return errors.Wrapf(err, "add attachment %q", fn)
		}
		return nil

	case issueListAttachmentsCmd.FullCommand(), attachmentListCmd.FullCommand():
		issueID := *issueListAttachmentsIssueID
		if issueID == 0 {
			issueID = *attachmentListIssueID
		}
		ctx, cancel := C(10)
		defer cancel()
		issue, err := cl.IssueGet(ctx, issueID)
		if err != nil {
			return err
		}
		E(issue.Attachments)

	case issueDownloadAttachmentCmd.FullCommand(), attachmentDownloadCmd.FullCommand():
		issueID := *issueDownloadAttachmentIssueID
		if issueID == 0 {
			issueID = *attachmentDownloadIssueID
		}
		ctx, cancel := C(30)
		defer cancel()
		issue, err := cl.IssueGet(ctx, issueID)
		if err != nil {
			return err
		}
		for _, att := range issue.Attachments {
			E(att.DownloadURL)
		}

	case pVersionsListCmd.FullCommand():
		ctx, cancel := C(10)
		defer cancel()
		versions, err := cl.ProjectVersionsList(ctx, *pVersionsListProjectID)
		enc := json.NewEncoder(os.Stdout)
		for _, v := range versions {
			enc.Encode(v)
		}
		return err

	case pVersionsAddCmd.FullCommand():
		ctx, cancel := C(10)
		defer cancel()
		id, err := cl.ProjectVersionAdd(ctx, *pVersionsAddProjectID, *pVersionsAddName, *pVersionsAddDescription, *pVersionsAddReleased, *pVersionsAddObsolete, nil)
		fmt.Println(id)
		return err

	case pVersionsDeleteCmd.FullCommand():
		ctx, cancel := C(10)
		defer cancel()
		return cl.ProjectVersionDelete(ctx, *pVersionsDeleteVersionID)
	}

	return nil
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

func C(seconds int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
}

// vim: set fileencoding=utf-8 noet:
