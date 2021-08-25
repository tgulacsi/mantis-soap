// Copyright 2015 Tamás Gulácsi
//
// SPDX-License-Identifier: Apache-2.0

package mantis

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"time"
)

type Time time.Time

func (t *Time) IsZero() bool {
	return t == nil || time.Time(*t).IsZero()
}

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
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	pr, pw := io.Pipe()
	go func() {
		w := base64.NewEncoder(base64.StdEncoding, pw)
		n, err := io.Copy(w, r.Reader)
		if err != nil {
			err = fmt.Errorf("base64-encode: %w", err)
		}
		Logger.Log("msg", "copied", "bytes", n, "error", err)
		if closeErr := w.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close base64-encoder: %w", closeErr)
		}
		_ = pw.CloseWithError(err)
	}()
	p := make([]byte, 4096)
	var n int
	var err error
	for {
		n, err = pr.Read(p)
		Logger.Log("msg", "read", "bytes", n, "error", err)
		if n > 0 {
			if encErr := e.EncodeToken(xml.CharData(p[:n])); encErr != nil && err == nil {
				err = encErr
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			break
		}
	}
	if closeErr := e.EncodeToken(start.End()); closeErr != nil && err == nil {
		return fmt.Errorf("closing token: %w", closeErr)
	}
	return err
}

// vim: set fileencoding=utf-8 noet:
