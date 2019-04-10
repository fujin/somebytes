package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/fujin/somebytes/internal/loremipsum"
	"github.com/urfave/cli"
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

func createObjects(ctx context.Context, b *blob.Bucket, limit int) {
	// Seed the pseudo-random number generator.
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < limit; i++ {
		key := fmt.Sprintf("somebytes-%v.txt", rand.Int63())

		w, err := b.NewWriter(ctx, key, nil)
		if err != nil {
			log.Fatalf("Could not obtain writer: %s", err)
		}

		bs := randomLoremIpsumCharacters()
		expected := len(bs)
		written, err := w.Write(bs)
		if err != nil {
			log.Fatalf("Failed to write to bucket: %s", err)
		}

		if written != len(bs) {
			log.Fatalf("Wrote incorrect number of bytes to key:", key, "expected:", expected, "got:")
		}

		if err := w.Close(); err != nil {
			log.Fatalf("Failed to close the writer: %s", err)
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
			log.Fatal(err)
		}

		if obj.Size >= int64(threshold) {
			log.Println(obj.Key, obj.Size)
		}
	}
}

func main() {
	app := cli.NewApp()

	app.Version = version

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "c, number",
			Value: defaultNumberOfObjects,
			Usage: "Set the `NUMBER` of objects to create. The default is 10.",
		},
		cli.IntFlag{
			Name:  "l, bytes",
			Value: defaultBytes,
			Usage: "List objects greater than or equal to the specified size in `BYTES`. The default is 1024.",
		},
		cli.StringFlag{
			Name:   "bucket",
			Hidden: true,
			EnvVar: "SOMEBYTES_BUCKET",
		},
	}

	app.Action = func(c *cli.Context) error {
		// If the environment variable SOMEBYTES_BUCKET is set, use that.
		bucket := c.String("bucket")

		// Otherwise, if an argument has been supplied, prefer that for the bucket.
		if c.NArg() > 0 {
			bucket = c.Args().Get(0)
		}

		if bucket == "" {
			return cli.NewExitError("Bucket missing. Set environment variable SOMEBYTES_BUCKET or as first argument.", -1)
		}

		// Acquire credentials from the runtime environment for AWS S3 -- standard environment variables.
		sess, err := session.NewSession()
		if err != nil {
			// It would be much better to handle the particular 'no
			// credentials' error here and instruct the user how to
			// fix the problem..
			return cli.NewExitError(err, -2)
		}

		// Get a background context.
		ctx := context.Background()

		// Open a connection to the bucket, using the session.
		b, err := s3blob.OpenBucket(ctx, sess, bucket, nil)
		if err != nil {
			return cli.NewExitError(err, -3)
		}
		defer b.Close()

		// Create mode! Create objects.
		createObjects(ctx, b, c.Int("number"))

		// List mode. List objects greater than or equal to the bytes size flag.
		listObjects(ctx, b, c.Int("bytes"))

		return nil
	}

	app.ArgsUsage = "[bucket]"

	app.Name = "somebytes"
	app.Usage = "create and query objects in a supported object/blob storage bucket"

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
