package blog

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/lainio/err2/try"
)

func ChooseNextPost(posts []Post) Post {
	filePath := USED_IDS_PATH
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		try.To(os.WriteFile(filePath, []byte("[]"), 0644))
	}
	fileContents := try.To1(os.ReadFile(filePath))
	var usedIDs []int64
	try.To(json.Unmarshal(fileContents, &usedIDs))
	count := len(posts) - len(usedIDs)
	randomIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(count)
	slog.Info("Picking random post", "index", randomIndex)
	var chosenPost *Post
	var filteredIndex = -1
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
	try.To(os.WriteFile(filePath, try.To1(json.Marshal(usedIDs)), 0644))
	return *chosenPost
}
