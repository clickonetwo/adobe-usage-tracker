/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

// Package tracker provides the caddy adobe_usage_tracker plugin.
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
func sendSessions(ep string, db string, pol string, tok string, sessions []logSession, logger *zap.Logger) error {
	if len(sessions) == 0 {
		return nil
	}
	var lines = make([]string, 0, len(sessions))
	for _, session := range sessions {
		lines = append(lines, sessionLine(session, logger))
	}
	return uploadLines(ep, db, pol, tok, lines, logger)
}

// sessionLine constructs a line protocol line for the given logSession
func sessionLine(s logSession, logger *zap.Logger) string {
	line := fmt.Sprintf("log-session,sessionId=%s launchDuration=%d,clientIp=%q",
		s.sessionId,
		s.launchDuration.Milliseconds(),
		s.clientIp,
	)
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
	logger.Debug("session-line-protocol", zap.Object("session", s), zap.String("line", line))
	return line
}

func uploadLines(ep string, db string, pol string, tok string, lines []string, logger *zap.Logger) error {
	content := strings.Join(lines, "\n") + "\n"
	logger.Debug("AdobeUsageTracker uploading line protocol",
		zap.Strings("incoming", lines), zap.String("outgoing", content))
	target := fmt.Sprintf("%s/write?db=%s&rp=%s&precision=ms", ep, url.QueryEscape(db), url.QueryEscape(pol))
	body := strings.NewReader(content)
	req, err := http.NewRequest("POST", target, body)
	if err != nil {
		caddy.Log().Error("AdobeUsageTracker upload create request error", zap.String("error", err.Error()))
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", tok))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("AdobeUsageTracker upload POST request error", zap.String("error", err.Error()))
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("AdobeUsageTracker POST response close error", zap.String("error", err.Error()))
		}
	}(res.Body)
	if res.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Error("AdobeUsageTracker upload error response invalid",
				zap.Int("status", res.StatusCode),
				zap.String("error", err.Error()),
			)
		} else {
			logger.Error("AdobeUsageTracker upload data issues",
				zap.Int("status", res.StatusCode),
				zap.String("error", string(body)),
			)
		}
		return fmt.Errorf("upload status code: %d", res.StatusCode)
	}
	return nil
}
