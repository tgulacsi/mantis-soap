// Copyright 2016, 2025 Tamás Gulácsi
//
// SPDX-License-Identifier: Apache-2.0

package mantis

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"unsafe"

	"github.com/UNO-SOFT/zlog/v2"
	"github.com/tgulacsi/go/soaphlp"
	"golang.org/x/net/publicsuffix"
)

var logger = slog.Default()

func SetLogger(lgr *slog.Logger) { logger = lgr }

func NewWithHTTPClient(ctx context.Context, c *http.Client, baseURL, username, password string) (Client, error) {
	select {
	case <-ctx.Done():
		return Client{}, ctx.Err()
	default:
	}
	if c == nil {
		c = http.DefaultClient
	}
	if c.Jar == nil {
		var err error
		if c.Jar, err = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List}); err != nil {
			return Client{}, err
		}
	}
	cl := Client{
		Caller: soaphlp.NewClient(
			baseURL+"/api/soap/mantisconnect.php",
			"http://www.mantisbt.org/bugs/api/soap/mantisconnect.php/",
			c,
		),
		auth: Auth{
			Username: username,
			Password: password,
		},
		httpClient: c, restURL: baseURL + "/api/rest/index.php",
	}
	var err error
	if cl.auth.IsAPIToken() {
		cl.User, err = cl.Me(ctx)
	} else {
		var resp LoginResponse
		if resp, err = cl.Login(ctx); err == nil {
			cl.User = resp.Return.Account
		}
	}
	return cl, err
}

func New(ctx context.Context, baseURL, username, password string) (Client, error) {
	return NewWithHTTPClient(ctx, nil, baseURL, username, password)
}

type Client struct {
	soaphlp.Caller
	httpClient *http.Client
	*slog.Logger
	User    AccountData
	auth    Auth
	restURL string
}

func (c Client) Call(ctx context.Context, method string, request, response interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if c.Caller == nil {
		panic("nil Caller")
	}
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	if err := xml.NewEncoder(buf).Encode(request); err != nil {
		return fmt.Errorf("marshal %#v: %w", request, err)
	}
	if zlog.SFromContext(ctx) == nil {
		ctx = zlog.NewSContext(ctx, c.Logger)
	}
	d, err := c.Caller.Call(ctx, method, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("call %s: %w", buf.String(), err)
	}
	buf.Reset()
	if err := d.Decode(response); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c Client) FilterSearchIssueIDs(ctx context.Context, filter FilterSearchData, pageNumber, perPage int) ([]int, error) {
	var resp FilterSearchIssueIDsResponse
	err := c.Call(ctx, "mc_filter_search_issue_ids",
		FilterSearchIssueIDsRequest{Auth: c.auth, Filter: filter,
			PageNumber: pageNumber, PerPage: perPage},
		&resp,
	)
	return *((*[]int)(unsafe.Pointer(&resp.IDs))), err
}

func (c Client) ProjectGetUsers(ctx context.Context, projectID, access int) ([]AccountData, error) {
	var resp ProjectGetUsersResponse
	err := c.Call(ctx, "mc_project_get_users",
		ProjectGetUsersRequest{Auth: c.auth, ProjectID: projectID, Access: access},
		&resp,
	)
	return resp.Users, err
}

func (c Client) ProjectsGetUserAccessible(ctx context.Context) ([]ProjectData, error) {
	var resp ProjectsGetUserAccessibleResponse
	err := c.Call(ctx, "mc_projects_get_user_accessible",
		ProjectsGetUserAccessibleRequest{Auth: c.auth},
		&resp)
	return resp.Projects, err
}

func (c Client) ProjectIssues(ctx context.Context, projectID, page, perPage int) ([]IssueData, error) {
	var resp ProjectIssuesResponse
	err := c.Call(ctx, "mc_project_get_issues",
		ProjectIssuesRequest{
			Auth:       c.auth,
			ProjectID:  projectID,
			PageNumber: page,
			PerPage:    perPage,
		}, &resp)
	return resp.Issues, err
}

func (c Client) IssueUpdate(ctx context.Context, issueID int, issue IssueData) (bool, error) {
	var resp IssueUpdateResponse
	iID := IssueID(issueID)
	issue.ID = &iID
	if err := c.Call(ctx, "mc_issue_update",
		IssueUpdateRequest{Auth: c.auth, IssueID: iID, Issue: issue},
		&resp,
	); err != nil {
		return false, err
	}
	return resp.Return, nil
}

func (c Client) IssueAdd(ctx context.Context, issue IssueData) (int, error) {
	var resp IssueAddResponse
	if err := c.Call(ctx, "mc_issue_add",
		IssueAddRequest{Auth: c.auth, Issue: issue},
		&resp,
	); err != nil {
		return 0, err
	}
	return resp.Return, nil
}

func (c Client) IssueAttachmentAdd(ctx context.Context, issueID int, name, fileType string, content io.Reader) (int, error) {
	var resp IssueAttachmentAddResponse
	if err := c.Call(ctx, "mc_issue_attachment_add",
		IssueAttachmentAddRequest{Auth: c.auth, IssueID: IssueID(issueID),
			Name: name, FileType: fileType,
			Content: Reader{content}},
		&resp); err != nil {
		return 0, err
	}
	return resp.Return, nil
}

func (c Client) IssueNoteAdd(ctx context.Context, issueID int, note IssueNoteData) (int, error) {
	var resp IssueNoteAddResponse
	if err := c.Call(ctx, "mc_issue_note_add",
		IssueNoteAddRequest{Auth: c.auth, IssueID: IssueID(issueID), Note: note},
		&resp,
	); err != nil {
		return 0, err
	}
	return resp.Return, nil
}

