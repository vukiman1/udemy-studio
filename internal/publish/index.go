package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
)

const (
	IndexFormat      = 1
	CoursesPrefix    = "courses/"
	indexContentType = "application/json"
)

type Index struct {
	Format      int         `json:"format"`
	PublishedAt time.Time   `json:"publishedAt"`
	Course      IndexCourse `json:"course"`
	Chapters    []string    `json:"chapters,omitempty"`
	Files       []IndexFile `json:"files"`
}

type IndexCourse struct {
	Dir          string `json:"dir"`
	Title        string `json:"title"`
	SizeBytes    int64  `json:"sizeBytes"`
	ChapterCount int    `json:"chapterCount,omitempty"`
}

func (index Index) PublishedChapters() int {
	if len(index.Chapters) == 0 {
		return index.Course.ChapterCount
	}
	return len(index.Chapters)
}

type IndexFile struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"sizeBytes"`
}

func ObjectKey(courseDir, filePath string) string {
	return CoursesPrefix + path.Join(courseDir, filePath)
}

func ReadIndex(ctx context.Context, bucket objectstore.Bucket) (Index, bool, error) {
	object, err := bucket.Open(ctx, objectstore.IndexKey)
	if errors.Is(err, objectstore.ErrObjectNotFound) {
		return Index{}, false, nil
	}
	if err != nil {
		return Index{}, false, err
	}
	defer func() { _ = object.Close() }()
	raw, err := io.ReadAll(object)
	if err != nil {
		return Index{}, false, fmt.Errorf("read %s: %w", objectstore.IndexKey, err)
	}
	var index Index
	if err := json.Unmarshal(raw, &index); err != nil {
		return Index{}, false, fmt.Errorf("decode %s: %w", objectstore.IndexKey, err)
	}
	return index, true, nil
}

func writeIndex(ctx context.Context, bucket objectstore.Bucket, index Index) error {
	raw, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("encode %s: %w", objectstore.IndexKey, err)
	}
	return bucket.Put(ctx, objectstore.Upload{
		Key:         objectstore.IndexKey,
		Body:        bytes.NewReader(raw),
		SizeBytes:   int64(len(raw)),
		ContentType: indexContentType,
	})
}
