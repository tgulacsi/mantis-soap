// Copyright 2013 Tamás Gulácsi
//
// SPDX-License-Identifier: Apache-2.0

package mantis

import (
	"encoding/xml"
	"strings"
	"testing"
)

const issueGetResponse = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="http://futureware.biz/mantisconnect" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <SOAP-ENV:Body>
    <ns1:mc_issue_getResponse>
      <return xsi:type="ns1:IssueData">
        <id xsi:type="xsd:integer">1000</id>
        <view_state xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">10</id>
          <name xsi:type="xsd:string">publikus</name>
        </view_state>
        <last_updated xsi:type="xsd:dateTime">2013-08-16T17:39:49+01:00</last_updated>
        <project xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">1</id>
          <name xsi:type="xsd:string">COMPANY</name>
        </project>
        <category xsi:type="xsd:string"/>
        <priority xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">30</id>
          <name xsi:type="xsd:string">normál</name>
        </priority>
        <severity xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">50</id>
          <name xsi:type="xsd:string">apró hiba</name>
        </severity>
        <status xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">402</id>
          <name xsi:type="xsd:string">ajánlat elfogadva</name>
        </status>
        <reporter xsi:type="ns1:AccountData">
          <id xsi:type="xsd:integer">48</id>
          <name xsi:type="xsd:string">r</name>
          <real_name xsi:type="xsd:string">R</real_name>
          <email xsi:type="xsd:string">r@company.com</email>
        </reporter>
        <summary xsi:type="xsd:string">REQ999999999 - XXX kezelése propagációnál - PÓTmegrendelő</summary>
        <reproducibility xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">70</id>
          <name xsi:type="xsd:string">nem próbáltam</name>
        </reproducibility>
        <date_submitted xsi:type="xsd:dateTime">2013-07-11T17:15:48+01:00</date_submitted>
        <sponsorship_total xsi:type="xsd:integer">0</sponsorship_total>
        <handler xsi:type="ns1:AccountData">
          <id xsi:type="xsd:integer">3</id>
          <name xsi:type="xsd:string">T</name>
          <real_name xsi:type="xsd:string">D</real_name>
          <email xsi:type="xsd:string">T@devco.com</email>
        </handler>
        <projection xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">10</id>
          <name xsi:type="xsd:string">nincs</name>
        </projection>
        <eta xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">10</id>
          <name xsi:type="xsd:string">semmi</name>
        </eta>
        <resolution xsi:type="ns1:ObjectRef">
          <id xsi:type="xsd:integer">10</id>
          <name xsi:type="xsd:string">nyitott</name>
        </resolution>
        <description xsi:type="xsd:string">Ha az ügyfél kéri a változást, akkor propagálódhat, csak ne cserélje le. Álljon rendelkezésre és ki tudja választani az ügyintéző az A-ban. </description>
        <additional_information xsi:type="xsd:string">A portfoliók az S törlésére vonatkozó propagálást tárolják el. Vizsgálják meg,hogy az ügyfél mely szerződésén van az adott és a törölt S bejegyezve.&#13;
