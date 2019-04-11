package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bloom42/rz-go/v2"
	"github.com/bloom42/rz-go/v2/log"
	"github.com/fujin/somebytes/internal/loremipsum"
	"github.com/jessevdk/go-flags"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const (
	version                 = "0.0.1"
	defaultNumberOfObjects  = 10
	defaultBytes            = 1024
	maximumRandomCharacters = 1e6
	maximumNumberOfObjects  = 1000
)

// "A random number of characters of Lorem Ipsum" Characters are
// kind of tricky in Go. We have to assume the test writer was
// referring to a UTF-8 character, e.g. potentially multi-byte.
func randomLoremIpsumCharacters() []byte {
	// Make a string builder to randomly write Lorem Ipsum's runes into.
	var builder strings.Builder

	// Iterate up to a random length, lower than the maximum.
	for c := 0; c < rand.Intn(maximumRandomCharacters); c++ {
		builder.WriteRune(loremipsum.Chars[rand.Intn(len(loremipsum.Chars))])
	}

	// Turn the string builder into byte slice, since we need one of those for writing later.
	return []byte(builder.String())
}

type objectCreator struct {
}

func (*objectCreator) Create() {

}

type objectLister struct {
}

func createObjects(ctx context.Context, b *blob.Bucket, limit int) {
	// Seed the pseudo-random number generator.
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < limit; i++ {
		key := fmt.Sprintf("somebytes-%v.txt", rand.Int63())

		w, err := b.NewWriter(ctx, key, nil)
		if err != nil {
			log.Info("could not obtain writer", rz.Err(err))
		}

		bs := randomLoremIpsumCharacters()
		expected := len(bs)
		written, err := w.Write(bs)
		if err != nil {
			log.Error("Failed to write to bucket: %s", rz.Err(err))
		}

		if written != len(bs) {
			log.Error(
				"wrote incorrect number of bytes",
				rz.String("key", key),
				rz.Int("expected", expected),
				rz.Int("written", written),
			)
		}

		if err := w.Close(); err != nil {
			log.Error("failed to close the writer", rz.Err(err))
		}
	}
}

func listObjects(ctx context.Context, b *blob.Bucket, threshold int) {
	iter := b.List(nil)
	for {
		obj, err := iter.Next(ctx)

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Info("unexpected error during blob storage iteration", rz.Err(err))
		}

		if obj.Size >= int64(threshold) {
			log.Info(
				"object greater than threshold detected",
				rz.String("key", obj.Key),
				rz.Int64("size", obj.Size),
			)
		}
	}
}

type ObjectCreator interface {
	Create(ctx context.Context, bucket *blob.Bucket, limit int)
}

type ObjectLister interface {
	List(ctx context.Context, bucket *blob.Bucket, threshold int)
}

type Somebytes interface {
	ObjectCreator
	ObjectLister
}

var opts struct {
	Number int `short:"c" description:"Set the number of objects to create. The default is 10." default:"10"`
	Bytes  int `short:"l" description:"List objects greater than or equal to the speficied size in bytes. The default is 1024." default:"1024"`
	Args   struct {
		Bucket string
	} `positional-args:"yes"`
}

// getenv fetches key 'key' from the process' environment variables, or returns
// the fallback string value if not present.
func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value

}

func main() {
	hostname, _ := os.Hostname()

	// Bring up the ripzap logger, add a hostname context value.
	log.SetLogger(
		log.With(rz.Fields(rz.String("hostname", hostname))),
	)

	// Default to info log level.
	log.SetLogger(
		log.With(rz.Level(rz.InfoLevel)),
	)

	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}

	// Prefer SOMEBYTES_BUCKET
	bucket := getenv("SOMEBYTES_BUCKET", opts.Args.Bucket)

	if bucket == "" {
		log.Info("Bucket missing. Set environment variable SOMEBYTES_BUCKET or as first argument.")
	}

	// Acquire credentials from the runtime environment for AWS S3 -- standard environment variables.
	sess, err := session.NewSession()
	if err != nil {
		// It would be much better to handle the particular 'no
		// credentials' error here and instruct the user how to
		// fix the problem..
		log.Info("could not acquire AWS session", rz.Err(err))
	}

	// Get a background context.
	ctx := context.Background()

	// Open a connection to the bucket, using the session.
	b, err := s3blob.OpenBucket(ctx, sess, bucket, nil)
	if err != nil {
		log.Info("could not open bucket", rz.Err(err))
	}
	defer b.Close()

	// Create mode! Create objects.
	createObjects(ctx, b, opts.Number)

	// List mode. List objects greater than or equal to the bytes size flag.
	listObjects(ctx, b, opts.Bytes)
}
