// Copyright 2015 Tamás Gulácsi
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

	"github.com/pkg/errors"
	"github.com/tgulacsi/go/soaphlp"
	"golang.org/x/net/context"
)

var Log = func(keyvals ...interface{}) error { return nil }

func New(ctx context.Context, baseURL, username, password string) (Client, error) {
	select {
	case <-ctx.Done():
		return Client{}, ctx.Err()
	default:
	}
	baseURL += "/api/soap/mantisconnect.php"
	cl := Client{
		caller: soaphlp.NewClient(baseURL, baseURL, nil),
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

type doer interface {
	Do(*http.Request) (*http.Response, error)
}
type caller interface {
	Call(context.Context, string, io.Reader) (*xml.Decoder, io.Closer, error)
}
type Client struct {
	caller
	auth Auth
	User AccountData
}

func (c Client) Call(ctx context.Context, method string, request, response interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	e := xml.NewEncoder(buf)
	if err := e.Encode(request); err != nil {
		return errors.Wrapf(err, "marshal %#v", request)
	}
	d, closer, err := c.caller.Call(ctx, method, bytes.NewReader(buf.Bytes()))
	if closer != nil {
		defer closer.Close()
	}
	if err != nil {
		return errors.Wrap(err, buf.String())
	}
	return d.Decode(response)
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

func (c Client) IssueUpdate(ctx context.Context, issueID int, issue IssueData) (bool, error) {
	var resp IssueUpdateResponse
	issue.ID = &issueID
	if err := c.Call(ctx, "mc_issue_update",
		IssueUpdateRequest{Auth: c.auth, IssueID: issueID, Issue: issue},
		&resp,
	); err != nil {
		return false, err
	}
	return resp.Return, nil
}

func (c Client) IssueAttachmentAdd(ctx context.Context, issueID int, name, fileType string, content io.Reader) (int, error) {
	var resp IssueAttachmentAddResponse
	if err := c.Call(ctx, "mc_issue_attachment_add",
		IssueAttachmentAddRequest{Auth: c.auth, IssueID: issueID,
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
		IssueNoteAddRequest{Auth: c.auth, IssueID: issueID, Note: note},
		&resp,
	); err != nil {
		return 0, err
	}
	return resp.Return, nil
}

func (c Client) IssueGet(ctx context.Context, issueID int) (IssueData, error) {
	var resp IssueGetResponse
	if err := c.Call(ctx, "mc_issue_get",
		IssueGetRequest{Auth: c.auth, IssueID: issueID},
		&resp,
	); err != nil {
		return IssueData{}, err
	}
	return resp.Return, nil
}

func (c Client) IssueExists(ctx context.Context, issueID int) (bool, error) {
	var resp IssueExistsResponse
	if err := c.Call(ctx, "mc_issue_exists", IssueExistsRequest{Auth: c.auth, IssueID: issueID}, &resp); err != nil {
		return false, err
	}
	return resp.Return, nil
}

func (c Client) Login(ctx context.Context) (LoginResponse, error) {
	var resp LoginResponse
	return resp, c.Call(ctx, "mc_login", LoginRequest{Auth: c.auth}, &resp)
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
