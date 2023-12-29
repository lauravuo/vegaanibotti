package base

import "strings"

const WritePerm = 0o600

const lineFeed = "\n\n"

type Post struct {
	ID          int64
	Title       string
	Description string
	URL         string
	Hashtags    []string
	Added       bool `json:"-"`
}

func (p *Post) IsValid() bool {
	return p.ID != 0 && p.Title != "" && p.Description != "" && p.URL != ""
}

func (p *Post) Summary() string {
	return p.Title + lineFeed +
		p.Description + lineFeed +
		p.URL + lineFeed +
		"#" + strings.Join(p.Hashtags, " #")
}

func (p *Post) ShortSummary() string {
	const shortSummaryMaxLen = 280

	summary := p.Summary()
	if len(summary) < shortSummaryMaxLen {
		return summary
	}

	return p.Title + lineFeed +
		p.URL + lineFeed +
		"#" + strings.Join(p.Hashtags, " #")
}
