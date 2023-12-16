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

const writePerm = 0o600

func ChooseNextPost(posts []Post) Post {
	filePath := UsedIDsPath
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		try.To(os.WriteFile(filePath, []byte("[]"), writePerm))
	}

	fileContents := try.To1(os.ReadFile(filePath))

	var usedIDs []int64

	try.To(json.Unmarshal(fileContents, &usedIDs))
	count := int64(len(posts) - len(usedIDs))

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
	// if all ids are used, reset array
	if len(usedIDs) == len(posts) {
		usedIDs = []int64{}
	}

	try.To(os.WriteFile(filePath, try.To1(json.Marshal(usedIDs)), writePerm))

	return *chosenPost
}