func (c Client) IssueGet(ctx context.Context, issueID int) (IssueData, error) {
	var resp IssueGetResponse
	if err := c.Call(ctx, "mc_issue_get",
		IssueGetRequest{Auth: c.auth, IssueID: IssueID(issueID)},
		&resp,
	); err != nil {
		return IssueData{}, err
	}
	return resp.Return, nil
}

func (c Client) IssueExists(ctx context.Context, issueID int) (bool, error) {
	var resp IssueExistsResponse
	if err := c.Call(ctx, "mc_issue_exists",
		IssueExistsRequest{Auth: c.auth, IssueID: IssueID(issueID)},
		&resp,
	); err != nil {
		return false, err
	}
	return resp.Return, nil
}

func (c Client) ProjectVersionsList(ctx context.Context, projectID int) ([]ProjectVersionData, error) {
	var resp ProjectGetVersionsResponse
	if err := c.Call(ctx, "mc_project_get_versions",
		ProjectGetVersionsRequest{Auth: c.auth, ProjectID: projectID},
		&resp,
	); err != nil {
		return nil, err
	}
	return resp.Return, nil
}
func (c Client) ProjectVersionAdd(ctx context.Context, projectID int, name, description string, released, obsolete bool, date *Time) (int, error) {
	var resp ProjectVersionAddResponse
	if err := c.Call(ctx, "mc_project_version_add",
		ProjectVersionAddRequest{Auth: c.auth,
			Version: ProjectVersionData{
				ProjectID: projectID,
				Name:      name, Description: description,
				Released: released, Obsolete: obsolete,
				DateOrder: date}},
		&resp,
	); err != nil {
		return 0, err
	}
	return resp.Return, nil
}
func (c Client) ProjectVersionUpdate(ctx context.Context, version ProjectVersionData) error {
	var resp ProjectVersionUpdateResponse
	return c.Call(ctx, "mc_project_version_updateRequest",
		ProjectVersionUpdateRequest{Auth: c.auth, VersionID: version.ID, Version: version},
		&resp,
	)
}
func (c Client) ProjectVersionDelete(ctx context.Context, versionID int) error {
	var resp ProjectVersionDeleteResponse
	return c.Call(ctx, "mc_project_version_deleteRequest",
		ProjectVersionDeleteRequest{Auth: c.auth, VersionID: versionID},
		&resp)
}

func (c Client) StatusEnum(ctx context.Context) ([]ObjectRef, error) {
	var resp StatusEnumResponse
	if err := c.Call(ctx, "mc_enum_status", StatusEnumRequest{Auth: c.auth}, &resp); err != nil {
		return nil, err
	}
	return resp.Statuses, nil
}

func (c Client) Login(ctx context.Context) (LoginResponse, error) {
	var resp LoginResponse
	return resp, c.Call(ctx, "mc_login", LoginRequest{Auth: c.auth}, &resp)
}

// GetCategoriesForProject - get the categories belonging to the specified project.
func (c Client) GetCategoriesForProject(ctx context.Context, projectID int) (ProjectCategoriesResp, error) {
	var resp ProjectCategoriesResp
	return resp, c.Call(ctx, "mc_project_get_categories", ProjectCategoriesReq{Auth: c.auth, ProjectID: projectID}, &resp)
}

func (c Client) CreateAPIToken(ctx context.Context, name string) (string, error) {
	b, err := json.Marshal(struct {
		Name string `json:"name"`
	}{Name: name})
	if err != nil {
		return "", err
	}
	var result struct {
		User  AccountData `json:"user"`
		Name  string      `json:"name"`
		Token string      `json:"token"`
		ID    int         `json:"id"`
	}
	err = c.restCall(ctx, &result, "POST", "/users/me/token/"+url.PathEscape(name),
		bytes.NewReader(b))
	return result.Token, err
}
func (c Client) DeleteAPIToken(ctx context.Context, name string) error {
	return c.restCall(ctx, nil, "DELETE", "/users/me/token/"+url.PathEscape(name), nil)
}
func (c Client) Me(ctx context.Context) (user AccountData, err error) {
	err = c.restCall(ctx, &user, "GET", "/users/me", nil)
	return user, err
}

func (c Client) restCall(ctx context.Context, response any, method, path string, body io.Reader) error {
	u := c.restURL + path
	if err := func() error {
		URL, err := url.Parse(u)
		if err != nil {
			return err
		}
		if !c.auth.IsAPIToken() {
			if URL.User == nil {
				URL.User = url.UserPassword(c.auth.Username, c.auth.Password)
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, URL.String(), body)
		if err != nil {
			return err
		}
		if c.auth.IsAPIToken() {
			req.Header.Add("Authorization", c.auth.Password)
		} else {
			req.Header.Add("Authorization", "Basic "+base64.URLEncoding.EncodeToString([]byte(c.auth.Username+":"+c.auth.Password)))
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("%s [%v]", resp.Status, req.Header)
		}
		if response == nil {
			return nil
		}
		return json.NewDecoder(resp.Body).Decode(response)
	}(); err != nil {
		return fmt.Errorf("%s: %w", u, err)
	}
	return nil

}

var bufPool = &bufferPool{
	Pool: sync.Pool{New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, 1024)) }},
}

type bufferPool struct {
	sync.Pool
}

func (p *bufferPool) Get() *bytes.Buffer {
	return p.Pool.Get().(*bytes.Buffer)
}
func (p *bufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	p.Pool.Put(b)
}

// vim: set fileencoding=utf-8 noet:
