package pagination_test

import (
	"encoding/json"
	"testing"

	"github.com/yuWorm/fba-go/core/pagination"
)

func TestNewPageDataCalculatesTotalPagesAndLinks(t *testing.T) {
	page := pagination.NewPageData([]string{"one", "two"}, 41, 2, 20, "/items")

	if page.TotalPages != 3 {
		t.Fatalf("TotalPages = %d, want 3", page.TotalPages)
	}
	if page.Links.First != "/items?page=1&size=20" {
		t.Fatalf("first link = %q", page.Links.First)
	}
	if page.Links.Last != "/items?page=3&size=20" {
		t.Fatalf("last link = %q", page.Links.Last)
	}
	if page.Links.Self != "/items?page=2&size=20" {
		t.Fatalf("self link = %q", page.Links.Self)
	}
	if page.Links.Next == nil || *page.Links.Next != "/items?page=3&size=20" {
		t.Fatalf("next link = %v", page.Links.Next)
	}
	if page.Links.Prev == nil || *page.Links.Prev != "/items?page=1&size=20" {
		t.Fatalf("prev link = %v", page.Links.Prev)
	}
}

func TestPageDataUsesSnakeCaseJSONFields(t *testing.T) {
	page := pagination.NewPageData([]string{}, 0, 1, 20, "")
	got, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	const want = `{"items":[],"total":0,"page":1,"size":20,"total_pages":0,"links":{"first":"?page=1\u0026size=20","last":"?page=1\u0026size=20","self":"?page=1\u0026size=20","next":null,"prev":null}}`
	if string(got) != want {
		t.Fatalf("PageData JSON = %s, want %s", got, want)
	}
}
