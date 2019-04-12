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
	maximumRandomCharacters = 1e6
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

var opts struct {
	Number int `short:"c" description:"Set the number of objects to create." default:"10"`
	Bytes  int `short:"l" description:"List objects greater than or equal to the speficied size in bytes." default:"1024"`
	Args   struct {
		Bucket string
	} `positional-args:"yes"`
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

	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		//		log.Info("could not parse flags", rz.Err(err))
	}

	// Prefer SOMEBYTES_BUCKET
	bucket := os.Getenv("SOMEBYTES_BUCKET")
	if len(bucket) == 0 {
		bucket = opts.Args.Bucket
	}

	if bucket == "" {
		log.Info("Bucket missing. Set environment variable SOMEBYTES_BUCKET or as first argument.")
	}

	// Acquire credentials from the runtime environment for AWS S3 -- standard environment variables.
	sess, err := session.NewSession()
	if err != nil {
		log.Error("could not acquire AWS session", rz.Err(err))
		return
	}

	// Get a background context for our bucket operations.
	ctx := context.Background()

	// Open a connection to the bucket, using the session previously created AWS.
	b, err := s3blob.OpenBucket(ctx, sess, bucket, nil)
	if err != nil {
		log.Info("could not open bucket", rz.Err(err))
	}
	defer b.Close()

	// Use the lesser known capabilities of go-flag to detect if the flag
	// has been set, but is not set to the default value. I think this is a
	// good approximation of the desired behavior.
	numberOpt := parser.FindOptionByShortName('c')
	if numberOpt.IsSet() && !numberOpt.IsSetDefault() {
		// Create objects.
		createObjects(ctx, b, opts.Number)
	}

	bytesOpt := parser.FindOptionByShortName('l')
	if bytesOpt.IsSet() && !bytesOpt.IsSetDefault() {
		// List objects greater than or equal to the bytes size flag.
		listObjects(ctx, b, opts.Bytes)
	}
}
