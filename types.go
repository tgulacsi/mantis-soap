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
	"encoding/base64"
	"encoding/xml"
	"io"
	"log"
	"time"
)

type Time time.Time

const timePattern = time.RFC3339

func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p := make([]byte, 0, 32)
Loop:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch x := tok.(type) {
		case xml.ProcInst, xml.Comment, xml.Directive:
			continue Loop
		case xml.CharData:
			p = append(p, x...)
		default:
			break Loop
		}
	}

	p = bytes.TrimSpace(p)
	if len(p) == 0 {
		return nil
	}
	// CCYY-MM-DDThh:mm:ss[Z|(+|-)hh:mm]
	n := len(p)
	if n > len(timePattern) {
		n = len(timePattern)
	}
	t2, err := time.Parse(timePattern[:n], string(p[:n]))
	*t = Time(t2)
	return err
}

type Reader struct {
	io.Reader
}

func (r Reader) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	var buf bytes.Buffer
	//pr, pw := io.Pipe()
	//go func() {
	//w := base64.NewEncoder(base64.StdEncoding, pw)
	//n, err := io.Copy(w, r.Reader)
	w := base64.NewEncoder(base64.StdEncoding, &buf)
	n, err := io.Copy(w, r.Reader)
	log.Printf("copied %d (%v)", n, err)
	if closeErr := w.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	//pw.CloseWithError(err)
	//}()
	//p := make([]byte, 4096)
	//for {
	//n, err := pr.Read(p)
	//log.Printf("read %d (%v)", n, err)
	p := buf.Bytes()
	if n > 0 {
		e.EncodeToken(xml.CharData(p[:n]))
	}
	//if err != nil {
	//break
	//}
	//}
	if closeErr := e.EncodeToken(start.End()); closeErr != nil && err == nil {
		return closeErr
	}
	return err
}
