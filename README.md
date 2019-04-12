# somebytes

See [spec.md](spec.md)

## Introduction

Hi! If you are reading this, it means you are reviewing my attempt at the Scytale somebytes takehome test. I appreciate the opportunity.

I have implemented the application as specified with Go. I decided to use the [Go
Cloud Development kit](https://gocloud.dev/) such that the code would be portable without any additional work between
AWS S3, GCP and Azure, in-memory and file-blacked blob storage. The facilities should readily support unit, integration, end-to-end testing.

I noticed a couple of things in the spec which required some non-intuitive choices. Mostly, the application options (-c / -l) are more like subcommands. The manual page suggests that they have a default value, but this conflicts with using them as a command selector.

I chose to select the mode of operation based on the flag being set, and not being set to the default (which kind of makes the defaults pointless). This feels like an OK compromise and jessevdk's 'go-flags' package makes this possible.

For the logger, I selected ripzap, as it's by far the highest performing and has zero allocations - although this is obviously not much of a concern for a tool like this.

I've included an 'internal'-type package named loremipsum which takes care of the specified random number of characters from the Lorem ipsum text.

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

## Testing
Some basic unit test coverage is included. The surface is very small, and dependencies extensively unit tested/benchmarked.

### Unit
`go test ./...`
