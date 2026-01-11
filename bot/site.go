package bot

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
)

type Site struct{}

func InitSite() *Site {
	return &Site{}
}

// escapeYAMLString escapes double quotes and backslashes in a string for use in YAML.
func escapeYAMLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")

	return s
}

func (s *Site) PostToSite(post *base.Post) error {
	const dirPermission = 0o700

	now := time.Now()
	year := fmt.Sprintf("%d", now.Year())
	month := fmt.Sprintf("%02d", now.Month())
	date := fmt.Sprintf("%s-%s-%02d", year, month, now.Day())
	folder := "./site/content/" + year + "/" + month
	path := folder + "/" + date + ".md"

	content := fmt.Sprintf(`---
title: "%s"
image: "%s"
date: %s
receipt_url: "%s"
author: "%s"
---`,
		escapeYAMLString(post.Title),
		escapeYAMLString(post.ThumbnailURL),
		date,
		escapeYAMLString(post.URL),
		escapeYAMLString(post.Author),
	)

	try.To(os.MkdirAll(folder, dirPermission))
	try.To(os.WriteFile(path, []byte(content), base.WritePerm))

	slog.Info("post written to file", "path", path)

	return nil
}
