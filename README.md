<p align="center">
    <img src="https://github.com/sonatype-nexus-community/nancy/blob/master/docs/images/nancy.png" width="350"/>
</p>
<p align="center">
    <a href="https://travis-ci.org/sonatype-nexus-community/nancy"><img src="https://travis-ci.org/sonatype-nexus-community/nancy.svg?branch=master" alt="Build Status"></img></a>
</p>

# Nancy

`nancy` is a tool to check for vulnerabilities in your Golang dependencies, powered by [Sonatype OSS Index](https://ossindex.sonatype.org/).

To use `nancy`, assuming you have a built version of it:

* `./nancy /path/to/your/Gopkg.lock`
* `./nancy /path/to/your/go.sum`

`nancy` currently works for projects that use `dep` or `go mod` for dependencies.

You can see an example of using `nancy` in Travis-CI at [this intentionally vulnerable repo we made](https://github.com/sonatype-nexus-community/intentionally-vulnerable-golang-project).

### DISCLAIMER

A portion of the golang ecosystem doesn't use proper versions, and instead uses a commit hash to resolve your dependency. Dependencies like this will not work with
`nancy` quite yet, as we don't have a mechanism on OSS Index to lookup vulnerabilities in that manner. 

## Why Nancy?

[Nancy Drew](https://en.wikipedia.org/wiki/Nancy_Drew) was the first female detective used extensively in literature, and gave women across the world a new hero.

This project is called `nancy` as like the great detective herself, it looks for problems you might not be aware of, and gives you the information to help put them to an end!

## Installation

At current time you have a few options:

* Build from source
* Download release binary from [here on GitHub](https://github.com/sonatype-nexus-community/nancy/releases)

### Build from source

* Run `go get -u github.com/sonatype-nexus-community/nancy`
* Nancy should now be available wherever your GOPATH is set
* Run `dep ensure` in the root of the project
* In the root of the project `go test ./...`
* If tests checkout go ahead and run `go build`.
* Use that binary where ever your heart so desires!

For the adventurous, we have `go.mod` files that enable you to build using [go modules](https://github.com/golang/go/wiki/Modules).

```
$ export GO111MODULE=on
$ go test ./...
$ go build
```

### Download release binary

Each commit to master creates a new release binary, and if you'd like to skip building from source, you can download a binary similar to:

`curl https://github.com/sonatype-nexus-community/nancy/releases/download/0.0.4/nancy-linux.amd64-0.0.4 > /path/where/you/want/nancy`

## Development

`nancy` is written using Golang 1.11, so it is best you start there.

This project also uses `dep` for dependencies, so you will need to download `dep`.

Tests can be run like `go test ./... -v`

## Contributing

We care a lot about making the world a safer place, and that's why we created `nancy`. If you as well want to
speed up the pace of software development by working on this project, jump on in! Before you start work, create
a new issue, or comment on an existing issue, to let others know you are!

## Acknowledgements

The `nancy` logo was created using a combo of [Gopherize.me](https://gopherize.me/) and good ole Photoshop. Thanks to the creators of 
Gopherize for an easy way to make a fun Gopher :)

Original Gopher designed by Renee French.

## The Fine Print

It is worth noting that this is **NOT SUPPORTED** by Sonatype, and is a contribution of ours
to the open source community (read: you!)

Remember:

* Use this contribution at the risk tolerance that you have
* Do NOT file Sonatype support tickets related to `nancy` support in regard to this project
* DO file issues here on GitHub, so that the community can pitch in

Phew, that was easier than I thought. Last but not least of all:

Have fun creating and using `nancy` and the [Sonatype OSS Index](https://ossindex.sonatype.org/), we are glad to have you here!

## Getting help

Looking to contribute to our code but need some help? There's a few ways to get information:

* Chat with us on [Gitter](https://gitter.im/sonatype/nexus-developers)
