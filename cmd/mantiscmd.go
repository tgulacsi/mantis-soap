// Copyright 2017, 2025 Tamás Gulácsi. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package mantiscmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/peterbourgon/ff/v4"
	"github.com/tgulacsi/mantis-soap"
	"github.com/titanous/json5"
	"github.com/zRedShift/mimemagic"
)

var logger = slog.Default()

func SetLogger(lgr *slog.Logger) { logger = lgr }

// App returns an *ff.Command usable as app.
func App(cl *mantis.Client) (*ff.Command, *ff.FlagSet) {
	existCmd := &ff.Command{Name: "exist", Usage: "check the existence of issues",
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
	getIssuesCmd := &ff.Command{Name: "get", Usage: "get",
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
	searchIssuesCmd := &ff.Command{Name: "search", Usage: "search",
		Exec: func(ctx context.Context, args []string) error {
			var filter mantis.FilterSearchData
			if err := json5.Unmarshal([]byte(strings.Join(args, " ")), &filter); err != nil {
				return fmt.Errorf("unmarshal %q as %#v: %w", args, filter, err)
			}
			ids, err := cl.FilterSearchIssueIDs(ctx, filter, 0, 1000)
			if err != nil {
				return err
			}
			slices.Sort(ids)
			return E(ids)
		},
	}

	getMonitorsCmd := &ff.Command{Name: "monitors", Usage: "get monitors",
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

	addAttachmentCmd := &ff.Command{Name: "attach", Usage: "attach a file to the issue",
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

	issueListAttachmentsCmd := &ff.Command{Name: "attachments", Usage: "list attachments",
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
	issueDownloadAttachmentCmd := &ff.Command{Name: "download", Usage: "download attachments of the issue",
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

	attachmentAddCmd := ff.Command{Name: "add", ShortHelp: "add attachment",
		Exec: addAttachmentCmd.Exec,
	}

	attachmentListCmd := ff.Command{Name: "list", ShortHelp: "list attachments",
		Exec: issueListAttachmentsCmd.Exec,
	}

	attachmentDownloadCmd := ff.Command{Name: "download", ShortHelp: "download attachments",
		Exec: issueDownloadAttachmentCmd.Exec,
	}

	attachmentCmd := ff.Command{Name: "attachment", ShortHelp: "do sth with attachments",
		Exec:        attachmentListCmd.Exec,
		Subcommands: []*ff.Command{&attachmentAddCmd, &attachmentListCmd, &attachmentDownloadCmd},
	}

	addMonitorsCmd := &ff.Command{Name: "add", Usage: "add monitor",
		Exec: func(ctx context.Context, args []string) error {
			issueID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return addMonitors(ctx, cl, issueID, args[1:])
		},
	}
	statusCmd := ff.Command{Name: "status", ShortHelp: "set issue's status",
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

	issueCmd := &ff.Command{Name: "issue", Usage: "do sth on issues",
		Subcommands: []*ff.Command{
			existCmd, getIssuesCmd, searchIssuesCmd,
			getMonitorsCmd, addMonitorsCmd,
			addAttachmentCmd, issueListAttachmentsCmd, issueDownloadAttachmentCmd,
			&statusCmd,
		},
	}

	addNoteCmd := &ff.Command{
		Name: "add", Usage: "add a note to an issue",
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
	noteCmd := &ff.Command{Name: "note", Usage: "do sth with notes",
		Subcommands: []*ff.Command{addNoteCmd},
	}

	listProjectsCmd := &ff.Command{Name: "list", Usage: "list projects",
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
	pVersionsListCmd := &ff.Command{Name: "list", Usage: "list project versions <projectID>",
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

	FS := ff.NewFlagSet("project-version-add")
	pVersionsAddDescription := FS.StringLong("description", "", "version description")
	pVersionsAddReleased := FS.BoolLongDefault("released", false, "released?")
	pVersionsAddObsolete := FS.BoolLongDefault("obsolete", false, "obsolete?")
	pVersionsAddDate := FS.StringLong("date", "", "date")
	pVersionsAddCmd := &ff.Command{Name: "add", Usage: "add project version", Flags: FS,
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

	FS = ff.NewFlagSet("project-version-add")
	pVersionsUpdateDescription := FS.StringLong("description", "", "version description")
	pVersionsUpdateReleased := FS.BoolLongDefault("released", false, "released?")
	pVersionsUpdateObsolete := FS.BoolLongDefault("obsolete", false, "obsolete?")
	pVersionsUpdateDate := FS.StringLong("date", "", "date")
	pVersionsUpdateCmd := &ff.Command{Name: "update",
		Usage: "update project version <versionID> [name]",
		Flags: FS,
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

	pVersionsDeleteCmd := &ff.Command{Name: "delete",
		Usage: "delete project version <versionID>",
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

	FS = ff.NewFlagSet("projects")
	FS.IntVar(&projectID, 0, "project", 0, "project id")
	projectVersionsCmd := &ff.Command{Name: "versions", Usage: "do sth with versions",
		Flags:       FS,
		Subcommands: []*ff.Command{pVersionsListCmd, pVersionsAddCmd, pVersionsDeleteCmd, pVersionsUpdateCmd},
	}

	projectsCmd := &ff.Command{Name: "project", Usage: "do sth with projects",
		Subcommands: []*ff.Command{listProjectsCmd, projectVersionsCmd},
	}

	FS = ff.NewFlagSet("project-list-users")
	usersAccessLevel := FS.IntLong("access-level", 10, "access level threshold")
	listUsersCmd := &ff.Command{Name: "list", Flags: FS,
		ShortHelp: "user list [projectID]",
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
	deleteAPITokenCmd := ff.Command{Name: "delete",
		Usage:     "delete <tokenName>",
		ShortHelp: "delete token",
		Exec: func(ctx context.Context, args []string) error {
			return cl.DeleteAPIToken(ctx, args[0])
		},
	}
	createAPITokenCmd := ff.Command{Name: "token",
		Usage:       "token [tokenName]",
		ShortHelp:   "create a new API token with the given name",
		Subcommands: []*ff.Command{&deleteAPITokenCmd},
		Exec: func(ctx context.Context, args []string) error {
			var name string
			if len(args) != 0 {
				name = args[0]
			}
			token, err := cl.CreateAPIToken(ctx, name)
			fmt.Println(token)
			return err
		},
	}
	usersCmd := &ff.Command{Name: "user", Usage: "do sth with users",
		Subcommands: []*ff.Command{listUsersCmd, &createAPITokenCmd},
	}

	FS = ff.NewFlagSet("mantiscli")
	return &ff.Command{Name: "mantiscli", Flags: FS,
		Usage: "Mantis Command-Line Interface",
		Subcommands: []*ff.Command{
			&attachmentCmd,
			issueCmd, noteCmd, projectsCmd, usersCmd},
	}, FS
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
