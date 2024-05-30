/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package tracker

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// sendSessions takes an InfluxDB upload URL and a sequence of logSessions
// and uploads the logSession data to InfluxDB.
func sendSessions(ep string, db string, pol string, tok string, sessions []logSession) {
	if len(sessions) > 0 {
		sendSessionsInternal(ep, db, pol, tok, sessions, nil)
	}
}

func sendSessionsInternal(ep string, db string, pol string, tok string, sessions []logSession, logger *zap.Logger) bool {
	var lines = make([]string, 0, len(sessions))
	for _, session := range sessions {
		lines = append(lines, sessionLine(session))
	}
	return uploadLines(ep, db, pol, tok, lines, logger)
}

// sessionLine constructs a line protocol line for the given logSession
func sessionLine(s logSession) string {
	line := fmt.Sprintf("log-session,sessionId=%s launchDuration=%d", s.sessionId, s.launchDuration.Milliseconds())
	if s.appId != "" {
		line = line + fmt.Sprintf(",appId=%q,appVersion=%q", s.appId, s.appVersion)
	}
	if s.appLocale != "" {
		line = line + fmt.Sprintf(",appLocale=%q", s.appLocale)
	}
	if s.nglVersion != "" {
		line = line + fmt.Sprintf(",nglVersion=%q", s.nglVersion)
	}
	if s.osName != "" {
		line = line + fmt.Sprintf(",osName=%q,osVersion=%q", s.osName, s.osVersion)
	}
	if s.userId != "" {
		line = line + fmt.Sprintf(",userId=%q", s.userId)
	}
	line = line + fmt.Sprintf(" %d", s.launchTime.UnixMilli())
	return line
}

func uploadLines(ep string, db string, pol string, tok string, lines []string, logger *zap.Logger) bool {
	if logger == nil {
		logger = caddy.Log()
	}
	target := fmt.Sprintf("%s/write?db=%s&rp=%s&precision=ms", ep, url.QueryEscape(db), url.QueryEscape(pol))
	body := strings.NewReader(strings.Join(lines, "\n") + "\n")
	req, err := http.NewRequest("POST", target, body)
	if err != nil {
		caddy.Log().Error("influx upload create request error", zap.String("error", err.Error()))
		return false
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", tok))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("influx upload POST request error", zap.String("error", err.Error()))
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("influx upload POST response body close error", zap.String("error", err.Error()))
		}
	}(res.Body)
	if res.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Error("influx upload read response error", zap.String("error", err.Error()))
		}
		logger.Error("influx upload data issues", zap.Int("status", res.StatusCode), zap.String("error", string(body)))
		return false
	}
	return true
}
