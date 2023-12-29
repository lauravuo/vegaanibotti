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
	"github.com/lauravuo/vegaanibotti/blog/base"
)

func ChooseNextPost(posts base.Collection) base.Post {
	filePath := posts["cc"].UsedIDsPath
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		slog.Info("Used ids file does not exist, creating...", "path", filePath)
		try.To(os.WriteFile(filePath, []byte("[]"), base.WritePerm))
	}

	fileContents := try.To1(os.ReadFile(filePath))

	var usedIDs []int64

	try.To(json.Unmarshal(fileContents, &usedIDs))
	count := int64(len(posts["cc"].Posts) - len(usedIDs))

	slog.Debug("Unused ids", "count", count)

	// all ids are used, reset
	if count == 0 {
		count = int64(len(posts))
		usedIDs = make([]int64, 0)
	}

	randomIndex := int(try.To1(rand.Int(rand.Reader, big.NewInt(count))).Int64())
	slog.Info("Picking random post", "index", randomIndex)

	var chosenPost *base.Post

	filteredIndex := -1

	for index, post := range posts["cc"].Posts {
		if !slices.Contains(usedIDs, post.ID) {
			filteredIndex++
		}

		if filteredIndex == randomIndex {
			chosenPost = &posts["cc"].Posts[index]

			break
		}
	}

	usedIDs = append(usedIDs, chosenPost.ID)

	try.To(os.WriteFile(filePath, try.To1(json.Marshal(usedIDs)), base.WritePerm))

	return *chosenPost
}
