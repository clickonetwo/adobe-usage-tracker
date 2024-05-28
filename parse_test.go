/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package tracker

import (
	"fmt"
	"os"
	"testing"
)

func TestParseSingleSessionLogs(t *testing.T) {
	for i := 1; i <= 2; i++ {
		path := fmt.Sprintf("testdata/indesign-single-session-%d.txt", i)
		buffer, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file %s: %s", path, err)
		}
		sessions := parseLog(string(buffer))
		if len(sessions) != 1 {
			t.Fatalf("Expected 1 session, got %d", len(sessions))
		}
		session := sessions[0]
		if session.appId != "InDesign1" {
			t.Fatalf("%d: Expected appId %q, got %q", i, "InDesign1", sessions[0].appId)
		}
		if session.osName != "MAC" {
			t.Fatalf("%d: Expected osName %q, got %q", i, "MAC", session.osName)
		}
		if session.nglVersion != "1.35.0.19" {
			t.Fatalf("%d: Expected nglVersion %q, got %q", i, "1.35.0.19", session.nglVersion)
		}
	}
}

func TestParseSplitSessionLogs(t *testing.T) {
	var buffer []byte
	var err error
	var sessions []logSession
	path1 := fmt.Sprintf("testdata/indesign-split-session-1-1.txt")
	path2 := fmt.Sprintf("testdata/indesign-split-session-1-2.txt")
	buffer, err = os.ReadFile(path1)
	if err != nil {
		t.Fatalf("Failed to read file %s: %s", path1, err)
	}
	sessions = parseLog(string(buffer))
	if len(sessions) != 1 {
		t.Fatalf("%s: Expected 1 session, got %d", path1, len(sessions))
	}
	session1 := sessions[0]
	buffer, err = os.ReadFile(path2)
	if err != nil {
		t.Fatalf("Failed to read file %s: %s", path2, err)
	}
	sessions = parseLog(string(buffer))
	if len(sessions) != 1 {
		t.Fatalf("%s: Expected 1 session, got %d", path2, len(sessions))
	}
	session2 := sessions[0]
	if session1.sessionId != session2.sessionId {
		t.Fatalf("Session ids differ in split-session logs")
	}
	if session1.launchDuration >= session2.launchDuration {
		t.Fatalf(
			"Session 2 launch duration (%v) < Session 1 launch duration (%v)",
			session2.launchDuration, session1.launchDuration,
		)
	}
}

func TestParseMultiSessionLogs(t *testing.T) {
	var buffer []byte
	var err error
	var sessions []logSession
	path1 := fmt.Sprintf("testdata/indesign-multi-session-1-1.txt")
	path2 := fmt.Sprintf("testdata/indesign-multi-session-1-2.txt")
	buffer, err = os.ReadFile(path1)
	if err != nil {
		t.Fatalf("Failed to read file %s: %s", path1, err)
	}
	sessions = parseLog(string(buffer))
	if len(sessions) != 1 {
		t.Fatalf("%s: Expected 1 session, got %d", path1, len(sessions))
	}
	session1 := sessions[0]
	buffer, err = os.ReadFile(path2)
	if err != nil {
		t.Fatalf("Failed to read file %s: %s", path2, err)
	}
	sessions = parseLog(string(buffer))
	if len(sessions) != 2 {
		t.Fatalf("%s: Expected 1 session, got %d", path2, len(sessions))
	}
	session2 := sessions[0]
	session3 := sessions[1]
	if session1.sessionId != session2.sessionId {
		t.Fatalf("Session ids differ in split-multi-session logs")
	}
	if session2.sessionId == session3.sessionId {
		t.Fatalf("Session ids don't differ in multi-session logs")
	}
	if session1.launchTime.Compare(session3.launchTime) != -1 {
		t.Fatalf(
			"Session 1 launch time (%v) < Session 3 launch time (%v)",
			session1.launchTime, session3.launchTime,
		)
	}
}
