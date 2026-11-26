package publish

import (
	"context"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
)

func (service *Service) Remove(ctx context.Context) error {
	bucket, err := service.deps.Buckets.Bucket(ctx)
	if err != nil {
		return err
	}
	if err := service.startRemoving(); err != nil {
		return err
	}
	defer service.stopRemoving()
	if err := bucket.Delete(ctx, objectstore.IndexKey); err != nil {
		return err
	}
	objects, err := bucket.List(ctx, CoursesPrefix)
	if err != nil {
		return err
	}
	for _, object := range objects {
		if err := bucket.Delete(ctx, object.Key); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) startRemoving() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.isRemoving || (service.run != nil && service.run.State == StateRunning) {
		return ErrBusy
	}
	service.isRemoving = true
	return nil
}

func (service *Service) stopRemoving() {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.isRemoving = false
}
