package masterdata

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// Loader reads downloaded MementoMori master-book files from disk.
// It is intentionally independent from internal/master, which owns syncing
// downloaded files. API code should depend on this package instead of reaching
// into downloader internals.
type Loader struct {
	Root string
}

type BookInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func New(root string) Loader {
	return Loader{Root: root}
}

func (l Loader) ListBooks() ([]BookInfo, error) {
	entries, err := os.ReadDir(l.Root)
	if err != nil {
		return nil, fmt.Errorf("read master root %q: %w", l.Root, err)
	}
	books := make([]BookInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat master book %q: %w", entry.Name(), err)
		}
		books = append(books, BookInfo{Name: entry.Name(), Size: info.Size()})
	}
	sort.Slice(books, func(i, j int) bool { return books[i].Name < books[j].Name })
	return books, nil
}

func (l Loader) LoadRows(bookName string) ([]map[string]any, error) {
	if err := ValidateBookName(bookName); err != nil {
		return nil, err
	}
	path := filepath.Join(l.Root, bookName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read master book %q: %w", bookName, err)
	}
	var rows []map[string]any
	if err := msgpack.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("decode master book %q: %w", bookName, err)
	}
	return rows, nil
}

func ValidateBookName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\\`) {
		return fmt.Errorf("invalid master book name %q", name)
	}
	return nil
}
