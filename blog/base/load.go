package base

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/lainio/err2/try"
)

func LoadExistingPosts(recipesFilePath string) (posts []Post, maxID int64) {
	// create file if it does not exist
	if _, err := os.Stat(recipesFilePath); errors.Is(err, os.ErrNotExist) {
		try.To(os.WriteFile(recipesFilePath, []byte("[]"), WritePerm))
	}

	// read existing posts
	fileContents := try.To1(os.ReadFile(recipesFilePath))
	try.To(json.Unmarshal(fileContents, &posts))

	if len(posts) > 0 {
		maxID = posts[0].ID
	}

	slog.Info("Existing posts", "count", len(posts), "maximum ID", maxID)

	return posts, maxID
}
