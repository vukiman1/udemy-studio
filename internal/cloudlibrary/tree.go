package cloudlibrary

import (
	"io/fs"
	"path"
	"slices"
	"time"

	"github.com/vukiman1/udemy-studio/internal/publish"
)

const (
	rootDir  = "."
	dirMode  = fs.ModeDir | 0o555
	fileMode = 0o444
)

type node struct {
	name      string
	isDir     bool
	sizeBytes int64
	children  []string
}

type tree struct {
	nodes      map[string]*node
	modifiedAt time.Time
}

func emptyTree() tree {
	return tree{nodes: map[string]*node{rootDir: {name: rootDir, isDir: true}}}
}

func treeOf(index publish.Index) tree {
	built := emptyTree()
	built.modifiedAt = index.PublishedAt
	for _, file := range index.Files {
		built.addFile(path.Join(index.Course.Dir, file.Path), file.SizeBytes)
	}
	for _, entry := range built.nodes {
		slices.Sort(entry.children)
	}
	return built
}

func (built tree) addFile(filePath string, sizeBytes int64) {
	built.nodes[filePath] = &node{name: path.Base(filePath), sizeBytes: sizeBytes}
	child := filePath
	for {
		parent := path.Dir(child)
		existing, found := built.nodes[parent]
		if !found {
			existing = &node{name: path.Base(parent), isDir: true}
			built.nodes[parent] = existing
		}
		existing.children = append(existing.children, path.Base(child))
		if found {
			return
		}
		child = parent
	}
}

type entryInfo struct {
	node       *node
	modifiedAt time.Time
}

func (info entryInfo) Name() string               { return info.node.name }
func (info entryInfo) Size() int64                { return info.node.sizeBytes }
func (info entryInfo) ModTime() time.Time         { return info.modifiedAt }
func (info entryInfo) IsDir() bool                { return info.node.isDir }
func (info entryInfo) Sys() any                   { return nil }
func (info entryInfo) Type() fs.FileMode          { return info.Mode().Type() }
func (info entryInfo) Info() (fs.FileInfo, error) { return info, nil }

func (info entryInfo) Mode() fs.FileMode {
	if info.node.isDir {
		return dirMode
	}
	return fileMode
}
