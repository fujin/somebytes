package blobber

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"gocloud.dev/blob"
)

type blobBucketWriter interface {
	NewWriter(ctx context.Context, key string) (io.WriteCloser, error)
}

type listIterator interface {
	Next(context.Context) (*blob.ListObject, error)
}

type blobBucketLister interface {
	List() listIterator
}

type blobBucket interface {
	blobBucketWriter
	blobBucketLister
}

type BlobBucket interface {
	List(*blob.ListOptions) *blob.ListIterator
	NewWriter(context.Context, string, *blob.WriterOptions) (*blob.Writer, error)
}

type blobBucketContainer struct {
	blobBucket BlobBucket
}

func (b blobBucketContainer) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return b.blobBucket.NewWriter(ctx, key, nil)
}

func (b blobBucketContainer) List() listIterator {
	return b.blobBucket.List(nil)
}

type Blobber struct {
	generator  func() ([]byte, error)
	blobBucket blobBucket
}

func New(generator func() ([]byte, error), bb BlobBucket) (*Blobber, error) {
	if generator == nil {
		return nil, errors.New("generator function cannot be nil")
	}
	if bb == nil {
		return nil, errors.New("blob bucket cannot be nil")
	}
	return &Blobber{
		generator:  generator,
		blobBucket: blobBucketContainer{bb},
	}, nil
}

func (b *Blobber) CreateObjects(ctx context.Context, limit int) error {
	// Seed the pseudo-random number generator.
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < limit; i++ {
		key := fmt.Sprintf("somebytes-%v.txt", rand.Int63())

		w, err := b.blobBucket.NewWriter(ctx, key)
		if err != nil {
			return errors.Wrap(err, "could not create bucket writer")
		}

		bs, err := b.generator()
		if err != nil {
			return errors.Wrap(err, "generating random characters failed")
		}
		expected := len(bs)
		written, err := w.Write(bs)
		if err != nil {
			return errors.Wrap(err, "could not write to bucket")
		}

		if expected != written {
			return errors.New("wrote incorrect number of bytes")
		}

		if err := w.Close(); err != nil {
			return errors.Wrap(err, "failed to close the blob bucket writer")
		}
	}

	return nil
}

type Object struct {
	Key  string
	Size int64
}

func (b *Blobber) ListObjects(ctx context.Context, threshold int) ([]Object, error) {
	var objects []Object
	iter := b.blobBucket.List()
	for {
		obj, err := iter.Next(ctx)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, errors.Wrap(err, "unexpected error during blob storage iteration")
		}

		if obj.Size >= int64(threshold) {
			objects = append(objects, Object{Key: obj.Key, Size: obj.Size})
		}
	}

	return objects, nil
}
