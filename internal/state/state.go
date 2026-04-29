package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/veirt/vivere/internal/fsutil"
)

type Store struct {
	root string
}

func New(root string) Store {
	return Store{root: filepath.Join(root, "state")}
}

func (s Store) ReadString(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(s.root, name))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read state %s: %w", name, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func (s Store) WriteString(name, value string) error {
	return fsutil.WriteFileAtomic(filepath.Join(s.root, name), []byte(strings.TrimSpace(value)+"\n"), 0o644)
}
