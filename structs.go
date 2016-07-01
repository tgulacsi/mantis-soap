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

import "encoding/xml"

// https://www.unosoft.hu/mantis/kobe/api/soap/mantisconnect.php?wsdl

type ProjectGetUsersRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_project_get_users"`
	Auth
	ProjectID int `xml:"project_id"`
	Access    int `xml:"access"`
}

type ProjectGetUsersResponse struct {
	XMLName xml.Name      `xml:"http://futureware.biz/mantisconnect mc_project_get_usersResponse"`
	Users   []AccountData `xml:"return>item"`
}

type ProjectsGetUserAccessibleRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_projects_get_user_accessible"`
	Auth
}

type ProjectsGetUserAccessibleResponse struct {
	XMLName  xml.Name      `xml:"http://futureware.biz/mantisconnect mc_projects_get_user_accessibleResponse"`
	Projects []ProjectData `xml:"return>item"`
}

type IssueUpdateRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_update"`
	Auth
	IssueID int       `xml:"issueId"`
	Issue   IssueData `xml:"issue"`
}

type IssueUpdateResponse struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_updateResponse"`
	Return  bool     `xml:"return"`
}

type IssueAttachmentAddRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_attachment_add"`
	Auth
	IssueID  int    `xml:"issue_id"`
	Name     string `xml:"name"`
	FileType string `xml:"file_type"`
	Content  Reader `xml:"content"`
}

type IssueAttachmentAddResponse struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_attachment_addResponse"`
	Return  int      `xml:"return"`
}

type IssueNoteAddRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_note_add"`
	Auth
	IssueID int           `xml:"issue_id"`
	Note    IssueNoteData `xml:"note"`
}
type IssueNoteAddResponse struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_note_addResponse"`
	Return  int      `xml:"return"`
}

type IssueGetRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_get"`
	Auth
	IssueID int `xml:"issue_id"`
}

type IssueGetResponse struct {
	XMLName xml.Name  `xml:"http://futureware.biz/mantisconnect mc_issue_getResponse"`
	Return  IssueData `xml:"return"`
}

type IssueAddRequest struct {
	XMLName xml.Name  `xml:"http://futureware.biz/mantisconnect mc_issue_add"`
	Issue   IssueData `xml:"issue"`
}
type IssueExistsRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_exists"`
	Auth
	IssueID int `xml:"issue_id"`
}
type IssueExistsResponse struct {
	Return bool `xml:"return"`
}

type IssueAddResponse struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_issue_addResponse"`
	Return  int      `xml:"return"`
}

type LoginRequest struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_login"`
	Auth
}
type LoginResponse struct {
	XMLName xml.Name `xml:"http://futureware.biz/mantisconnect mc_loginResponse"`
	Return  UserData `xml:"return"`
}

type IssueNoteData struct {
	ID            *int        `xml:"id,omitempty"`
	Reporter      AccountData `xml:"reporter,omitempty"`
	Text          string      `xml:"text,omitempty"`
	ViewState     *ObjectRef  `xml:"view_state,omitempty"`
	DateSubmitted Time        `xml:"date_submitted,omitempty"`
	LastModified  Time        `xml:"last_modified,omitempty"`
	TimeTracking  *int        `xml:"time_tracking,omitempty"`
	NoteType      *int        `xml:"note_type,omitempty"`
	NoteAttr      string      `xml:"note_attr,omitempty"`
}

