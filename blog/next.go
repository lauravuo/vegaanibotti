package blog

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"os"
	"slices"

	"github.com/lainio/err2/try"
)

const WritePerm = 0o600

func ChooseNextPost(posts []Post, usedIDsPath string) Post {
	filePath := usedIDsPath
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		try.To(os.WriteFile(filePath, []byte("[]"), WritePerm))
	}

	fileContents := try.To1(os.ReadFile(filePath))

	var usedIDs []int64

	try.To(json.Unmarshal(fileContents, &usedIDs))
	count := int64(len(posts) - len(usedIDs))

	// all ids are used, reset
	if count == 0 {
		count = int64(len(posts))
		usedIDs = make([]int64, 0)
	}

	randomIndex := int(try.To1(rand.Int(rand.Reader, big.NewInt(count))).Int64())
	slog.Info("Picking random post", "index", randomIndex)

	var chosenPost *Post

	filteredIndex := -1

	for index, post := range posts {
		if !slices.Contains(usedIDs, post.ID) {
			filteredIndex++
		}

		if filteredIndex == randomIndex {
			chosenPost = &posts[index]

			break
		}
	}

	usedIDs = append(usedIDs, chosenPost.ID)

	try.To(os.WriteFile(filePath, try.To1(json.Marshal(usedIDs)), WritePerm))

	return *chosenPost
}
