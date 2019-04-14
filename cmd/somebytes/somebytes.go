package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/bloom42/rz-go/v2"
	"github.com/bloom42/rz-go/v2/log"

	"github.com/jessevdk/go-flags"

	"gocloud.dev/blob/s3blob"

	"github.com/fujin/somebytes/internal/blobber"
	"github.com/fujin/somebytes/internal/loremipsum"
)

var opts struct {
	Number int `short:"c" description:"Set the number of objects to create." optional:"true" optional-value:"10"`
	Bytes  int `short:"l" description:"List objects greater than or equal to the specified size in bytes." optional:"true"  optional-value:"1024"`
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
		os.Exit(1)
	}

	// Prefer SOMEBYTES_BUCKET
	bucket := os.Getenv("SOMEBYTES_BUCKET")
	if len(bucket) == 0 {
		bucket = opts.Args.Bucket
	}

	if bucket == "" {
		log.Error("Bucket missing. Set environment variable SOMEBYTES_BUCKET or as first argument.")
		os.Exit(2)
	}

	// Acquire credentials from the runtime environment for AWS S3 -- standard environment variables.
	sess, err := session.NewSession()
	if err != nil {
		log.Error("could not acquire AWS session", rz.Err(err))
		os.Exit(3)
	}

	// Get a background context for our bucket operations.
	ctx := context.Background()

	// Open a connection to the bucket, using the session previously created AWS.
	b, err := s3blob.OpenBucket(ctx, sess, bucket, nil)
	if err != nil {
		log.Error("could not open bucket", rz.Err(err))
		os.Exit(4)
	}
	defer b.Close()

	blobber, err := blobber.New(loremipsum.RandomCharacters, b)
	if err != nil {
		log.Error("could not create blobber", rz.Err(err))
		os.Exit(5)
	}

	// Create objects.
	if parser.FindOptionByShortName('c').IsSet() {
		log.Info("creating multiple objects up to limit", rz.Int("number", opts.Number))

		err := blobber.CreateObjects(ctx, opts.Number)

		if err != nil {
			log.Error("could not create objects", rz.Err(err))
			os.Exit(6)
		}
	}

	// List objects greater than or equal to the bytes size flag.
	if parser.FindOptionByShortName('l').IsSet() {
		log.Info("listing objects greater than bytes threshold", rz.Int("bytes", opts.Bytes))

		objects, err := blobber.ListObjects(ctx, opts.Bytes)
		if err != nil {
			log.Error("error during list operation", rz.Err(err))
			os.Exit(7)
		}

		for _, o := range objects {
			log.Info("object above threshold", rz.String("key", o.Key), rz.Int64("size", o.Size))
		}
	}
}