&#13;
Ha a vizsgált szerződés Z, a portfolió a következő gyakorisági időszakban állítsa át a szerződést, akkor, ha nincs rá előjegyzésben váltás. Az ügyfelet ezt követően a lokális levél előjegyzésével és nyomdai nyomtatásával és küldésével tájékoztatjuk. &#13;
&#13;
Ha szerződés E, akkor a portfolióknak nincs teendője.&#13;
&#13;
Ha a szerződés C, szintén nincs teendője.&#13;
</additional_information>
        <attachments SOAP-ENC:arrayType="ns1:AttachmentData[2]" xsi:type="SOAP-ENC:Array">
          <item xsi:type="ns1:AttachmentData">
            <id xsi:type="xsd:integer">1601</id>
            <filename xsi:type="xsd:string">REQ999999999_XXX_változas_PÓTmegnedelő.docx</filename>
            <size xsi:type="xsd:integer">76747</size>
            <content_type xsi:type="xsd:string">application/vnd.openxmlformats-officedocument.wordprocessingml.document</content_type>
            <date_submitted xsi:type="xsd:dateTime">2013-07-11T17:15:48+01:00</date_submitted>
            <download_url xsi:type="xsd:anyURI">https://www.devco.com/mantis/company/file_download.php?file_id=1601&amp;amp;type=bug</download_url>
            <user_id xsi:type="xsd:integer">48</user_id>
          </item>
          <item xsi:type="ns1:AttachmentData">
            <id xsi:type="xsd:integer">1779</id>
            <filename xsi:type="xsd:string">REQ999999999_változas_PÓTmegnedelő_20131216.docx</filename>
            <size xsi:type="xsd:integer">77544</size>
            <content_type xsi:type="xsd:string">application/vnd.openxmlformats-officedocument.wordprocessingml.document</content_type>
            <date_submitted xsi:type="xsd:dateTime">2013-08-16T15:02:46+01:00</date_submitted>
            <download_url xsi:type="xsd:anyURI">https://www.devco.com/mantis/company/file_download.php?file_id=1779&amp;amp;type=bug</download_url>
            <user_id xsi:type="xsd:integer">48</user_id>
          </item>
        </attachments>
        <notes SOAP-ENC:arrayType="ns1:IssueNoteData[10]" xsi:type="SOAP-ENC:Array">
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">7515</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">48</id>
              <name xsi:type="xsd:string">r</name>
              <real_name xsi:type="xsd:string">R</real_name>
              <email xsi:type="xsd:string">r@company.com</email>
            </reporter>
            <text xsi:type="xsd:string">Kedves fejlesztők!&#13;
Kérem a csatolt megrendelőre adjatok ajánlatot. Kívánt határidő: 2013.11.20&#13;
Köszönöm,&#13;
R&#13;
&#13;
ui. ÉN VAGYOK 1000-es MANTISJEGY boldog tulajdonosa!</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-11T17:17:50+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-11T17:17:50+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">7538</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">3</id>
              <name xsi:type="xsd:string">T</name>
              <real_name xsi:type="xsd:string">D</real_name>
              <email xsi:type="xsd:string">T@devco.com</email>
            </reporter>
            <text xsi:type="xsd:string">Ráfordítás: 0.12345678 embernap, a sürgősség és bonyolultság miatt.&#13;
Kérek rá összeget!&#13;
T

</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">50</id>
              <name xsi:type="xsd:string">privát</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-12T12:14:11+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-12T12:38:31+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">7569</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">16</id>
              <name xsi:type="xsd:string">v</name>
              <real_name xsi:type="xsd:string">V</real_name>
              <email xsi:type="xsd:string">V@devco.com</email>
            </reporter>
            <text xsi:type="xsd:string">623 792,25</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">50</id>
              <name xsi:type="xsd:string">privát</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-13T09:01:30+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-13T09:01:30+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">7570</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">3</id>
              <name xsi:type="xsd:string">T</name>
              <real_name xsi:type="xsd:string">D</real_name>
              <email xsi:type="xsd:string">T@devco.com</email>
            </reporter>
            <text xsi:type="xsd:string">Ráfordítás: 0.12345678 embernap, azaz 123 456,78 Ft - a sürgősség és bonyolultság miatt.&#13;
T</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-13T09:20:09+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-13T09:20:09+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">7924</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">48</id>
              <name xsi:type="xsd:string">r</name>
              <real_name xsi:type="xsd:string">R</real_name>
              <email xsi:type="xsd:string">r@company.com</email>
            </reporter>
            <text xsi:type="xsd:string">Az ajánlat elfogadva, kérem a fejlesztés megkezdését.&#13;
Köszönöm, R</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-25T14:36:38+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-25T14:36:38+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">7987</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">3</id>
              <name xsi:type="xsd:string">T</name>
              <real_name xsi:type="xsd:string">D</real_name>
              <email xsi:type="xsd:string">T@devco.com</email>
            </reporter>
            <text xsi:type="xsd:string">Kedves R!&#13;
