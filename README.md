# somebytes

```
SOMEBYTES(1)		  BSD General Commands Manual		  SOMEBYTES(1)

NAME
     somebytes -- create and query objects in an AWS S3 bucket

SYNOPSIS
     somebytes -c [number] [bucket]
     somebytes -l [characters] [bucket]

DESCRIPTION
     The somebytes tool manages and queries objects in an AWS S3 bucket.

     In its first form, it creates a number of files in the named bucket.
     These files contain a random number of characters of Lorem Ipsum.

     In its second form, it returns a list of objects and their sizes in the
     named bucket equal to or greater than a specified number of bytes.

OPTIONS
     -c [number]
	     Set the number of objects to create. The default is 10.

     -l [bytes]
	     List objects greater than or equal to the speficied size in
	     bytes. The default is 1024.

DIAGNOSTICS
     The somebytes utility exits 0 on success, and >0 if an error occurs.

ENVIRONMENT
     The following environment variables affect the execution of somebytes:

     SOMEBYTES_BUCKET
		     If the environment variable SOMEBYTES_BUCKET is set, the
		     named AWS bucket will be used. If a bucket is specified
		     on the command line, it overrides this variable.

BSD				  Feb 2, 2018				   BSD
```

## Introduction

Hi! If you are reading this, it means you are reviewing my attempt at the Scytale somebytes takehome test. I appreciate the opportunity.

I have implemented the application as specified with the Go language. I decided to use the [Go
Cloud Development kit](https://gocloud.dev/) such that the code would be portable without much additional work between
AWS S3, GCP and Azure, in-memory and file-blacked blob storage. The facilities should readily support unit, integration, end-to-end testing. The recommended way to support multiple providers is through the new 'wire' dependency injection system, which I opted not to make use of, although would be possible to integrate later if multi-provider capabilities became necessary.

For the logger, I selected ripzap, as it's by far the highest performing and has zero allocations or reflection - although this is obviously not much of a concern for a tool like this.

I've included an 'internal'-type package named blobber, which handles interacting with the blob storage (through the go CDK), and loremipsum which takes care of the specified random number of characters from the Lorem ipsum text.

## Installation

- go version go1.12.3

## Building

- `cd cmd && go build -o somebytes`

## Running
- Set the required environment variables for access to AWS S3. They are as follows:
```
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_REGION
```

Despite the general recommendation against doing so, in the interest of the test working, I have included credentials for a bucket in ap-southeast-2. Feel free to substitute your own. I will disable these credentials once everyone has had a chance to review the submission.
```
export AWS_ACCESS_KEY_ID=AKIAZKGCQ7DUMSC3QZUU AWS_SECRET_ACCESS_KEY=cK5CkvPWw8Uxal/utndILL1cdpV5EGGzQUJq1xss AWS_REGION=ap-southeast-2
```

I suggest erasing the keys from the bucket prior to running the test e.g. with AWS CLI; running the test binary in the 'list' mode, to confirm the bucket is empty, followed by the 'create' mode, to create keys up to the limit, then finally the test binary in the 'list' mode again.

## Linting
The code was developed against the standards checked by [golangci-lint](https://golangci.com/)

To install, run:
```
# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin latest
```

Then invoke in the root of the project:
```
golangci-lint run
```

## Testing
Extensive unit test coverage is included. The surface is quite small with dependencies extensively unit tested/benchmarked.

### Unit
```
go test ./... -v
```

### Coverage
```
go test -covermode atomic -cover -coverprofile profile.out ./...
go tool cover -html=profile.out
go tool cover -func=profile.out
```
