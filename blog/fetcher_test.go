package blog_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
)

const testDataPath = "./test_data/"

func getter(url, _ string) ([]byte, error) {
	if strings.Contains(url, "1") {
		return []byte(`<article class="teaser post-19050 post type-post status-publish format-standard has-post-thumbnail hentry category-paaruoat category-salaatit tag-kimchi tag-mungpavun-idut tag-nuudeli tag-nyhtokaura tag-tahini"><div class="entry-thumbnail"> <a href="https://chocochili.net/2023/09/nyhtokaura-nuudelikulho/"><img width="300" height="200" src="data:image/gif;base64,R0lGODdhAQABAPAAAP///wAAACwAAAAAAQABAEACAkQBADs=" data-lazy-src="/app/uploads/2023/09/nyhtokaura-nuudelikulho-300x200.jpg" class="attachment-teaser size-teaser wp-post-image" alt="" loading="lazy" data-lazy-srcset="/app/uploads/2023/09/nyhtokaura-nuudelikulho-300x200.jpg 300w, /app/uploads/2023/09/nyhtokaura-nuudelikulho-700x470.jpg 700w" sizes="(max-width: 300px) 100vw, 300px"><noscript><img width="300" height="200" src="/app/uploads/2023/09/nyhtokaura-nuudelikulho-300x200.jpg" class="attachment-teaser size-teaser wp-post-image" alt="" loading="lazy" srcset="/app/uploads/2023/09/nyhtokaura-nuudelikulho-300x200.jpg 300w, /app/uploads/2023/09/nyhtokaura-nuudelikulho-700x470.jpg 700w" sizes="(max-width: 300px) 100vw, 300px"></noscript></a></div><header><h2 class="entry-title"><a href="https://chocochili.net/2023/09/nyhtokaura-nuudelikulho/">Nyhtökaura-nuudelikulho</a></h2></header><div class="entry-summary"><p>Tämä kulhoruoka sisältää mm. karamellisoitua gochujang-nyhtökauraa, nuudeleita, kimchiä ja kermaista seesamikastiketta.</p></div><footer class="entry-footer"> <a class="read-more" href="https://chocochili.net/2023/09/nyhtokaura-nuudelikulho/">Lue lisää</a></footer></article>`), nil
	}
	return nil, errors.New("Not found")
}

func setup() {
	try.To(os.MkdirAll(testDataPath, 0700))
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

func TestFetchNewPosts(t *testing.T) {
	t.Parallel()
	posts, err := blog.FetchNewPosts("./test_data/recipes.json", getter)
	if err != nil {
		t.Errorf("Expected success, got: %s", err)
	}
	if len(posts) == 0 {
		t.Errorf("Expected to find posts, got 0 posts.")
	}
}