&#13;
B oldalon mi legyen?&#13;
Amit lehet:&#13;
a) ugyanaz mint A oldalon&#13;
b) csak töröljük és tegyük egyedi átutalásosba (ha csoportos)?&#13;
&#13;
T</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-26T14:55:08+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-26T14:55:08+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">7989</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">48</id>
              <name xsi:type="xsd:string">r</name>
              <real_name xsi:type="xsd:string">R</real_name>
              <email xsi:type="xsd:string">r@company.com</email>
            </reporter>
            <text xsi:type="xsd:string">utánánézekR</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-26T15:33:46+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-26T15:33:46+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">8031</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">48</id>
              <name xsi:type="xsd:string">r</name>
              <real_name xsi:type="xsd:string">R</real_name>
              <email xsi:type="xsd:string">r@company.com</email>
            </reporter>
            <text xsi:type="xsd:string">A kérdéseidre hétfőn lesz válasz.&#13;
üdv, R</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-07-27T15:16:21+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-07-27T15:16:21+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">8538</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">48</id>
              <name xsi:type="xsd:string">r</name>
              <real_name xsi:type="xsd:string">R</real_name>
              <email xsi:type="xsd:string">r@company.com</email>
            </reporter>
            <text xsi:type="xsd:string">Kiegészítettem a megrendelőt B oldalról:&#13;
&#13;
Ha S törlés érkezik az ügyfélre, akkor Q, majd a központi vagy a manuális menesztéskor is alkalmazott ellenőrzési logika szerint rakja rá a hibakódot az ajánlatra, mivel az ajánlaton megadott díjkezelési rendszernek  ellentmond az ajánlati adattartalom</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-08-16T15:02:27+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-08-16T15:02:27+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
          <item xsi:type="ns1:IssueNoteData">
            <id xsi:type="xsd:integer">8547</id>
            <reporter xsi:type="ns1:AccountData">
              <id xsi:type="xsd:integer">3</id>
              <name xsi:type="xsd:string">T</name>
              <real_name xsi:type="xsd:string">D</real_name>
              <email xsi:type="xsd:string">T@devco.com</email>
            </reporter>
            <text xsi:type="xsd:string">Rendben, beletettem.&#13;
