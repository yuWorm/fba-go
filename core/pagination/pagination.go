package pagination

import "fmt"

type PageData[T any] struct {
	Items      []T       `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	Size       int       `json:"size"`
	TotalPages int       `json:"total_pages"`
	Links      PageLinks `json:"links"`
}

type PageLinks struct {
	First string  `json:"first"`
	Last  string  `json:"last"`
	Self  string  `json:"self"`
	Next  *string `json:"next"`
	Prev  *string `json:"prev"`
}

func NewPageData[T any](items []T, total int64, page int, size int, basePath string) PageData[T] {
	if items == nil {
		items = []T{}
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}

	lastPage := totalPages
	if lastPage < 1 {
		lastPage = 1
	}

	links := PageLinks{
		First: pageLink(basePath, 1, size),
		Last:  pageLink(basePath, lastPage, size),
		Self:  pageLink(basePath, page, size),
	}
	if totalPages > 0 && page < totalPages {
		next := pageLink(basePath, page+1, size)
		links.Next = &next
	}
	if page > 1 {
		prev := pageLink(basePath, page-1, size)
		links.Prev = &prev
	}

	return PageData[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
		Links:      links,
	}
}

func pageLink(basePath string, page int, size int) string {
	return fmt.Sprintf("%s?page=%d&size=%d", basePath, page, size)
}
