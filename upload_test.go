/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package tracker

import (
	"fmt"
	"go.uber.org/zap/zaptest"
	"os"
	"path/filepath"
	"strings"
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
	logger := zaptest.NewLogger(t)
	expected := `log-session,sessionId=testSession1 launchDuration=320010,clientIp="127.0.0.1:53450" 1716994039000`

	s := logSession{
		sessionId:      sessionId,
		launchTime:     time.UnixMilli(int64(launchTime)),
		launchDuration: time.Duration(launchDuration * 1000000),
		clientIp:       "127.0.0.1:53450",
	}
	l := sessionLine(s, logger)
	if l != expected {
		t.Errorf("sessionLine(%v): expected %q,\ngot %q", sessionId, expected, l)
	}
}

func TestSessionLineAllFields(t *testing.T) {
	logger := zaptest.NewLogger(t)
	expected := `log-session,sessionId=testSession1 launchDuration=320010,clientIp="127.0.0.1:53450"` +
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
		clientIp:       "127.0.0.1:53450",
		appId:          appId,
		appVersion:     appVersion,
		appLocale:      appLocale,
		nglVersion:     nglVersion,
		osName:         osName,
		osVersion:      osVersion,
		userId:         userId,
	}
	l := sessionLine(s, logger)
	if l != expected {
		t.Errorf("sessionLine(%v): expected %q,\ngot %q", sessionId, expected, l)
	}
}

func TestSessionLineLatestLogs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	files, err := filepath.Glob("testdata/*.log")
	if err != nil {
		t.Fatalf("Cannot glob testdata/*.log: %s", err)
	}
	for _, file := range files {
		buffer, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Cannot read file %s: %s", file, err)
		}
		sessions := parseLog(string(buffer), "127.0.0.1:53450")
		for _, session := range sessions {
			l := sessionLine(session, logger)
			if !strings.Contains(l, ",appId=") || !strings.Contains(l, ",osName") {
				_ = fmt.Errorf("missing fields in line protocol %q for file %s", l, file)
			}
		}
	}
}

func TestUploadSessionLine1(t *testing.T) {
	logger := zaptest.NewLogger(t)
	line1 := `log-session,sessionId=testSession1 launchDuration=320010,clientIp="127.0.0.1:53450"` +
		`,appId="InDesign1",appVersion="19.2"` +
		`,appLocale="en_US"` +
		`,nglVersion="1.35.0.19"` +
		`,osName="MAC",osVersion="14.3.1"` +
		`,userId="9e5fa"` +
		` 1716994039000`
	lines := []string{line1}
	if err := uploadLines(ep, db, pol, tok, lines, logger); err != nil {
		t.Errorf("uploadLines failed: %s", err.Error())
	}
}

func TestUploadSessionLine2(t *testing.T) {
	logger := zaptest.NewLogger(t)
	line2 := `log-session,sessionId=testSession1 launchDuration=640020,clientIp="127.0.0.1:53450" 1716994039000`
	lines := []string{line2}
	if err := uploadLines(ep, db, pol, tok, lines, logger); err != nil {
		t.Errorf("uploadLines failed: %s", err.Error())
	}
}

func TestUploadSessionBothLines(t *testing.T) {
	logger := zaptest.NewLogger(t)
	line1 := `log-session,sessionId=testSession1 launchDuration=320010,clientIp="127.0.0.1:53450"` +
		`,appId="InDesign1",appVersion="19.2"` +
		`,appLocale="en_US"` +
		`,nglVersion="1.35.0.19"` +
		`,osName="MAC",osVersion="14.3.1"` +
		`,userId="9e5fa"` +
		` 1716994039000`
	line2 := `log-session,sessionId=testSession1 launchDuration=640020,clientIp="127.0.0.1:53450" 1716994039000`
	lines := []string{line1, line2}
	if err := uploadLines(ep, db, pol, tok, lines, logger); err != nil {
		t.Errorf("uploadLines failed: %s", err.Error())
	}
}

func TestUploadLatestLogs(t *testing.T) {
	files, err := filepath.Glob("testdata/*.log")
	if err != nil {
		t.Fatalf("Cannot glob testdata/*.log: %s", err)
	}
	for _, file := range files {
		buffer, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Cannot read file %s: %s", file, err)
		}
		sessions := parseLog(string(buffer), "127.0.0.1:53450")
		logger := zaptest.NewLogger(t)
		if err = sendSessions(ep, db, pol, tok, sessions, logger); err != nil {
			t.Errorf("Failed to send sessions from: %s", file)
		}
	}
}
