// Copyright 2016, 2020 Tamás Gulácsi
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package mantis

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"sync"
	"unsafe"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/tgulacsi/go/soaphlp"
	"golang.org/x/net/context"
)

var Log = func(keyvals ...interface{}) error { return nil }

func NewWithHTTPClient(ctx context.Context, c *http.Client, baseURL, username, password string) (Client, error) {
	select {
	case <-ctx.Done():
		return Client{}, ctx.Err()
	default:
	}
	baseURL += "/api/soap/mantisconnect.php"
	cl := Client{
		Caller: soaphlp.NewClient(baseURL, baseURL, c),
		auth: Auth{
			Username: username,
			Password: password,
		},
	}
	resp, err := cl.Login(ctx)
	if err != nil {
		return cl, err
	}
	cl.User = resp.Return.Account
	return cl, nil
}

func New(ctx context.Context, baseURL, username, password string) (Client, error) {
	return NewWithHTTPClient(ctx, nil, baseURL, username, password)
}

type Client struct {
	soaphlp.Caller
	auth Auth
	User AccountData
	log.Logger
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
		return errors.Wrapf(err, "marshal %#v", request)
	}
	resp := bufPool.Get()
	defer bufPool.Put(resp)
	if c.Logger != nil {
		ctx = soaphlp.WithLog(ctx, c.Logger.Log)
	}
	d, err := c.Caller.Call(ctx, resp, method, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return errors.Wrap(err, buf.String())
	}
	buf.Reset()
	if err := d.Decode(response); err != nil {
		return errors.Wrap(err, resp.String())
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