T</text>
            <view_state xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">10</id>
              <name xsi:type="xsd:string">publikus</name>
            </view_state>
            <date_submitted xsi:type="xsd:dateTime">2013-08-16T17:39:49+01:00</date_submitted>
            <last_modified xsi:type="xsd:dateTime">2013-08-16T17:39:49+01:00</last_modified>
            <time_tracking xsi:type="xsd:integer">0</time_tracking>
            <note_type xsi:type="xsd:integer">0</note_type>
            <note_attr xsi:type="xsd:string"/>
          </item>
        </notes>
        <custom_fields SOAP-ENC:arrayType="ns1:CustomFieldValueForIssueData[7]" xsi:type="SOAP-ENC:Array">
          <item xsi:type="ns1:CustomFieldValueForIssueData">
            <field xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">4</id>
              <name xsi:type="xsd:string">Nyilv.szám</name>
            </field>
            <value xsi:type="xsd:string">REQ999999999</value>
          </item>
          <item xsi:type="ns1:CustomFieldValueForIssueData">
            <field xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">2</id>
              <name xsi:type="xsd:string">Határidő</name>
            </field>
            <value xsi:type="xsd:string">1449183600</value>
          </item>
          <item xsi:type="ns1:CustomFieldValueForIssueData">
            <field xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">3</id>
              <name xsi:type="xsd:string">Típus</name>
            </field>
            <value xsi:type="xsd:string">Fejlesztés</value>
          </item>
          <item xsi:type="ns1:CustomFieldValueForIssueData">
            <field xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">1</id>
              <name xsi:type="xsd:string">Szervezés</name>
            </field>
            <value xsi:type="xsd:string">.3</value>
          </item>
          <item xsi:type="ns1:CustomFieldValueForIssueData">
            <field xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">5</id>
              <name xsi:type="xsd:string">Programozás</name>
            </field>
            <value xsi:type="xsd:string">1</value>
          </item>
          <item xsi:type="ns1:CustomFieldValueForIssueData">
            <field xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">6</id>
              <name xsi:type="xsd:string">Tesztelés</name>
            </field>
            <value xsi:type="xsd:string">.7</value>
          </item>
          <item xsi:type="ns1:CustomFieldValueForIssueData">
            <field xsi:type="ns1:ObjectRef">
              <id xsi:type="xsd:integer">7</id>
              <name xsi:type="xsd:string">Végrehajtás</name>
            </field>
            <value xsi:type="xsd:string">.5</value>
          </item>
        </custom_fields>
        <due_date xsi:nil="true" xsi:type="xsd:dateTime"/>
        <monitors SOAP-ENC:arrayType="ns1:AccountData[8]" xsi:type="SOAP-ENC:Array">
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">94</id>
            <name xsi:type="xsd:string">Gy</name>
            <real_name xsi:type="xsd:string">Gyö</real_name>
            <email xsi:type="xsd:string">gy@company.com</email>
          </item>
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">54</id>
            <name xsi:type="xsd:string">A</name>
            <real_name xsi:type="xsd:string">A</real_name>
            <email xsi:type="xsd:string">A@company.com</email>
          </item>
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">159</id>
            <name xsi:type="xsd:string">B</name>
            <real_name xsi:type="xsd:string">B</real_name>
            <email xsi:type="xsd:string">B@company.com</email>
          </item>
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">3</id>
            <name xsi:type="xsd:string">T</name>
            <real_name xsi:type="xsd:string">D</real_name>
            <email xsi:type="xsd:string">T@devco.com</email>
          </item>
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">168</id>
            <name xsi:type="xsd:string">J</name>
            <real_name xsi:type="xsd:string">J</real_name>
            <email xsi:type="xsd:string">J@company.com</email>
          </item>
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">92</id>
            <name xsi:type="xsd:string">K</name>
            <real_name xsi:type="xsd:string">Kü</real_name>
            <email xsi:type="xsd:string">K@company.com</email>
          </item>
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">126</id>
            <name xsi:type="xsd:string">Z</name>
            <real_name xsi:type="xsd:string">Z</real_name>
            <email xsi:type="xsd:string">Z@company.com</email>
          </item>
          <item xsi:type="ns1:AccountData">
            <id xsi:type="xsd:integer">16</id>
            <name xsi:type="xsd:string">V</name>
            <real_name xsi:type="xsd:string">Vé</real_name>
            <email xsi:type="xsd:string">V@devco.com</email>
          </item>
        </monitors>
        <sticky xsi:type="xsd:boolean">false</sticky>
        <tags SOAP-ENC:arrayType="ns1:ObjectRef[0]" xsi:type="SOAP-ENC:Array"/>
      </return>
    </ns1:mc_issue_getResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

func TestIssueGet(t *testing.T) {
	d := xml.NewDecoder(strings.NewReader(issueGetResponse))
	for {
		tok, _ := d.Token()
		if tok == nil {
			break
		}
		if x, ok := tok.(xml.StartElement); ok && x.Name.Local == "Body" {
			break
		}
	}

	var response IssueGetResponse
	if err := d.Decode(&response); err != nil {
		t.Error(err)
	}
	data := response.Return
	if data.LastUpdated.IsZero() {
		t.Errorf("bad LastUpdated date: got %v", data.LastUpdated)
	}
	if len(data.Attachments) == 0 || data.Attachments[0].FileName == "" {
		t.Errorf("no attachment! Got\n%#v\nwanted\n%s", data.Attachments, issueGetResponse)
	}
	t.Logf("response=%#v", response)
}