type IssueData struct {
	ID                    *int               `xml:"id,omitempty"`
	ViewState             *ObjectRef         `xml:"view_state,omitempty"`
	LastUpdated           *Time              `xml:"last_updated,omitempty"`
	Project               *ObjectRef         `xml:"project,omitempty"`
	Category              *string            `xml:"category,omitempty"`
	Priority              *ObjectRef         `xml:"priority,omitempty"`
	Severity              *ObjectRef         `xml:"severity,omitempty"`
	Status                *ObjectRef         `xml:"status,omitempty"`
	Reporter              *AccountData       `xml:"reporter,omitempty"`
	Summary               *string            `xml:"summary,omitempty"`
	Version               *string            `xml:"version,omitempty"`
	Build                 *string            `xml:"build,omitempty"`
	Platform              *string            `xml:"platform,omitempty"`
	Os                    *string            `xml:"os,omitempty"`
	OsBuild               *string            `xml:"os_build,omitempty"`
	Reproducibility       *ObjectRef         `xml:"reproducibility,omitempty"`
	DateSubmitted         *Time              `xml:"date_submitted,omitempty"`
	SponsorshipTotal      *int               `xml:"sponsorship_total,omitempty"`
	Handler               *AccountData       `xml:"handler,omitempty"`
	Projection            *ObjectRef         `xml:"projection,omitempty"`
	ETA                   *ObjectRef         `xml:"eta,omitempty"`
	Resolution            *ObjectRef         `xml:"resolution,omitempty"`
	FixedInVersion        *string            `xml:"fixed_in_version,omitempty"`
	TargetVersion         *string            `xml:"target_version,omitempty"`
	Description           *string            `xml:"description,omitempty"`
	StepsToReproduce      *string            `xml:"steps_to_reproduce,omitempty"`
	AdditionalInformation *string            `xml:"additional_information,omitempty"`
	Attachments           []AttachmentData   `xml:"attachments>item,omitempty"`
	Relationships         []RelationshipData `xml:"relationships>item,omitempty"`
	Notes                 []NoteData         `xml:"notes>item,omitempty"`
	CustomFields          []CustomFieldData  `xml:"custom_fields>item,omitempty"`
	DueDate               *Time              `xml:"due_date,omitempty"`
	Monitors              []AccountData      `xml:"monitors>item,omitempty"`
	Sticky                *bool              `xml:"sticky,omitempty"`
	Tags                  []ObjectRef        `xml:"tags>item,omitempty"`
}

type ObjectRef struct {
	ID   int    `xml:"id,omitempty"`
	Name string `xml:"name,omitempty"`
}

type AttachmentData struct {
	ID            int    `xml:"id,omitempty"`
	FileName      string `xml:"filename,omitempty"`
	Size          int    `xml:"size,omitempty"`
	ContentType   string `xml:"content_type,omitempty"`
	DateSubmitted Time   `xml:"date_submitted,omitempty"`
	DownloadURL   string `xml:"download_url,omitempty"`
	UserID        int    `xml:"user_id,omitempty"`
}

type RelationshipData struct {
	ID       int       `xml:"id,omitempty"`
	Type     ObjectRef `xml:"type,omitempty"`
	TargetID int       `xml:"target_id,omitempty"`
}

type NoteData struct {
	ID            int         `xml:"id,omitempty"`
	Reporter      AccountData `xml:"reporter,omitempty"`
	Text          string      `xml:"text,omitempty"`
	ViewState     *ObjectRef  `xml:"view_state,omitempty"`
	DateSubmitted Time        `xml:"date_submitted,omitempty"`
	LastModified  Time        `xml:"last_modified,omitempty"`
	TimeTracking  int         `xml:"time_tracking,omitempty"`
	NoteType      int         `xml:"note_type,omitempty"`
	NoteAttr      string      `xml:"note_attr,omitempty"`
}

type CustomFieldData struct {
	Field ObjectRef `xml:"field,omitempty"`
	Value string    `xml:"value,omitempty"`
}

type Auth struct {
	Username string `xml:"username"`
	Password string `xml:"password"`
}

type UserData struct {
	Account     AccountData `xml:"account_data,omitempty"`
	AccessLevel int         `xml:"access_level,omitempty"`
	Timezone    string      `xml:"timezone,omitempty"`
}

type AccountData struct {
	//XMLName  xml.Name `xml:"account_data,omitempty"`
	ID       int    `xml:"id,omitempty"`
	Name     string `xml:"name,omitempty"`
	RealName string `xml:"real_name,omitempty"`
	Email    string `xml:"email,omitempty"`
}

type ProjectData struct {
	ID            int           `xml:"id,omitempty"`
	Name          string        `xml:"name,omitempty"`
	Status        *ObjectRef    `xml:"status,omitempty"`
	Enabled       bool          `xml:"enabled,omitempy"`
	ViewState     *ObjectRef    `xml:"view_state,omitempty"`
	AccessMin     *ObjectRef    `xml:"access_min,omitempty"`
	FilePath      string        `xml:"file_path,omitempty"`
	Description   string        `xml:"description,omitempty"`
	Subprojects   []ProjectData `xml:"subprojects>item,omitempty"`
	InheritGlobal bool          `xml:"inherit_global,omitempty"`
}
