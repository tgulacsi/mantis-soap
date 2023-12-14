// Copyright 2017, 2022 Tamás Gulácsi. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package mantiscmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tgulacsi/mantis-soap"
	"github.com/zRedShift/mimemagic"
)

var logger = slog.Default()

func SetLogger(lgr *slog.Logger) { logger = lgr }

// App returns an *ffcli.Command usable as app.
func App(cl *mantis.Client) *ffcli.Command {
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
					logger.Info("Attachment already there.", "file", fn)
					return nil
				}
			}
			fh, err := os.Open(fn)
			if err != nil {
				return err
			}
			defer fh.Close()

			t, err := mimemagic.MatchFile(fh)
			if err != nil {
				return err
			}
			if _, err = fh.Seek(0, 0); err != nil {
				return err
			}
			if _, err := cl.IssueAttachmentAdd(ctx, issueID, filepath.Base(fn), t.MediaType(), fh); err != nil {
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

	attachmentAddCmd := ffcli.Command{Name: "add", ShortHelp: "add attachment",
		Exec: addAttachmentCmd.Exec,
	}

	attachmentListCmd := ffcli.Command{Name: "list", ShortHelp: "list attachments",
		Exec: issueListAttachmentsCmd.Exec,
	}

	attachmentDownloadCmd := ffcli.Command{Name: "download", ShortHelp: "download attachments",
		Exec: issueDownloadAttachmentCmd.Exec,
	}

	attachmentCmd := ffcli.Command{Name: "attachment", ShortHelp: "do sth with attachments",
		Exec:        attachmentListCmd.Exec,
		Subcommands: []*ffcli.Command{&attachmentAddCmd, &attachmentListCmd, &attachmentDownloadCmd},
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
	statusCmd := ffcli.Command{Name: "status", ShortHelp: "set issue's status",
		Exec: func(ctx context.Context, args []string) error {
			status, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			issueIDs, err := toInts(args[1:])
			if err != nil {
				return err
			}
			for _, issueID := range issueIDs {
				issue, err := cl.IssueGet(ctx, issueID)
				if err != nil {
					return err
				}
				if issue.Status.ID >= status {
					fmt.Printf("SKIP %d (%d=%q)\n", issueID, issue.Status.ID, issue.Status.Name)
					continue
				}
				issue.Status.ID, issue.Status.Name = status, ""
				custFields := make([]mantis.CustomFieldData, 0, len(issue.CustomFields))
				for _, f := range issue.CustomFields {
					if f.Value != "" {
						custFields = append(custFields, f)
					}
				}
				issue.CustomFields = custFields
				if _, err = cl.IssueUpdate(ctx, issueID, issue); err != nil {
					return err
				}
			}
			return nil
		},
	}

	issueCmd := &ffcli.Command{Name: "issue", ShortUsage: "do sth on issues",
		Subcommands: []*ffcli.Command{
			existCmd, getIssuesCmd,
			getMonitorsCmd, addMonitorsCmd,
			addAttachmentCmd, issueListAttachmentsCmd, issueDownloadAttachmentCmd,
			&statusCmd,
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
			logger.Info("added", "note", noteID)
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
			if len(args) != 0 {
				m := make(map[string]int, len(projects))
				for i := range projects {
					m[projects[i].Name] = i
				}
				pp := make([]mantis.ProjectData, 0, len(args))
				for _, a := range args {
					if i, found := m[a]; !found {
						return fmt.Errorf("%q not found", a)
					} else {
						pp = append(pp, projects[i])
					}
				}
				projects = pp
			}
			return E(projects)
		},
	}

	var projectID int
	pVersionsListCmd := &ffcli.Command{Name: "list", ShortUsage: "list project versions <projectID>",
		Exec: func(ctx context.Context, args []string) error {
			versions, err := cl.ProjectVersionsList(ctx, projectID)
			enc := json.NewEncoder(os.Stdout)
			for _, v := range versions {
				if encErr := enc.Encode(v); encErr != nil && err == nil {
					err = encErr
				}
			}
			return err
		},
	}

	fs := flag.NewFlagSet("project-version-add", flag.ContinueOnError)
	pVersionsAddDescription := fs.String("description", "", "version description")
	pVersionsAddReleased := fs.Bool("released", false, "released?")
	pVersionsAddObsolete := fs.Bool("obsolete", false, "obsolete?")
	pVersionsAddDate := fs.String("date", "", "date")
	pVersionsAddCmd := &ffcli.Command{Name: "add", ShortUsage: "add project version", FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("version name is required")
			}
			var date *mantis.Time
			if *pVersionsAddDate != "" {
				date = new(mantis.Time)
				if err := date.UnmarshalText([]byte(*pVersionsAddDate)); err != nil {
					return err
				}
			}
			id, err := cl.ProjectVersionAdd(ctx, projectID, args[0], *pVersionsAddDescription, *pVersionsAddReleased, *pVersionsAddObsolete, date)
			fmt.Println(id)
			return err
		},
	}

	fs = flag.NewFlagSet("project-version-add", flag.ContinueOnError)
	pVersionsUpdateDescription := fs.String("description", "", "version description")
	pVersionsUpdateReleased := fs.Bool("released", false, "released?")
	pVersionsUpdateObsolete := fs.Bool("obsolete", false, "obsolete?")
	pVersionsUpdateDate := fs.String("date", "", "date")
	pVersionsUpdateCmd := &ffcli.Command{Name: "update",
		ShortUsage: "update project version <versionID> [name]",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("versionID is required")
			}
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			projects, err := cl.ProjectsGetUserAccessible(ctx)
			if err != nil {
				return err
			}
			var data mantis.ProjectVersionData
		OuterLoop:
			for _, p := range projects {
				versions, err := cl.ProjectVersionsList(ctx, p.ID)
				if err != nil {
					return err
				}
				for _, v := range versions {
					if v.ID == id {
						data = v
						break OuterLoop
					}
				}
			}
			if data.ID != id {
				return fmt.Errorf("versionID %d not found", id)
			}

			var date *mantis.Time
			if *pVersionsUpdateDate != "" {
				date = new(mantis.Time)
				if err := date.UnmarshalText([]byte(*pVersionsUpdateDate)); err != nil {
					return err
				}
			}
			var name string
			if len(args) > 1 {
				name = args[1]
			}
			data.ProjectID, data.ID = projectID, id
			if name != "" {
				data.Name = name
			}
			if *pVersionsUpdateDescription != "" {
				data.Description = *pVersionsUpdateDescription
			}
			if !date.IsZero() {
				data.DateOrder = date
			}
			data.Released = *pVersionsUpdateReleased
			data.Obsolete = *pVersionsUpdateObsolete
			logger.Debug("ProjectVersionUpdate", "data", data)
			return cl.ProjectVersionUpdate(ctx, data)
		},
	}

	pVersionsDeleteCmd := &ffcli.Command{Name: "delete",
		ShortUsage: "delete project version <versionID>",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("versionID is required")
			}
			versionID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return cl.ProjectVersionDelete(ctx, versionID)
		},
	}

	fs = flag.NewFlagSet("projects", flag.ContinueOnError)
	fs.IntVar(&projectID, "project", 0, "project id")
	projectVersionsCmd := &ffcli.Command{Name: "versions", ShortUsage: "do sth with versions",
		FlagSet:     fs,
		Subcommands: []*ffcli.Command{pVersionsListCmd, pVersionsAddCmd, pVersionsDeleteCmd, pVersionsUpdateCmd},
	}

	projectsCmd := &ffcli.Command{Name: "project", ShortUsage: "do sth with projects",
		Subcommands: []*ffcli.Command{listProjectsCmd, projectVersionsCmd},
	}

	fs = flag.NewFlagSet("project-list-users", flag.ContinueOnError)
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
	return &ffcli.Command{Name: "mantiscli", ShortUsage: "Mantis Command-Line Interface", FlagSet: fs,
		Subcommands: []*ffcli.Command{
			&attachmentCmd,
			issueCmd, noteCmd, projectsCmd, usersCmd},
	}
}

// E encodes the answer as JSON.
func E(answer interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ") // Go1.7
	if err := enc.Encode(answer); err != nil {
		logger.Error("ERROR encoding answer", "error", err)
		return err
	}
	return nil
}

func toInts(args []string) ([]int, error) {
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
func addMonitors(ctx context.Context, cl *mantis.Client, issueID int, plusMonitors []string) error {
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
			logger.Info("unknown user", "plusMonitor", p)
			continue
		}
		issue.Monitors = append(issue.Monitors, u)
		exists[p] = struct{}{}
		n++
	}
	if encErr := E(issue.Monitors); encErr != nil && err == nil {
		err = encErr
	}

	if n == 0 || err != nil {
		return err
	}
	issue.CustomFields, issue.Attachments, issue.Notes = nil, nil, nil
	_, err = cl.IssueUpdate(ctx, issueID, issue)
	return err
}
