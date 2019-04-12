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

// blobBucketWriter represents the internal interface between upstream *blob.Bucket#NewWriter().
type blobBucketWriter interface {
	NewWriter(ctx context.Context, key string) (io.WriteCloser, error)
}

// blobBucketLister represents the internal interface between upstream *blob.Bucket#List().
type blobBucketLister interface {
	List() listIterator
}

// listIterator represents the internal interface between upstream *blob.ListIterator#Next()
type listIterator interface {
	Next(context.Context) (*blob.ListObject, error)
}

// blobBucket is a convenience interface
type blobBucket interface {
	blobBucketWriter
	blobBucketLister
}

// BlobBucket represents the public interface of the upstream *blob.Bucket
type BlobBucket interface {
	List(*blob.ListOptions) *blob.ListIterator
	NewWriter(context.Context, string, *blob.WriterOptions) (*blob.Writer, error)
}

// blobBucketContainer is an internal structure which allows us to decorate/delegate method calls to the upstream implementation.
type blobBucketContainer struct {
	blobBucket BlobBucket
}

// NewWriter is our internal decorated method of upstream NewWriter.
func (b blobBucketContainer) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return b.blobBucket.NewWriter(ctx, key, nil)
}

// List is our internal decorated method of upstream List().
func (b blobBucketContainer) List() listIterator {
	return b.blobBucket.List(nil)
}

// Blobber allows us to create an encapsulated *blob.Bucket, while carefully
// swapping out some of its fields, to allow for testing. The public interfaces
// do not easily allow for this.
type Blobber struct {
	generator  func() ([]byte, error)
	blobBucket blobBucket
}

// New creates a Blobber object with error checking for the string generator, and the upstream *blob.Bucket.
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

// CreateObjects instructs the Blobber to seed PRNG and create individual
// randomized keys in the bucket, up to the limit. Upstream errors are wrapped
// for the call-site to handle. Additionally, the expected number of bytes
// written is checked.
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

// Object is our internal version of objects returned from ListObjects. It's
// used primarily for display and testing purposes.
type Object struct {
	Key  string
	Size int64
}

// ListObjects instructs the Blobber to query the bucket for objects, followed
// by determining if said object is greater than the threshold (in bytes). A
// slice of internal Object types is returned for display purposes.
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
