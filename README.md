<!--

    Copyright 2018-present Sonatype Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

-->
<p align="center">
    <img src="https://github.com/sonatype-nexus-community/nancy/blob/main/docs/images/nancy.png?raw=true" width="350" alt="nancy logo"/>
</p>

# Nancy

[![shield_gh-workflow-test]][link_gh-workflow-test]
[![shield_license]][license_file]

`nancy` is a tool to check for vulnerabilities in your Golang dependencies, powered by [Sonatype Guide](https://guide.sonatype.com/), and also works with Sonatype Lifecycle (formerly Nexus IQ Server), allowing you a smooth experience as a Golang developer!

## Table of Contents

- [Authentication](#authentication)
- [Installation](#installation)
  - [Build from source](#build-from-source)
  - [Download release binary](#download-release-binary)
  - [Install via Homebrew (macOS)](#install-via-homebrew-macos)
  - [Install from the AUR (Arch Linux)](#install-from-the-aur-arch-linux)
- [Usage](#usage)
  - [What is the best usage of Nancy?](#what-is-the-best-usage-of-nancy)
  - [CI Usage](#ci-usage)
  - [Docker usage](#docker-usage)
- [Sonatype Guide Options](#sonatype-guide-options)
  - [Rate limiting / Setting config](#rate-limiting--setting-config)
  - [Using a custom server URL](#using-a-custom-server-url)
  - [Loud mode](#loud-mode)
  - [Exclude vulnerabilities](#exclude-vulnerabilities)
  - [Output](#output)
- [Sonatype Lifecycle Options](#sonatype-lifecycle-options)
  - [Persistent Lifecycle Config](#persistent-lifecycle-config)
- [Usage in CI](#usage-in-ci)
- [Why Nancy?](#why-nancy)
  - [Relationship to govulncheck](#relationship-to-govulncheck)
- [How to Fix Vulnerabilities](#how-to-fix-vulnerabilities)
- [Development](#development)
  - [Release Process](#release-process)
- [Contributing](#contributing)
- [Acknowledgements](#acknowledgements)
- [The Fine Print](#the-fine-print)

## Authentication

Nancy v2.0.0 supports three authentication modes for vulnerability scanning via Sonatype Guide.

### Sonatype Guide Bearer Token (Recommended)

Obtain a free token from [https://guide.sonatype.com](https://guide.sonatype.com).

Use via flag:
```shell
go list -json -deps ./... | nancy sleuth --guide-token $GUIDE_TOKEN
```

Or via environment variable:
```shell
export GUIDE_TOKEN=YOUR_TOKEN
go list -json -deps ./... | nancy sleuth
```

### OSS Index Credentials (Deprecated)

OSS Index credentials (`--username` / `--token`) continue to work via the Sonatype Guide compatibility API.

> **Deprecated**: OSS Index credentials will be removed in v3.x. Please migrate to a Sonatype Guide Bearer token.

If you have existing OSS Index credentials they will continue to work, but nancy will display a deprecation warning.

```shell
nancy sleuth --username auser@anemailaddress.com --token A4@k3@p1T0k3n
```

Or via environment variables:
```shell
export OSSI_USERNAME=auser@anemailaddress.com   # Deprecated — migrate to GUIDE_TOKEN
export OSSI_TOKEN=A4@k3@p1T0k3n                 # Deprecated — migrate to GUIDE_TOKEN
go list -json -deps ./... | nancy sleuth
```

### Unauthenticated

Nancy works without any credentials but is rate-limited by the upstream service.

---

## Installation

At the current time you have a few options:

- Build from source
- Download release binary from [here on GitHub](https://github.com/sonatype-nexus-community/nancy/releases)
- Install via Homebrew (macOS)
- Install from the AUR (Arch Linux)

### Build from source

- Clone the project `git clone github.com/sonatype-nexus-community/nancy`
- In the root of the project run `make`
  - This will execute multiple targets so if you want to short circuit some of that process you can also just run `make build` to get the binary without running tests, linting, etc
- Use that binary wherever your heart so desires!

### Download release binary

Each tag pushed to this repo creates a new release binary, and if you'd like to skip building from source, you can download a binary similar to:

```console
$ curl -o /path/where/you/want/nancy \
  https://github.com/sonatype-nexus-community/nancy/releases/download/v0.0.44/nancy-darwin.amd64-v0.0.44
```

### Install via Homebrew (macOS)

On macOS, `nancy` can be installed using `brew`:

```shell
brew tap sonatype-nexus-community/tap
brew install sonatype-nexus-community/tap/nancy
```

`brew` formulae are created and published to that tap with each new release, so you can use `brew` to upgrade, etc... as you wish.

You can see more about the formulae, etc... at [this repo](https://github.com/sonatype-nexus-community/homebrew-nancy-tap).

### Install from the AUR (Arch Linux)

On Arch Linux, `nancy` can be installed using the [AUR](https://aur.archlinux.org/packages/nancy-bin/):

```shell
$ yay -S nancy-bin
```

---

## Usage

`nancy` works for projects that use `go mod` for dependencies.

> **New in v2.0.0**: Running `nancy sleuth` with no piped input will automatically invoke `go list -json -deps ./...` for you. The pipe form still works and may be preferred in CI for explicitness.

```
 ~ > nancy --help
nancy is a tool to check for vulnerabilities in your Golang dependencies,
powered by Sonatype Guide, and also works with Sonatype Lifecycle, allowing
you a smooth experience as a Golang developer!

Usage:
  nancy [flags]
  nancy [command]

Examples:
  Typical usage:
  nancy sleuth [flags]
  nancy lifecycle [flags]

  Or explicitly piping go list output:
  go list -json -deps ./... | nancy sleuth [flags]
  go list -json -deps ./... | nancy lifecycle [flags]

Available Commands:
  config      Setup credentials to use when connecting to services
  help        Help about any command
  lifecycle   Check for vulnerabilities in your Golang dependencies using Sonatype Lifecycle
  sleuth      Check for vulnerabilities in your Golang dependencies using Sonatype Guide
  update      Check if there are any updates available

Flags:
  -v, -- count                 Set log level, multiple v's is more verbose
  -c, --clean-cache            Deletes local cache directory
  -d, --db-cache-path string   Specify an alternate path for caching responses from Sonatype Guide, example: /tmp
  -h, --help                   help for nancy
      --loud                   indicate output should include non-vulnerable packages
  -q, --quiet                  indicate output should contain only packages with vulnerabilities (default true)
      --skip-update-check      Skip the check for updates.
  -t, --token string           Specify OSS Index API token for request (deprecated, use --guide-token)
  -u, --username string        Specify OSS Index username for request (deprecated, use --guide-token)
  -V, --version                Get the version

Use "nancy [command] --help" for more information about a command.


$ > nancy sleuth --help
'nancy sleuth' is a command to check for vulnerabilities in your Golang dependencies, powered by Sonatype Guide.

Usage:
  nancy sleuth [flags]

Examples:
  nancy sleuth --guide-token $GUIDE_TOKEN
  go list -json -deps ./... | nancy sleuth --guide-token $GUIDE_TOKEN

Flags:
  -a, --additional-exclude-vulnerability-files strings   Path to additional files containing newline separated CVEs or Sonatype IDs to be excluded
  -e, --exclude-vulnerability CveListFlag                Comma separated list of CVEs or Sonatype IDs to exclude (default [])
  -x, --exclude-vulnerability-file string                Path to a file containing newline separated CVEs or Sonatype IDs to be excluded (default "./.nancy-ignore")
      --guide-token string                               Specify Sonatype Guide Bearer token for request
  -h, --help                                             help for sleuth
  -n, --no-color                                         indicate output should not be colorized
  -o, --output string                                    Styling for output format. json, json-pretty, text, csv (default "text")

Global Flags:
  -v, -- count                 Set log level, multiple v's is more verbose
  -d, --db-cache-path string   Specify an alternate path for caching responses from Sonatype Guide, example: /tmp
      --loud                   indicate output should include non-vulnerable packages
  -q, --quiet                  indicate output should contain only packages with vulnerabilities (default true)
      --skip-update-check      Skip the check for updates.
  -t, --token string           Specify OSS Index API token for request (deprecated, use --guide-token)
  -u, --username string        Specify OSS Index username for request (deprecated, use --guide-token)
  -V, --version                Get the version

$ > nancy lifecycle --help
'nancy lifecycle' is a command to check for vulnerabilities in your Golang dependencies using Sonatype Lifecycle.
Note: 'nancy iq' is a deprecated alias for 'nancy lifecycle'.

Usage:
  nancy lifecycle [flags]

Examples:
  go list -json -deps ./... | nancy lifecycle --lifecycle-application your_public_application_id --lifecycle-server-url http://your_lifecycle_server_url:port --lifecycle-username your_user --lifecycle-token your_token --lifecycle-stage develop

Flags:
  -h, --help                              help for lifecycle
  -a, --lifecycle-application string      Specify Sonatype Lifecycle public application ID for request
  -x, --lifecycle-server-url string       Specify Sonatype Lifecycle server url for request (default "http://localhost:8070")
  -s, --lifecycle-stage string            Specify Sonatype Lifecycle stage for request (default "develop")
  -k, --lifecycle-token string            Specify Sonatype Lifecycle token for request (default "admin123")
  -l, --lifecycle-username string         Specify Sonatype Lifecycle username for request (default "admin")

Global Flags:
  -v, -- count                 Set log level, multiple v's is more verbose
  -d, --db-cache-path string   Specify an alternate path for caching responses from Sonatype Guide, example: /tmp
      --loud                   indicate output should include non-vulnerable packages
  -q, --quiet                  indicate output should contain only packages with vulnerabilities (default true)
      --skip-update-check      Skip the check for updates.
  -t, --token string           Specify OSS Index API token for request (deprecated, use --guide-token)
  -u, --username string        Specify OSS Index username for request (deprecated, use --guide-token)
  -V, --version                Get the version
```

> **Gopkg.lock support removed in v2.0.0**
>
> The `golang/dep` project has been archived since 2020. Nancy v2.0.0 removes Gopkg.lock scanning.
>
> **Migration**: Migrate your project to Go modules. Then scan with:
> ```
> go list -json -deps ./... | nancy sleuth
> ```
> See https://go.dev/doc/gopath_vs_modules for migration guidance.

### What is the best usage of Nancy?

The preferred way to use Nancy is simply:

- `nancy sleuth --guide-token $GUIDE_TOKEN`

Or explicitly piping `go list` output (equivalent, and preferred in CI):

- `go list -json -deps ./... | nancy sleuth --guide-token $GUIDE_TOKEN`

If you would like to scan all dependencies, including those that do not end up in the final binary, you can use
`go list -json -m all` instead:

- `go list -json -m all | nancy sleuth --guide-token $GUIDE_TOKEN`

### CI Usage

Here are some additional tools to simplify using Nancy in your CI environment:

* [Nancy CircleCI Orb](https://github.com/sonatype-nexus-community/circleci-nancy-orb)
* [Nancy GitHub Action](https://github.com/sonatype-nexus-community/nancy-github-action)

#### GitHub Actions Example

```yaml
name: Vulnerability Scan
on: [push, pull_request]

jobs:
  nancy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Scan for vulnerabilities
        run: go list -json -deps ./... | nancy sleuth --guide-token ${{ secrets.GUIDE_TOKEN }}
```

For Sonatype Lifecycle (formerly Nexus IQ) scanning:

```yaml
      - name: Scan with Sonatype Lifecycle
        run: go list -json -deps ./... | nancy lifecycle --lifecycle-application my-app-id --lifecycle-server-url ${{ vars.LIFECYCLE_URL }} --lifecycle-username ${{ secrets.LIFECYCLE_USERNAME }} --lifecycle-token ${{ secrets.LIFECYCLE_TOKEN }}
```

> **Note**: `nancy iq` is a deprecated alias for `nancy lifecycle`. The `--iq-*` flags are deprecated aliases for `--lifecycle-*` flags. Both continue to work in v2.0.0 but will be removed in v3.x.

### Docker usage

<p align="center">
    <img src="https://github.com/sonatype-nexus-community/nancy/blob/main/docs/images/nancy_docker.png" width="350" alt="nancy docker logo"/>
</p>

`nancy` now comes in a boat! For ease of use, we've dockerized `nancy`. To use our Dockerfile:

```shell
go list -json -deps ./... | docker run --rm -i sonatypecommunity/nancy:latest sleuth --guide-token $GUIDE_TOKEN
```

We publish a few different flavors for convenience:

- Latest if you want to be on the bleeding edge ex: `latest`
- The full tag for those concerned with 100% reliability of underlying Nancy ex: `v0.1.1`
- The major version (we respect semver) ex: `v0`
- The major/minor version (seriously, we respect semver) ex: `v0.1`

##### Want to build them locally??

1. Install `goreleaser` or use their provided docker image (https://goreleaser.com/install/)
2. Run `goreleaser` with the following options

    ```
    goreleaser release --skip-publish --snapshot --rm-dist
    ```

    or docker version of `goreleaser`

    ```
    docker run --privileged \
      -v $PWD:/go/src/github.com/user/repo \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -w /go/src/github.com/user/repo \
      goreleaser/goreleaser release --skip-publish --snapshot --rm-dist
    ```

3. Once complete you will have the images now built locally. Use `docker images` to see them

    ```
    > docker images                                                                                                                                                                [789c9df]
    REPOSITORY                TAG                           IMAGE ID            CREATED             SIZE
    sonatypecommunity/nancy   alpine                        f966c833c762        52 seconds ago      19.9MB
    sonatypecommunity/nancy   v1-alpine                     f966c833c762        52 seconds ago      19.9MB
    sonatypecommunity/nancy   v1.0-alpine                   f966c833c762        52 seconds ago      19.9MB
    sonatypecommunity/nancy   v1.0.0-alpine                 f966c833c762        52 seconds ago      19.9MB
    sonatypecommunity/nancy   latest                        7cb89e362115        53 seconds ago      14.1MB
    sonatypecommunity/nancy   v1                            7cb89e362115        53 seconds ago      14.1MB
    sonatypecommunity/nancy   v1.0                          7cb89e362115        53 seconds ago      14.1MB
    sonatypecommunity/nancy   v1.0.0                        7cb89e362115        53 seconds ago      14.1MB
    ```

---

## Sonatype Guide Options

### Rate limiting / Setting config

If you start using Nancy extensively without authentication, you might run into rate limiting from Sonatype Guide. To avoid this, authenticate with a Guide Bearer token (see [Authentication](#authentication) above).

If you run into rate limiting you should receive an error with instructions on how to obtain a token:

```
You have been rate limited by Sonatype Guide.
Please visit https://guide.sonatype.com to obtain a free Bearer token.
Upon retrieving your token, run nancy with --guide-token YOUR_TOKEN or set GUIDE_TOKEN=YOUR_TOKEN.
```

You can set your token via the command line:

`nancy sleuth --guide-token YOUR_TOKEN`

Or as an environment variable:

```shell
export GUIDE_TOKEN=YOUR_TOKEN
go list -json -deps ./... | nancy sleuth
```

### Using a custom server URL

If you need to use a custom or on-premise Sonatype Guide server, you can configure the URL in several ways:

1. **Via command-line flag:**
   ```shell
   nancy sleuth --ossindex-url https://custom.guide.sonatype.com
   ```

2. **Via environment variable:**
   ```shell
   export OSSI_OSSINDEXURL=https://custom.guide.sonatype.com
   go list -json -deps ./... | nancy sleuth
   ```

3. **Via legacy environment variable (for backwards compatibility):**
   ```shell
   export OSSIndexURL=https://custom.guide.sonatype.com
   go list -json -deps ./... | nancy sleuth
   ```

The priority order is: command-line flag > `OSSI_OSSINDEXURL` environment variable > `OSSIndexURL` environment variable > default URL.

### Loud mode

By default, `nancy` runs in a "quiet" mode, only displaying a list of vulnerable components.
You can run `nancy` in a loud manner, showing all components by running:

- `go list -json -deps ./... | nancy sleuth --loud`

### Exclude vulnerabilities

Sometimes you'll run into a dependency that after taking a look at, you either aren't affected by, or cannot resolve for some reason. Nancy understands, and will let you
exclude these vulnerabilities, so you can get back to a passing build:

Vulnerabilities excluded will then be silenced and not show up in the output or fail your build.

We support exclusion of vulnerability either by CVE-ID (ex: `CVE-2018-20303`) or via the Sonatype vulnerability ID (ex: `a8c20c84-1f6a-472a-ba1b-3eaedb2a2a14`) as not all vulnerabilities have a CVE-ID.

#### Via CLI flag

- `go list -json -deps ./... | nancy sleuth --exclude-vulnerability CVE-789,bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2`

#### Via file

By default, if a file named `.nancy-ignore` exists in the same directory that nancy is run it will use it - no other options need to be passed.

If you would like to define the path to the file you can use the following

- `go list -json -deps ./... | nancy sleuth --exclude-vulnerability-file=/path/to/your/exclude-file`

If you would like to split up your excludes into multiple files besides your root `.nancy-ignore` you can pass them via the `-a` or `--additional-exclude-vulnerability-files` flags.

- `nancy sleuth --additional-exclude-vulnerability-files=/path/to/first,/path/to/second`
- `nancy sleuth -a /path/to/first -a /path/to/second`

You can also combine it with the `-x` / `--exclude-vulnerability-file` flag. Nancy merges the additional files on top of the root `.nancy-ignore`.

- `nancy sleuth -x .nancy-ignore.global -a .nancy-ignore.local`

The file format requires each vulnerability that you want to exclude to be on a separate line. Comments are allowed in the file as well to help provide context when needed. See an example file below.

```
# This vulnerability is coming from package xyz, we are ok with this for now
CVN-111
CVN-123 # Mitigated the risk of this since we only use one method in this package and the affected code doesn't matter
CVN-543
```

It's also possible to define expiring ignores. Meaning that if you define a date on a vulnerability ignore until that date it will be ignored and once that
date is passed it will now be reported by nancy if it's still an issue. Format to add an expiring ignore looks as follows. They can also be followed up by comments
to provide context as to why it's been ignored until that date.

```
CVN-111 until=2021-01-01
CVN-543 until=2018-02-12 #Waiting on release from third party. Should be out before this date but gives us a little time to fix it.
```

### Output

We support multiple different output formats. Examples can be found below for each. [This intentionally vulnerable repo](https://github.com/sonatype-nexus-community/intentionally-vulnerable-golang-project) was used to generate the example output.
Quiet option is supported in text and csv. json formatting will ignore the Quiet option and output the same values if it's passed or not.

_text (default)_

```
Nancy version: development
!!!!! WARNING !!!!!
Scanning cannot be completed on the following package(s) since they do not use semver.
[1/1]pkg:golang/github.com/go-gitea/gitea@1.3.0.rc1

------------------------------------------------------------
[1/10]pkg:golang/github.com/bitly/oauth2_proxy@0.1  [Vulnerable]   1 known vulnerabilities affecting installed version

[CVE-2017-1000070]  URL Redirection to Untrusted Site ("Open Redirect")
The Bitly oauth2_proxy in version 2.1 and earlier was affected by an open redirect vulnerability during the start and termination of the 2-legged OAuth flow. This issue was caused by improper input validation and a violation of RFC-6819

ID:9eb9a5bc-8310-4104-bf85-3a820d28ba79
Details:https://ossindex.sonatype.org/vuln/9eb9a5bc-8310-4104-bf85-3a820d28ba79
[2/10]pkg:golang/github.com/cockroachdb/cockroach@2.1.4   No known vulnerabilities against package/version
------------------------------------------------------------
[3/10]pkg:golang/github.com/ethereum/go-ethereum@1.8.15  [Vulnerable]   1 known vulnerabilities affecting installed version

CWE-400: Uncontrolled Resource Consumption ('Resource Exhaustion')
The software does not properly restrict the size or amount of resources that are requested or influenced by an actor, which can be used to consume more resources than intended.

...

Audited dependencies:10,Vulnerable:6
```

_json_

```json
{"audited":[{"Coordinates":"pkg:golang/github.com/bitly/oauth2_proxy@0.1","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/bitly/oauth2_proxy@0.1","Vulnerabilities":[{"Id":"9eb9a5bc-8310-4104-bf85-3a820d28ba79","Title":"[CVE-2017-1000070]  URL Redirection to Untrusted Site (\"Open Redirect\")","Description":"The Bitly oauth2_proxy in version 2.1 and earlier was affected by an open redirect vulnerability during the start and termination of the 2-legged OAuth flow. This issue was caused by improper input validation and a violation of RFC-6819","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"CVE-2017-1000070","Reference":"https://ossindex.sonatype.org/vuln/9eb9a5bc-8310-4104-bf85-3a820d28ba79","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/cockroachdb/cockroach@2.1.4","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/cockroachdb/cockroach@2.1.4","Vulnerabilities":[],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/ethereum/go-ethereum@1.8.15","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/ethereum/go-ethereum@1.8.15","Vulnerabilities":[{"Id":"4efaed86-e62e-4c0c-b812-36c07e61ede4","Title":"CWE-400: Uncontrolled Resource Consumption ('Resource Exhaustion')","Description":"The software does not properly restrict the size or amount of resources that are requested or influenced by an actor, which can be used to consume more resources than intended.","CvssScore":"7.5","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/4efaed86-e62e-4c0c-b812-36c07e61ede4","Excluded":false}],"InvalidSemVer":false},...}
```

_json-pretty_

```json
{
  "audited": [
    {
      "Coordinates": "pkg:golang/github.com/bitly/oauth2_proxy@0.1",
      "Reference": "https://ossindex.sonatype.org/component/pkg:golang/github.com/bitly/oauth2_proxy@0.1",
      "Vulnerabilities": [
        {
          "Id": "9eb9a5bc-8310-4104-bf85-3a820d28ba79",
          "Title": "[CVE-2017-1000070]  URL Redirection to Untrusted Site (\"Open Redirect\")",
          "Description": "The Bitly oauth2_proxy in version 2.1 and earlier was affected by an open redirect vulnerability during the start and termination of the 2-legged OAuth flow. This issue was caused by improper input validation and a violation of RFC-6819",
          "CvssScore": "6.1",
          "CvssVector": "CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N",
          "Cve": "CVE-2017-1000070",
          "Reference": "https://ossindex.sonatype.org/vuln/9eb9a5bc-8310-4104-bf85-3a820d28ba79",
          "Excluded": false
        }
      ],
      "InvalidSemVer": false
    },
    ...
  ],
  "exclusions": [],
  "invalid": [
    {
      "Coordinates": "pkg:golang/github.com/go-gitea/gitea@1.3.0.rc1",
      "Reference": "",
      "Vulnerabilities": null,
      "InvalidSemVer": true
    }
  ],
  "num_audited": 10,
  "num_vulnerable": 6,
  "version": "development",
  "vulnerable": [...]
}
```

_csv_

```csv
Summary
Audited Count,Vulnerable Count,Build Version
10,6,development

Invalid Package(s)
Count,Package,Reason
[1/1],pkg:golang/github.com/go-gitea/gitea@1.3.0.rc1,Does not use SemVer

Audited Package(s)
Count,Package,Is Vulnerable,Num Vulnerabilities,Vulnerabilities
[1/10],pkg:golang/github.com/bitly/oauth2_proxy@0.1,true,1,"[{""Id"":""9eb9a5bc-8310-4104-bf85-3a820d28ba79"",""Title"":""[CVE-2017-1000070]  URL Redirection to Untrusted Site (\""Open Redirect\"")"",""Description"":""The Bitly oauth2_proxy in version 2.1 and earlier was affected by an open redirect vulnerability during the start and termination of the 2-legged OAuth flow. This issue was caused by improper input validation and a violation of RFC-6819"",""CvssScore"":""6.1"",""CvssVector"":""CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N"",""Cve"":""CVE-2017-1000070"",""Reference"":""https://ossindex.sonatype.org/vuln/9eb9a5bc-8310-4104-bf85-3a820d28ba79"",""Excluded"":false}]"
[2/10],pkg:golang/github.com/cockroachdb/cockroach@2.1.4,false,0,[]
...
```

---

## Sonatype Lifecycle Options

By default, assuming you have an out-of-the-box Sonatype Lifecycle server running, you can run `nancy` like so:

`go list -json -deps ./... | nancy lifecycle --lifecycle-application public-application-id`

It is STRONGLY suggested that you do not do this, and we will warn you on output if you are.

A more logical use of `nancy` against Sonatype Lifecycle will look like so:

`go list -json -deps ./... | nancy lifecycle --lifecycle-application public-application-id --lifecycle-username nondefaultuser --lifecycle-token yourtoken --lifecycle-server-url http://adifferentserverurl:port --lifecycle-stage develop`

Options for stage are as follows:

`build, develop, stage-release, release`

By default `--lifecycle-stage` will be `develop`.

> **Note**: `nancy iq` is a deprecated alias for `nancy lifecycle`. The `--iq-*` flags are deprecated aliases for the `--lifecycle-*` flags. Both continue to work in v2.0.0 but will be removed in v3.x.

Successful submissions to Sonatype Lifecycle will result in either an OS exit of 0, meaning all is clear and a response akin to:

```
Wonderbar! No policy violations reported for this audit!
Report URL:  http://reportURL
```

Failed submissions will either indicate failure because of an issue with processing the request, or a policy violation. Both will exit with a code of 1, allowing you to fail your build in CI. Policy Violation failures will include a report URL where you can learn more about why you encountered a failure.

Policy violations will look like:

```
Hi, Nancy here, you have some policy violations to clean up!
Report URL:  http://reportURL
```

Errors processing in Sonatype Lifecycle will look like:

```
Uh oh! There was an error with your request to Sonatype Lifecycle: <error>
```

### Persistent Lifecycle Config

Nancy lets you set the Sonatype Lifecycle Address, User and Token as persistent config (application and stage are generally per project, so we do not let you set these globally).

To set your Lifecycle config run:

`nancy config`

Choose `lifecycle` as an option and run through the rest of the config. Once you are done, Nancy should use this config for communicating with Sonatype Lifecycle, simplifying your use of the tool.

You can also specify configuration values using environment variables:

```shell
export IQ_USERNAME=nondefaultuser
export IQ_TOKEN=yourtoken
export IQ_SERVER=http://adifferentserverurl:port
go list -json -deps ./... | nancy lifecycle --lifecycle-application public-application-id
```

---

## Usage in CI

You can see an example of using `nancy` in Travis-CI at [this intentionally vulnerable repo we made](https://github.com/sonatype-nexus-community/intentionally-vulnerable-golang-project).

Nancy as well runs on itself (delicious dog food!) in CircleCI, in a myriad of fashions. You can see how we do that here in [our repo's CircleCI config](https://github.com/sonatype-nexus-community/nancy/blob/main/.circleci/config.yml).

#### Big CI Note:

Nancy will automatically check for newer releases of Nancy, and will prompt you when updates are detected.
The automatic update check will only occur once every 28 hours, and the date stamp of the last update check is stored
in the file: `~/.ossindex/.nancy-config/update_check.yml`.

If you have a huge CI matrix build, and want to avoid all the builds performing the automatic update check, you may
want to configure your CI build to cache the above directory.

---

## Why Nancy?

[Nancy Drew](https://en.wikipedia.org/wiki/Nancy_Drew) was the first female detective used extensively in literature, and gave women across the world a new hero.

This project is called `nancy` as like the great detective herself, it looks for problems you might not be aware of, and gives you the information to help put them to an end!

### Relationship to govulncheck

Go community starting 1.18, has used a tool called `govulncheck` shipped with golang distribution to verify vulnerabilities. Govulncheck reports known vulnerabilities using static analysis of source code or a binary's symbol table.
Nancy uses Sonatype's vulnerability database. Nancy inspects dependency files to look at all possible vulnerable library usage.

---

## DISCLAIMER

A portion of the golang ecosystem doesn't use proper versions, and instead uses a commit hash to resolve your dependency. Dependencies like this will not work with
`nancy` quite yet, as we don't have a mechanism to lookup vulnerabilities in that manner.

---

## How to Fix Vulnerabilities

So you've found a vulnerability. Now what? The best case is to upgrade the vulnerable component to a newer/non-vulnerable
version. However, it is likely the vulnerable component is not a direct dependency, but instead is a transitive dependency
(a dependency of a dependency, of a dependency, wash-rinse-repeat). In such a case, the first step is to figure out which
direct dependency (and sub-dependencies) depend on the vulnerable component.

The command `go mod graph | grep my/vulnerable` will show which module(s) pulls in the `my/vulnerable` package.

As an example, suppose we've learned that component `github.com/gogo/protobuf`, version 1.2.1 is vulnerable (CVE-2021-3121).
Use the following command to determine which components depend on `github.com/gogo/protobuf`.
```shell
$ go mod graph | grep github.com/gogo/protobuf
github.com/gogo/protobuf@v1.2.1 github.com/kisielk/errcheck@v1.1.0
github.com/spf13/viper@v1.4.0 github.com/gogo/protobuf@v1.2.1
github.com/prometheus/common@v0.4.0 github.com/gogo/protobuf@v1.1.1
github.com/prometheus/tsdb@v0.7.1 github.com/gogo/protobuf@v1.1.1
github.com/spf13/viper@v1.7.1 github.com/gogo/protobuf@v1.2.1
```

There are a number of approaches to resolving the vulnerability, but no matter which approach you choose, you should
probably make sure all the tests are passing before making any dependency changes.
<details>
  <summary>Click to expand output of command:

```shell
$ go test ./...
```
  </summary>

```shell
$ go test ./...
?       github.com/sonatype-nexus-community/nancy       [no test files]
ok      github.com/sonatype-nexus-community/nancy/buildversion  (cached)
ok      github.com/sonatype-nexus-community/nancy/internal/audit        (cached)
ok      github.com/sonatype-nexus-community/nancy/internal/cmd  0.206s
ok      github.com/sonatype-nexus-community/nancy/internal/customerrors (cached)
?       github.com/sonatype-nexus-community/nancy/internal/logger       [no test files]
ok      github.com/sonatype-nexus-community/nancy/packages      (cached)
ok      github.com/sonatype-nexus-community/nancy/parse (cached)
?       github.com/sonatype-nexus-community/nancy/settings      [no test files]
ok      github.com/sonatype-nexus-community/nancy/types (cached)
ok      github.com/sonatype-nexus-community/nancy/update        (cached)
```
</details>

We now know the vulnerable component is pulled in by `github.com/spf13/viper@v1.7.1` (among others). Ideally, we could
upgrade the direct dependency (`github.com/spf13/viper`) to a version that does not depend on a vulnerable version of
the transitive dependency (`github.com/gogo/protobuf`).

In some cases, no such upgrade of the direct dependency exists that avoids a dependence on the vulnerable component.
In such a case, the next step is to file an issue with the direct dependency project for them to update the vulnerable
sub-dependencies. Be sure to read and follow any vulnerability reporting instructions published by the project: Look for
a `SECURITY.md` file, or other instructions on how to report vulnerabilities. Some projects may prefer you not report
the vulnerability publicly. Here's an example of such a bug report: [Issue #1066](https://github.com/spf13/viper/pull/1066)

#### Avoid use of `replace` command to permit use of new `go install` command.

  * The section below describing the use of the `replace` directive is no longer ideal due to changes in how the
    `go install` command behaves with projects containing `replace` directives.
     See [Deprecation of 'go get' for installing executables](https://go.dev/doc/go-get-install-deprecation).

     Here's an example of the issue:
     [cmd/go: go install cmd@version errors out when module with main package has replace directive](https://github.com/golang/go/issues/44840)


  * Instead of `replace`, you can update the `// indirect` dependency version to a non-vulnerable version. e.g.: In the second
    `require` stanza of `go.mod` where all the `indirect` dependencies are listed, update the dependency version:

        require (
            ... <first require stanza - direct dependencies listed here>
        )

        require (
            ... <second require stanza - indirect dependencies listed here>
            // fix vulnerability: CVE-2021-38561 in golang.org/x/text v0.3.5
            golang.org/x/text v0.3.7 // indirect
            ...
        )

(*Deprecated* see above) Until the direct dependency is updated, the next best solution is to use a `replace` directive in the `go.mod` file
to use a newer version of the transitive dependency.
See [replace directive](https://golang.org/ref/mod#go-mod-file-replace).

To avoid semver issues, you probably want to use a newer dependency version that is in the same "major.minor" version
as the vulnerable dependency version.

(*Deprecated* see above) You can add the following `replace` directive to your `go.mod` file to us a newer version of
`github.com/gogo/protobuf`:

```
// fix vulnerability: CVE-2021-3121 in github.com/gogo/protobuf v1.2.1
replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
```

Be aware that even after you add a `replace` directive, `go mod graph` will still show the old dependency version.
You can verify the new version is actually used via the `go list` command:
```shell
$ go mod tidy -compat=1.17
$ go list -deps | grep github.com/gogo/protobuf
github.com/gogo/protobuf v1.2.1 => github.com/gogo/protobuf v1.3.2
```
You can see the v1.2.1 is replaced with v1.3.2.

Finally, you may want to submit a PR to the project with the vulnerable dependency (to fix the issues you reported
earlier) in a new release of the direct dependency. Even better, also tell them about `nancy` and maybe they will add
`nancy` to their own CI system.

Yet another resolution, if no other options make sense, is to knowingly ignore the vulnerability. This may be the best
option if you know the application does not use the vulnerable code path and no upgraded/non-vulnerable versions are
available. See: [Exclude vulnerabilities](#exclude-vulnerabilities)

---

## Development

`nancy` is written using Go (see `go.mod` for the minimum required version).

Tests can be run like this `make test`

Adding new files? Get the license header correct with:

> go get -u github.com/google/addlicense
> addlicense -v -f ./header.txt .

### Release Process

Follow the steps below to release a new version of Nancy. You need push access to the repository.

1. Checkout/pull the latest `main` branch, and create a new tag with the desired semantic version and a helpful note:

   ```shell
   $ git tag -a v1.0.x -m "Helpful message in tag"
   ```

2. Push the tag up:

   ```shell
   $ git push origin v1.0.x
   ```

3. There is no step 3.

---

## Contributing

We care a lot about making the world a safer place, and that's why we created `nancy`. If you as well want to
speed up the pace of software development by working on this project, jump on in! Before you start work, create
a new issue, or comment on an existing issue, to let others know you are!

---

## Acknowledgements

The `nancy` logo was created using a combo of [Gopherize.me](https://gopherize.me/) and good ole Photoshop. Thanks to the creators of
Gopherize for an easy way to make a fun Gopher :)

Original Gopher designed by Renee French.

---

## The Fine Print

Remember:

This project is part of the [Sonatype Nexus Community](https://github.com/sonatype-nexus-community) organization, which is not officially supported by Sonatype. Please review the latest pull requests, issues, and commits to understand this project's readiness for contribution and use.

-   File suggestions and requests on this repo through GitHub Issues, so that the community can pitch in
-   Use or contribute to this project according to your organization's policies and your own risk tolerance
-   Don't file Sonatype support tickets related to this project— it won't reach the right people that way

Last but not least of all - have fun!

<!-- Links Section -->

[shield_gh-workflow-test]: https://img.shields.io/github/actions/workflow/status/sonatype-nexus-community/nancy/build.yaml?branch=main&logo=GitHub&logoColor=white 'build'
[shield_license]: https://img.shields.io/github/license/sonatype-nexus-community/nancy?logo=open%20source%20initiative&logoColor=white 'license'
[link_gh-workflow-test]: https://github.com/sonatype-nexus-community/nancy/actions/workflows/build.yaml?query=branch%3Amain
[license_file]: https://github.com/sonatype-nexus-community/nancy/blob/main/LICENSE
