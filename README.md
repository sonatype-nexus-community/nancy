<p align="center">
    <img src="https://github.com/sonatype-nexus-community/nancy/blob/master/docs/images/nancy.png" width="350"/>
</p>

<p align="center">
    <a href="https://gitter.im/sonatype-nexus-community/nancy?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge">
    <img src="https://badges.gitter.im/sonatype-nexus-community/nancy.svg" alt="Gitter"></img></a>
    <a href="https://travis-ci.org/sonatype-nexus-community/nancy"><img src="https://travis-ci.org/sonatype-nexus-community/nancy.svg?branch=master" alt="Build Status"></img></a>
</p>

# Nancy

`nancy` is a tool to check for vulnerabilities in your Golang dependencies, powered by [Sonatype OSS Index](https://ossindex.sonatype.org/).

### Usage

```
 ~ > nancy
Usage:
go list -m all | nancy [options]
nancy [options] </path/to/Gopkg.lock>
nancy [options] </path/to/go.sum>

Options:
  -clean-cache
    	Deletes local cache directory
  -exclude-vulnerability value
        Comma separated list of CVEs to exclude
  -exclude-vulnerability-file string
        Path to a file containing newline separated CVEs to be excluded (default "./.nancy-ignore")
  -no-color
        indicate output should not be colorized
  -noColor
        indicate output should not be colorized (deprecated: please use no-color)
  -quiet
        indicate output should contain only packages with vulnerabilities
  -version
        prints current nancy version
```

`nancy` currently works for projects that use `dep` or `go mod` for dependencies.

### Options

#### Quiet mode

You can run `nancy` in a quiet manner, only getting back a list of vulnerable components by running:

* `./nancy -quiet /path/to/your/Gopkg.lock `
* `./nancy -quiet /path/to/your/go.sum `

#### Exclude vulnerabilities

Sometimes you'll run into a dependency that after taking a look at, you either aren't affected by, or cannot resolve for some reason. Nancy understands, and will let you 
exclude these vulnerabilities so you can get back to a passing build:

Vulnerabilities excluded will then be silenced and not show up in the output or fail your build.

We support exclusion of vulnerability either by CVE-ID (ex: `CVE-2018-20303`) or via the OSS Index ID (ex: `a8c20c84-1f6a-472a-ba1b-3eaedb2a2a14`) as not all vulnerabilities have a CVE-ID.

##### Via CLI flag
* `./nancy -exclude-vulnerability CVE-789,bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2 /path/to/your/Gopkg.lock`
* `./nancy -exclude-vulnerability CVE-789,bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2 /path/to/your/go.sum`

##### Via file
By default if a file named `.nancy-ignore` exists in the same directory that nancy is run it will use it, will no other options need to be passed.

If you would like to define the path to the file you can use the following
* `./nancy -exclude-vulnerability-file=/path/to/your/exclude-file /path/to/your/Gopkg.lock`
* `./nancy -exclude-vulnerability-file=/path/to/your/exclude-file /path/to/your/go.sum`  

The file format requires each vulnerability that you want to exclude to be on a separate line. Comments are allowed in the file as well to help provide context when needed. See an example file below.

```
# This vulnerability is coming from package xyz, we are ok with this for now
CVN-111 
CVN-123 # Mitigated the risk of this since we only use one method in this package and the affected code doesn't matter
CVN-543
``` 

### Usage in CI

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

* Clone the project `git clone github.com/sonatype-nexus-community/nancy`
* Depending on your env you may have to enable go.mod `export GO111MODULE=on`
* In the root of the project `go test ./...`
* If tests checkout go ahead and run `go build`.
* Use that binary where ever your heart so desires!

### Download release binary

Each commit to master creates a new release binary, and if you'd like to skip building from source, you can download a binary similar to:

```console
$ curl -o /path/where/you/want/nancy \
  https://github.com/sonatype-nexus-community/nancy/releases/download/0.0.4/nancy-linux.amd64-0.0.4
```

## Development

`nancy` is written using Golang 1.13, so it is best you start there.

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

* Chat with us on [Gitter](https://gitter.im/sonatype-nexus-community/nancy)
