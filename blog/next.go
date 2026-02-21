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

const UsedBlogsIDsPath = base.DataPath + "/used.json"

func getRandomIndex(filteredCount, totalCount int64, accept func(int) bool) int {
	randomIndex := int(try.To1(rand.Int(rand.Reader, big.NewInt(filteredCount))).Int64())
	filteredIndex := -1

	for index := 0; index < int(totalCount); index++ {
		if accept(index) {
			filteredIndex++
		}

		if filteredIndex == randomIndex {
			return index
		}
	}

	return 0
}

func getUsedIDs[V int64 | string](filePath string, totalCount int64) (usedIDs []V, filteredCount int64) {
	slog.Debug("Fetching used ids", "path", filePath)

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		slog.Info("Used ids file does not exist, creating...", "path", filePath)
		try.To(os.WriteFile(filePath, []byte("[]"), base.WritePerm))
	}

	fileContents := try.To1(os.ReadFile(filePath))

	try.To(json.Unmarshal(fileContents, &usedIDs))

	filteredCount = totalCount - int64(len(usedIDs))

	// all ids are used (or used count exceeds total, e.g. posts were removed), reset
	if filteredCount <= 0 {
		filteredCount = totalCount
		usedIDs = make([]V, 0)
	}

	return usedIDs, filteredCount
}

func ChooseNextPost(posts base.Collection, usedBlogsIDsPath string) base.Post {
	usedBlogIDs, filteredBlogsCount := getUsedIDs[string](usedBlogsIDsPath, int64(len(posts)))

	blogIDs := make([]string, 0)
	for key := range posts {
		if len(posts[key].Posts) > 0 {
			blogIDs = append(blogIDs, key)
		} else {
			slog.Warn("Skipping blog with no posts", "id", key)
		}
	}

	if len(blogIDs) == 0 {
		slog.Error("No blogs with posts available")

		return base.Post{}
	}

	// Re-count filtered blogs among those with posts
	filteredBlogsCount = min(filteredBlogsCount, int64(len(blogIDs)))
	if filteredBlogsCount <= 0 {
		filteredBlogsCount = int64(len(blogIDs))
		usedBlogIDs = make([]string, 0)
	}

	randomBlogIndex := getRandomIndex(filteredBlogsCount, int64(len(blogIDs)), func(i int) bool {
		return !slices.Contains(usedBlogIDs, blogIDs[i])
	})
	blogID := blogIDs[randomBlogIndex]

	slog.Info("Choosing blog", "id", blogID)

	usedIDs, filteredPostCount := getUsedIDs[int64](posts[blogID].UsedIDsPath, int64(len(posts[blogID].Posts)))

	slog.Debug("Unused post ids", "count", filteredPostCount)

	randomPostIndex := getRandomIndex(filteredPostCount, int64(len(posts[blogID].Posts)), func(i int) bool {
		return !slices.Contains(usedIDs, posts[blogID].Posts[i].ID)
	})

	slog.Info("Picking random post", "index", randomPostIndex)
	chosenPost := &posts[blogID].Posts[randomPostIndex]
	chosenPost.Author = getFetchers()[blogID].author

	usedIDs = append(usedIDs, chosenPost.ID)
	usedBlogIDs = append(usedBlogIDs, blogID)

	try.To(os.WriteFile(posts[blogID].UsedIDsPath, try.To1(json.Marshal(usedIDs)), base.WritePerm))
	try.To(os.WriteFile(usedBlogsIDsPath, try.To1(json.Marshal(usedBlogIDs)), base.WritePerm))

	return *chosenPost
}
