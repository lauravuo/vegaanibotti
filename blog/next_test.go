package blog_test

import (
	"os"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/blog/base"
)

const (
	testDataPath    = "./test_data"
	usedBlogIDsPath = testDataPath + "/used.json"
	usedIDsPath     = testDataPath + "/cc/used.json"
	kkUsedIDsPath   = testDataPath + "/kk/used.json"
	ccTestDataPath  = testDataPath + "/cc"
	kkTestDataPath  = testDataPath + "/kk"
)

func setup() {
	try.To(os.MkdirAll(ccTestDataPath, 0o700))
}

func teardown() {
	os.RemoveAll(testDataPath)
}

func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	teardown()

	os.Exit(code)
}

func TestChooseNextPost(t *testing.T) {
	t.Parallel()

	// test when empty used ids
	posts := base.Collection{
		"cc": {
			Posts: []base.Post{{
				ID:          1,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: usedIDsPath,
		},
	}
	nextPost := blog.ChooseNextPost(posts, usedBlogIDsPath)

	if nextPost.ID != posts["cc"].Posts[0].ID {
		t.Errorf("Mismatch with expected post id %d (%d)", nextPost.ID, posts["cc"].Posts[0].ID)
	}

	// test when one of the ids used
	posts = base.Collection{
		"cc": {
			Posts: []base.Post{
				{ID: 1, Title: "title", Description: "description", URL: "https://example.com", Hashtags: []string{"food"}, Added: true},

				{ID: 2, Title: "title", Description: "description", URL: "https://example.com", Hashtags: []string{"food"}, Added: true},
			},
			UsedIDsPath: usedIDsPath,
		},
	}

	try.To(os.WriteFile(usedIDsPath, []byte("[1]"), base.WritePerm))

	nextPost = blog.ChooseNextPost(posts, usedBlogIDsPath)

	if nextPost.ID != posts["cc"].Posts[1].ID {
		t.Errorf("Mismatch with expected post id %d (%d)", nextPost.ID, posts["cc"].Posts[1].ID)
	}

	// test when all of the ids used
	contents := try.To1(os.ReadFile(usedIDsPath))
	if string(contents) != "[1,2]" {
		t.Errorf("Mismatch with expected ids %s", string(contents))
	}

	nextPost = blog.ChooseNextPost(posts, usedBlogIDsPath)
	expected := "[1]"

	if nextPost.ID == 2 {
		expected = "[2]"
	}

	contents = try.To1(os.ReadFile(usedIDsPath))

	if string(contents) != expected {
		t.Errorf("Mismatch with expected ids %s (%s)", string(contents), expected)
	}
}

func TestChooseNextPostNoBlogsWithPosts(t *testing.T) {
	t.Parallel()

	// All blogs have empty post lists - should return empty Post
	posts := base.Collection{
		"cc": {
			Posts:       []base.Post{},
			UsedIDsPath: testDataPath + "/cc/used_noblogs.json",
		},
	}

	nextPost := blog.ChooseNextPost(posts, testDataPath+"/used_noblogs.json")

	if nextPost.ID != 0 {
		t.Errorf("Expected empty post, got id %d", nextPost.ID)
	}
}

func TestChooseNextPostSkipsEmptyBlog(t *testing.T) {
	t.Parallel()

	try.To(os.MkdirAll(kkTestDataPath, 0o700))

	// "kk" has no posts, only "cc" should be chosen
	posts := base.Collection{
		"cc": {
			Posts: []base.Post{{
				ID:          10,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: testDataPath + "/cc/used_skip.json",
		},
		"kk": {
			Posts:       []base.Post{},
			UsedIDsPath: kkUsedIDsPath,
		},
	}

	nextPost := blog.ChooseNextPost(posts, testDataPath+"/used_skip.json")

	if nextPost.ID != 10 {
		t.Errorf("Expected post id 10, got %d", nextPost.ID)
	}
}

func TestChooseNextPostStaleUsedBlogIDs(t *testing.T) {
	t.Parallel()

	usedPath := testDataPath + "/used_stale.json"

	// usedBlogIDsPath contains "kk" which has no posts - it should be filtered out
	// so "cc" should still be selected
	try.To(os.WriteFile(usedPath, []byte(`["kk"]`), base.WritePerm))

	posts := base.Collection{
		"cc": {
			Posts: []base.Post{{
				ID:          20,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: testDataPath + "/cc/used_stale.json",
		},
	}

	nextPost := blog.ChooseNextPost(posts, usedPath)

	if nextPost.ID != 20 {
		t.Errorf("Expected post id 20, got %d", nextPost.ID)
	}
}

func TestChooseNextPostAllBlogsUsedReset(t *testing.T) {
	t.Parallel()

	usedPath := testDataPath + "/used_reset.json"

	// Write used blog IDs with "cc" already listed - all blogs used, should reset
	try.To(os.WriteFile(usedPath, []byte(`["cc"]`), base.WritePerm))

	posts := base.Collection{
		"cc": {
			Posts: []base.Post{{
				ID:          30,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: testDataPath + "/cc/used_reset.json",
		},
	}

	nextPost := blog.ChooseNextPost(posts, usedPath)

	if nextPost.ID != 30 {
		t.Errorf("Expected post id 30 after reset, got %d", nextPost.ID)
	}
}

func TestChooseNextPostFilteringLoop(t *testing.T) {
	t.Parallel()

	try.To(os.MkdirAll(kkTestDataPath, 0o700))

	usedPath := testDataPath + "/used_filtering.json"

	// usedPath has "cc" (valid) - getUsedIDs gets totalCount=2, returns ["cc"], no reset
	// filtering loop runs and keeps "cc" → filteredBlogsCount = 2-1 = 1 > 0, picks "kk"
	try.To(os.WriteFile(usedPath, []byte(`["cc"]`), base.WritePerm))

	posts := base.Collection{
		"cc": {
			Posts: []base.Post{{
				ID:          40,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: testDataPath + "/cc/used_filtering.json",
		},
		"kk": {
			Posts: []base.Post{{
				ID:          41,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: testDataPath + "/kk/used_filtering.json",
		},
	}

	nextPost := blog.ChooseNextPost(posts, usedPath)

	if nextPost.ID != 40 && nextPost.ID != 41 {
		t.Errorf("Expected post id 40 or 41, got %d", nextPost.ID)
	}
}

func TestChooseNextPostFilteringLoopReset(t *testing.T) {
	t.Parallel()

	try.To(os.MkdirAll(kkTestDataPath, 0o700))

	usedPath := testDataPath + "/used_filtering_reset.json"

	// posts map has 3 blogs but "old" has no posts → totalCount=3, blogIDs=["cc","kk"]
	// usedPath has ["cc","kk"] - getUsedIDs(3) returns ["cc","kk"] (3-2=1>0, no reset)
	// filtering: both "cc" and "kk" are in blogIDs → filteredUsedBlogIDs=["cc","kk"]
	// filteredBlogsCount = 2-2 = 0 → triggers reset in ChooseNextPost
	try.To(os.WriteFile(usedPath, []byte(`["cc","kk"]`), base.WritePerm))

	posts := base.Collection{
		"cc": {
			Posts: []base.Post{{
				ID:          50,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: testDataPath + "/cc/used_filtering_reset.json",
		},
		"kk": {
			Posts: []base.Post{{
				ID:          51,
				Title:       "title",
				Description: "description",
				URL:         "https://example.com",
				Hashtags:    []string{"food"},
				Added:       true,
			}},
			UsedIDsPath: testDataPath + "/kk/used_filtering_reset.json",
		},
		"old": {
			Posts:       []base.Post{},
			UsedIDsPath: testDataPath + "/old/used.json",
		},
	}

	nextPost := blog.ChooseNextPost(posts, usedPath)

	if nextPost.ID != 50 && nextPost.ID != 51 {
		t.Errorf("Expected post id 50 or 51 after reset, got %d", nextPost.ID)
	}
}

