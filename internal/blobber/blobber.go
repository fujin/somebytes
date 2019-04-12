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

type BlobBucketWriter interface {
	NewWriter(ctx context.Context, key string, opts *blob.WriterOptions) (*blob.Writer, error)
}

type BlobBucketLister interface {
	List(opts *blob.ListOptions) *blob.ListIterator
}
type BlobBucket interface {
	BlobBucketWriter
	BlobBucketLister
}

type Blobber struct {
	generator  func() ([]byte, error)
	blobBucket BlobBucket
}

func New(generator func() ([]byte, error), bb BlobBucket) *Blobber {
	return &Blobber{
		generator:  generator,
		blobBucket: bb,
	}
}

func (b *Blobber) CreateObjects(ctx context.Context, limit int) error {
	// Seed the pseudo-random number generator.
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < limit; i++ {
		key := fmt.Sprintf("somebytes-%v.txt", rand.Int63())

		w, err := b.blobBucket.NewWriter(ctx, key, nil)
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
			return errors.Wrap(err, "wrote incorrect number of bytes")
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
	iter := b.blobBucket.List(nil)
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
