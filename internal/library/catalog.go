package library

import (
	"io/fs"
	"sync"
	"sync/atomic"
)

type Catalog struct {
	downloads fs.FS
	current   atomic.Pointer[Library]
	// rescanning keeps a scan that started earlier from storing its older snapshot over a newer one.
	rescanning sync.Mutex
}

func NewCatalog(downloads fs.FS) (*Catalog, error) {
	catalog := &Catalog{downloads: downloads}
	if err := catalog.Rescan(); err != nil {
		return nil, err
	}
	return catalog, nil
}

func (catalog *Catalog) Current() *Library {
	return catalog.current.Load()
}

// Rescan keeps serving the previous scan when the new one fails, so a half-written download never empties the library.
func (catalog *Catalog) Rescan() error {
	catalog.rescanning.Lock()
	defer catalog.rescanning.Unlock()
	scanned, err := Scan(catalog.downloads)
	if err != nil {
		return err
	}
	catalog.current.Store(scanned)
	return nil
}
