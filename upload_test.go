/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package tracker

import (
	"go.uber.org/zap/zaptest"
	"os"
	"testing"
	"time"
)

var (
	ep             = os.Getenv("TRACKER_URL")
	db             = os.Getenv("TRACKER_DB")
	pol            = os.Getenv("TRACKER_RP")
	tok            = os.Getenv("TRACKER_TOKEN")
	sessionId      = "testSession1"
	launchTime     = 1716994039000
	launchDuration = 320010
	appId          = "InDesign1"
	appVersion     = "19.2"
	appLocale      = "en_US"
	nglVersion     = "1.35.0.19"
	osName         = "MAC"
	osVersion      = "14.3.1"
	userId         = "9e5fa"
)

func TestSessionLineDurationOnly(t *testing.T) {
	expected := `launch,sessionId=testSession1 launchDuration=320010 1716994039000`

	s := logSession{
		sessionId:      sessionId,
		launchTime:     time.UnixMilli(int64(launchTime)),
		launchDuration: time.Duration(launchDuration * 1000000),
	}
	l := sessionLine(s)
	if l != expected {
		t.Errorf("sessionLine(%v): expected %q,\ngot %q", sessionId, expected, l)
	}
}

func TestSessionLineAllFields(t *testing.T) {
	expected := `launch,sessionId=testSession1 launchDuration=320010` +
		`,appId="InDesign1",appVersion="19.2"` +
		`,appLocale="en_US"` +
		`,nglVersion="1.35.0.19"` +
		`,osName="MAC",osVersion="14.3.1"` +
		`,userId="9e5fa"` +
		` 1716994039000`

	s := logSession{
		sessionId:      sessionId,
		launchTime:     time.UnixMilli(int64(launchTime)),
		launchDuration: time.Duration(launchDuration * 1000000),
		appId:          appId,
		appVersion:     appVersion,
		appLocale:      appLocale,
		nglVersion:     nglVersion,
		osName:         osName,
		osVersion:      osVersion,
		userId:         userId,
	}
	l := sessionLine(s)
	if l != expected {
		t.Errorf("sessionLine(%v): expected %q,\ngot %q", sessionId, expected, l)
	}
}

func TestUploadSessionLine1(t *testing.T) {
	line1 := `launch,sessionId=testSession1 launchDuration=320010` +
		`,appId="InDesign1",appVersion="19.2"` +
		`,appLocale="en_US"` +
		`,nglVersion="1.35.0.19"` +
		`,osName="MAC",osVersion="14.3.1"` +
		`,userId="9e5fa"` +
		` 1716994039000`
	lines := []string{line1}
	logger := zaptest.NewLogger(t)
	if !uploadLines(ep, db, pol, tok, lines, logger) {
		t.Errorf("uploadLines(%v): expected true, got false", lines)
	}
}

func TestUploadSessionLine2(t *testing.T) {
	line2 := `launch,sessionId=testSession1 launchDuration=640020 1716994039000`
	lines := []string{line2}
	logger := zaptest.NewLogger(t)
	if !uploadLines(ep, db, pol, tok, lines, logger) {
		t.Errorf("uploadLines(%v): expected true, got false", lines)
	}
}

func TestUploadSessionBothLines(t *testing.T) {
	line1 := `launch,sessionId=testSession1 launchDuration=320010` +
		`,appId="InDesign1",appVersion="19.2"` +
		`,appLocale="en_US"` +
		`,nglVersion="1.35.0.19"` +
		`,osName="MAC",osVersion="14.3.1"` +
		`,userId="9e5fa"` +
		` 1716994039000`
	line2 := `launch,sessionId=testSession1 launchDuration=640020 1716994039000`
	lines := []string{line1, line2}
	logger := zaptest.NewLogger(t)
	if !uploadLines(ep, db, pol, tok, lines, logger) {
		t.Errorf("uploadLines(%v): expected true, got false", lines)
	}
}
