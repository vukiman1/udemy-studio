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
	stopped, err := service.startRemoving()
	if err != nil {
		return err
	}
	defer service.stopRemoving()
	select {
	case <-stopped:
	case <-ctx.Done():
		return ctx.Err()
	}
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

func (service *Service) startRemoving() (<-chan struct{}, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.isRemoving {
		return nil, ErrBusy
	}
	service.isRemoving = true
	if service.run != nil && service.run.State == StateRunning {
		service.cancelLocked()
	}
	if service.done == nil {
		idle := make(chan struct{})
		close(idle)
		return idle, nil
	}
	return service.done, nil
}

func (service *Service) stopRemoving() {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.isRemoving = false
}
