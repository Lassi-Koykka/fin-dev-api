package postings

import (
	"testing"

	"github.com/lassi-koykka/fin-dev-api/src/utils/fileutils"
)

func TestFetchAndProcessPostings(t *testing.T) {
	keywords := fileutils.ParseFileLines("../../keywords/technologies.txt")
	result := FetchAndProcessPostings(keywords)
	if len(result) < 1 {
		t.Error("No results returned")
	}
}
