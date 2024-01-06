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
	now := time.Now()
	date := fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())
	path := "./site/content/" + date + ".md"

	content := fmt.Sprintf(`---
title: "%s"
image: "%s"
date: %s
receipt_url: "%s"
---`, post.Title, post.ThumbnailURL, date, post.URL,
	)

	try.To(os.WriteFile(path, []byte(content), base.WritePerm))

	slog.Info("post written to file", "path", path)

	return nil
}
