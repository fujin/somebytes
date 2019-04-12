package blobber

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/fujin/somebytes/internal/loremipsum"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
)

type mockBlobBucket struct {
	newWriterErr         error
	newWriterWriteCloser io.WriteCloser
	listIterator         listIterator
}

type mockWriteCloser struct {
	writeInt int
	writeErr error
	closeErr error
}

func (m *mockWriteCloser) Write([]byte) (int, error) {
	return m.writeInt, m.writeErr
}

func (m *mockWriteCloser) Close() error {
	return m.closeErr
}

func (m *mockBlobBucket) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return m.newWriterWriteCloser, m.newWriterErr
}

func (m *mockBlobBucket) List() listIterator {
	return m.listIterator
}

type mockListIterator struct {
	listIteratorFinished   bool
	listIteratorBlobObject *blob.ListObject
	listIteratorError      error
}

func (m *mockListIterator) Next(context.Context) (*blob.ListObject, error) {
	if m.listIteratorFinished {
		return nil, io.EOF
	}

	m.listIteratorFinished = true
	return m.listIteratorBlobObject, m.listIteratorError
}

func TestCreateObjects(t *testing.T) {
	tests := []struct {
		name  string
		mb    *Blobber
		e     string
		limit int
	}{
		{
			name: "CreateObjectsNoErrorIfZero",
			mb: &Blobber{
				generator: loremipsum.RandomCharacters,
				blobBucket: &mockBlobBucket{
					newWriterErr: errors.New("could not create new writer"),
				},
			},
		},
		{
			name: "CreateObjectsCouldNotCreateWriter",
			mb: &Blobber{
				generator: loremipsum.RandomCharacters,
				blobBucket: &mockBlobBucket{
					newWriterErr: errors.New("could not create new writer"),
				},
			},
			e:     "could not create new writer",
			limit: 2,
		},
		{
			name: "CreateObjectsBadGenerator",
			mb: &Blobber{
				generator:  func() ([]byte, error) { return nil, errors.New("bad generator") },
				blobBucket: &mockBlobBucket{},
			},
			e:     "bad generator",
			limit: 2,
		},

		{
			name: "CreateObjectsWriteFailed",
			mb: &Blobber{
				generator: loremipsum.RandomCharacters,
				blobBucket: &mockBlobBucket{
					newWriterWriteCloser: &mockWriteCloser{
						writeErr: errors.New("write failed"),
					},
				},
			},
			e:     "write failed",
			limit: 2,
		},

		{
			name: "CreateObjectsIncorrectBytes",
			mb: &Blobber{
				generator: loremipsum.RandomCharacters,
				blobBucket: &mockBlobBucket{
					newWriterWriteCloser: &mockWriteCloser{},
				},
			},
			e:     "wrote incorrect number of bytes",
			limit: 2,
		},

		{
			name: "CreateObjectsFailToCloseWriter",
			mb: &Blobber{
				generator: func() ([]byte, error) { return nil, nil },
				blobBucket: &mockBlobBucket{
					newWriterWriteCloser: &mockWriteCloser{
						closeErr: errors.New("who left this open"),
					},
				},
			},
			e:     "who left this open",
			limit: 2,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mb.CreateObjects(ctx, tt.limit)
			if err != nil {
				if len(tt.e) > 0 {
					if strings.Contains(err.Error(), tt.e) {
						return
					}
					t.Fatalf("did not find %q in error %q", tt.e, err)
				}
				t.Fatalf("CreateObjects() unexpected error: %s", err)
			}

			if len(tt.e) > 0 {
				t.Fatalf("error %q did not occur as expected", tt.e)
			}
		})

	}
}

func TestListObjects(t *testing.T) {
	tests := []struct {
		name      string
		mb        *Blobber
		e         string
		threshold int
		expected  *Object
	}{
		{
			name: "ListObjectsErrorOnNext",
			mb: &Blobber{
				generator: loremipsum.RandomCharacters,
				blobBucket: &mockBlobBucket{
					listIterator: &mockListIterator{
						listIteratorError: errors.New("ran out of objects bro"),
					},
				},
			},
			threshold: 1024,
			e:         "ran out of objects",
		},

		{
			name: "ListObjectsOverThreshold",
			mb: &Blobber{
				generator: loremipsum.RandomCharacters,
				blobBucket: &mockBlobBucket{
					listIterator: &mockListIterator{
						listIteratorBlobObject: &blob.ListObject{
							Key:  "boom shakalaka",
							Size: 1025,
						},
					},
				},
			},
			threshold: 1024,
			expected: &Object{
				Key:  "boom shakalaka",
				Size: 1025,
			},
		},

		{
			name: "ListObjectsBelowThreshold",
			mb: &Blobber{
				generator: loremipsum.RandomCharacters,
				blobBucket: &mockBlobBucket{
					listIterator: &mockListIterator{
						listIteratorBlobObject: &blob.ListObject{
							Key:  "oh carolina",
							Size: 1023,
						},
					},
				},
			},
			threshold: 1024,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects, err := tt.mb.ListObjects(ctx, tt.threshold)
			if err != nil {
				if len(tt.e) > 0 {
					if strings.Contains(err.Error(), tt.e) {
						return
					}
					t.Fatalf("did not find %q in error %q", tt.e, err)
				}
				t.Fatalf("ListObjects() unexpected error: %s", err)
			}

			if len(tt.e) > 0 {
				t.Fatalf("error %q did not occur as expected", tt.e)
			}

			if tt.expected != nil {
				if len(objects) != 1 {
					t.Fatal("did not get expected object")
				}

				if objects[0].Key != tt.expected.Key {
					t.Fatalf("key = %q, want %q", objects[0].Key, tt.expected.Key)
				}
				if objects[0].Size != tt.expected.Size {
					t.Fatalf("size = %q, want %q", objects[0].Size, tt.expected.Size)
				}
			}

			if tt.expected == nil && len(objects) > 0 {
				t.Fatalf("expected no objects, but received some. objects = %#v", objects)
			}
		})

	}

}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		generator func() ([]byte, error)
		bb        BlobBucket
		e         string
	}{
		{
			name:      "New",
			generator: loremipsum.RandomCharacters,
			bb:        &blob.Bucket{},
		},
		{
			name:      "NewNoGenerator",
			generator: nil,
			bb:        &blob.Bucket{},
			e:         "generator function cannot be nil",
		},

		{
			name:      "NewNoBlobBUcket",
			generator: loremipsum.RandomCharacters,
			bb:        nil,
			e:         "blob bucket cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.generator, tt.bb)
			if err != nil {
				if len(tt.e) > 0 {
					if strings.Contains(err.Error(), tt.e) {
						return
					}
					t.Fatalf("did not find %q in error %q", tt.e, err)
				}
				t.Fatalf("New() unexpected error: %s", err)
			}

			if len(tt.e) > 0 {
				t.Fatalf("error %q did not occur as expected", tt.e)
			}

			if b == nil {
				t.Fatal("b is nil. could not create Blobber.")
			}
		})
	}
}
