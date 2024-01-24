package bot

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
)

type Site struct{}

func InitSite() *Site {
	return &Site{}
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
image: "./vegaanibotti.png"
date: %s
receipt_url: "%s"
---`, post.Title, date, post.URL,
	)

	try.To(os.MkdirAll(folder, dirPermission))
	try.To(os.WriteFile(path, []byte(content), base.WritePerm))

	slog.Info("post written to file", "path", path)

	return nil
}
