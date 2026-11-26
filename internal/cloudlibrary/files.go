package cloudlibrary

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"time"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/publish"
)

const (
	maxProxyBytes = 32 << 20
	fetchTimeout  = 30 * time.Second
)

var ErrTooLargeToProxy = errors.New("the file is served from a signed link, not through the server")

type snapshot struct {
	tree   tree
	bucket objectstore.Bucket
}

type libraryFS struct {
	library *Library
}

func (files libraryFS) Open(name string) (fs.File, error) {
	current := files.library.current()
	entry, err := current.lookup("open", name)
	if err != nil {
		return nil, err
	}
	info := entryInfo{node: entry, modifiedAt: current.tree.modifiedAt}
	if entry.isDir {
		return &openDir{info: info, entries: current.children(name, entry)}, nil
	}
	if entry.sizeBytes > maxProxyBytes {
		return nil, &fs.PathError{Op: "open", Path: name, Err: ErrTooLargeToProxy}
	}
	body, err := current.fetch(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	return &openFile{info: info, Reader: bytes.NewReader(body)}, nil
}

func (files libraryFS) Stat(name string) (fs.FileInfo, error) {
	current := files.library.current()
	entry, err := current.lookup("stat", name)
	if err != nil {
		return nil, err
	}
	return entryInfo{node: entry, modifiedAt: current.tree.modifiedAt}, nil
}

func (files libraryFS) ReadDir(name string) ([]fs.DirEntry, error) {
	current := files.library.current()
	entry, err := current.lookup("readdir", name)
	if err != nil {
		return nil, err
	}
	if !entry.isDir {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	return current.children(name, entry), nil
}

func (current snapshot) lookup(operation, name string) (*node, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: operation, Path: name, Err: fs.ErrInvalid}
	}
	entry, found := current.tree.nodes[name]
	if !found {
		return nil, &fs.PathError{Op: operation, Path: name, Err: fs.ErrNotExist}
	}
	return entry, nil
}

func (current snapshot) children(dir string, entry *node) []fs.DirEntry {
	listed := make([]fs.DirEntry, len(entry.children))
	for position, name := range entry.children {
		listed[position] = entryInfo{node: current.tree.nodes[path.Join(dir, name)], modifiedAt: current.tree.modifiedAt}
	}
	return listed
}

func (current snapshot) fetch(name string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()
	object, err := current.bucket.Open(ctx, publish.CoursesPrefix+name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = object.Close() }()
	body, err := io.ReadAll(io.LimitReader(object, maxProxyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s from the bucket: %w", name, err)
	}
	return body, nil
}

type openFile struct {
	*bytes.Reader
	info entryInfo
}

func (file *openFile) Stat() (fs.FileInfo, error) { return file.info, nil }
func (file *openFile) Close() error               { return nil }

type openDir struct {
	info    entryInfo
	entries []fs.DirEntry
	offset  int
}

func (dir *openDir) Stat() (fs.FileInfo, error) { return dir.info, nil }
func (dir *openDir) Close() error               { return nil }

func (dir *openDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: dir.info.Name(), Err: fs.ErrInvalid}
}

func (dir *openDir) ReadDir(count int) ([]fs.DirEntry, error) {
	remaining := dir.entries[dir.offset:]
	if count <= 0 {
		dir.offset = len(dir.entries)
		return remaining, nil
	}
	if len(remaining) == 0 {
		return nil, io.EOF
	}
	taken := remaining[:min(count, len(remaining))]
	dir.offset += len(taken)
	return taken, nil
}
