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
    <img src="https://github.com/sonatype-nexus-community/nancy/blob/main/docs/images/nancy.png" width="350" alt="nancy logo"/>
</p>

<p align="center">
    <a href="https://circleci.com/gh/sonatype-nexus-community/nancy"><img src="https://circleci.com/gh/sonatype-nexus-community/nancy.svg?style=shield" alt="Circle CI Build Status"/></a>
    <a href="https://gitter.im/sonatype-nexus-community/nancy?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge"><img src="https://badges.gitter.im/sonatype-nexus-community/nancy.svg" alt="Gitter"/></a>
</p>

# Nancy

`nancy` is a tool to check for vulnerabilities in your Golang dependencies, powered by [Sonatype OSS Index](https://ossindex.sonatype.org/), and as well, works with Nexus IQ Server, allowing you a smooth experience as a Golang developer, using the best tools in the market!

### Usage

`nancy` currently works for projects that use `dep` or `go mod` for dependencies.

```
 ~ > nancy --help
nancy is a tool to check for vulnerabilities in your Golang dependencies,
powered by the 'Sonatype OSS Index', and as well, works with Nexus IQ Server, allowing you
a smooth experience as a Golang developer, using the best tools in the market!

Usage:
  nancy [flags]
  nancy [command]

Examples:
  Typical usage will pipe the output of 'go list -json -m all' to 'nancy':
  go list -json -m all | nancy sleuth [flags]
  go list -json -m all | nancy iq [flags]

  If using dep typical usage is as follows :
  nancy sleuth -p Gopkg.lock [flags]
  nancy iq -p Gopkg.lock [flags]


Available Commands:
  config      Setup credentials to use when connecting to services
  help        Help about any command
  iq          Check for vulnerabilities in your Golang dependencies using 'Sonatype's Nexus IQ IQServer'
  sleuth      Check for vulnerabilities in your Golang dependencies using Sonatype's OSS Index
  update      Check if there are any updates available

Flags:
  -v, -- count            Set log level, multiple v's is more verbose
  -c, --clean-cache       Deletes local cache directory
  -h, --help              help for nancy
      --loud              indicate output should include non-vulnerable packages
  -p, --path string       Specify a path to a dep Gopkg.lock file for scanning
  -q, --quiet             indicate output should contain only packages with vulnerabilities (default true)
  -t, --token string      Specify OSS Index API token for request
  -u, --username string   Specify OSS Index username for request
  -V, --version           Get the version

Use "nancy [command] --help" for more information about a command.


$ > nancy sleuth --help
'nancy sleuth' is a command to check for vulnerabilities in your Golang dependencies, powered by the 'Sonatype OSS Index'.

Usage:
  nancy sleuth [flags]

Examples:
  go list -json -m all | nancy sleuth --username your_user --token your_token
  nancy sleuth -p Gopkg.lock --username your_user --token your_token

Flags:
  -e, --exclude-vulnerability CveListFlag   Comma separated list of CVEs or OSS Index IDs to exclude (default [])
  -x, --exclude-vulnerability-file string   Path to a file containing newline separated CVEs or OSS Index IDs to be excluded (default "./.nancy-ignore")
  -h, --help                                help for sleuth
  -n, --no-color                            indicate output should not be colorized
  -o, --output string                       Styling for output format. json, json-pretty, text, csv (default "text")

Global Flags:
  -v, -- count            Set log level, multiple v's is more verbose
      --loud              indicate output should include non-vulnerable packages
  -p, --path string       Specify a path to a dep Gopkg.lock file for scanning
  -q, --quiet             indicate output should contain only packages with vulnerabilities (default true)
  -t, --token string      Specify OSS Index API token for request
  -u, --username string   Specify OSS Index username for request
  -V, --version           Get the version


$ > nancy iq --help
'nancy iq' is a command to check for vulnerabilities in your Golang dependencies, powered by 'Sonatype's Nexus IQ IQServer', allowing you a smooth experience as a Golang developer, using the best tools in the market!

Usage:
  nancy iq [flags]

Examples:
  go list -json -m all | nancy iq --iq-application your_public_application_id --iq-server-url http://your_iq_server_url:port --iq-username your_user --iq-token your_token --iq-stage develop
  nancy iq -p Gopkg.lock --iq-application your_public_application_id --iq-server-url http://your_iq_server_url:port --iq-username your_user --iq-token your_token --iq-stage develop

Flags:
  -h, --help                    help for iq
  -a, --iq-application string   Specify Nexus IQ public application ID for request
  -x, --iq-server-url string    Specify Nexus IQ server url for request (default "http://localhost:8070")
  -s, --iq-stage string         Specify Nexus IQ stage for request (default "develop")
  -k, --iq-token string         Specify Nexus IQ token for request (default "admin123")
  -l, --iq-username string      Specify Nexus IQ username for request (default "admin")

Global Flags:
  -v, -- count            Set log level, multiple v's is more verbose
      --loud              indicate output should include non-vulnerable packages
  -p, --path string       Specify a path to a dep Gopkg.lock file for scanning
  -q, --quiet             indicate output should contain only packages with vulnerabilities (default true)
  -t, --token string      Specify OSS Index API token for request
  -u, --username string   Specify OSS Index username for request
  -V, --version           Get the version
```

#### What is the best usage of Nancy?

The preferred way to use Nancy is:
- `go list -json -m all | nancy sleuth`
- `nancy sleuth -p /path/to/Gopkg.lock`

#### Homebrew usage

`nancy` can be installed using `brew`:

- `brew tap sonatype-nexus-community/homebrew-nancy-tap`
- `brew install nancy`

`brew` formulae are created and published to that tap with each new release, so you can use `brew` to upgrade, etc... as you wish.

You can see more about the formulae, etc... at [this repo](https://github.com/sonatype-nexus-community/homebrew-nancy-tap).

#### Docker usage

<p align="center">
    <img src="https://github.com/sonatype-nexus-community/nancy/blob/main/docs/images/nancy_docker.png" width="350" alt="nancy docker logo"/>
</p>

`nancy` now comes in a boat! For ease of use, we've dockerized `nancy`. To use our Dockerfile:

`go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth`

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

### OSS Index Options

#### Rate limiting / Setting OSS Index config

**NOTE: New as of Nancy v0.1.17**

If you start using Nancy extensively, you might run into Rate Limiting from OSS Index! Don't worry, we've got your back!

If you run into Rate Limiting you should receive an error that will give you instructions on how to register on OSS Index:

```
You have been rate limited by OSS Index.
If you do not have a OSS Index account, please visit https://ossindex.sonatype.org/user/register to register an account.
After registering and verifying your account, you can retrieve your username (Email Address), and API Token
at https://ossindex.sonatype.org/user/settings. Upon retrieving those, run 'nancy config', set your OSS Index
settings, and rerun Nancy.
```

After setting this config, you'll be gifted a nice new higher rate limit. If you escape this limit, you might take a look at using Nexus IQ Server, or reach out to the friendly people at OSS Index for partnership opportunities.

You can also set the user and token via the command line like so:

`nancy sleuth --username auser@anemailaddress.com --token A4@k3@p1T0k3n`

This can be handy for testing your account out, or if you want to override your set config with a different user.

#### Loud mode

By default, `nancy` runs in a "quiet" mode, only displaying a list of vulnerable components. 
You can run `nancy` in a loud manner, showing all components by running:

* `nancy sleuth --loud -p /path/to/your/Gopkg.lock`
* `go list -json -m all | nancy sleuth --loud`

#### Exclude vulnerabilities

Sometimes you'll run into a dependency that after taking a look at, you either aren't affected by, or cannot resolve for some reason. Nancy understands, and will let you 
exclude these vulnerabilities so you can get back to a passing build:

Vulnerabilities excluded will then be silenced and not show up in the output or fail your build.

We support exclusion of vulnerability either by CVE-ID (ex: `CVE-2018-20303`) or via the OSS Index ID (ex: `a8c20c84-1f6a-472a-ba1b-3eaedb2a2a14`) as not all vulnerabilities have a CVE-ID.

##### Via CLI flag
* `nancy sleuth --exclude-vulnerability CVE-789,bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2 -p /path/to/your/Gopkg.lock`
* `go list -json -m all | nancy sleuth --exclude-vulnerability CVE-789,bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2`

##### Via file
By default if a file named `.nancy-ignore` exists in the same directory that nancy is run it will use it, will no other options need to be passed.

If you would like to define the path to the file you can use the following
* `nancy sleuth --exclude-vulnerability-file=/path/to/your/exclude-file -p /path/to/your/Gopkg.lock`
* `go list -json -m all | nancy sleuth --exclude-vulnerability-file=/path/to/your/exclude-file`

The file format requires each vulnerability that you want to exclude to be on a separate line. Comments are allowed in the file as well to help provide context when needed. See an example file below.

```
# This vulnerability is coming from package xyz, we are ok with this for now
CVN-111 
CVN-123 # Mitigated the risk of this since we only use one method in this package and the affected code doesn't matter
CVN-543
``` 
It's also possible to define expiring ignores. Meaning that if you define a date on a vulnerability ignore until that date it will be ignored and once that 
date is passed it will now be reported by nancy if its still an issue. Format to add an expiring ignore looks as follows. They can also be followed up by comments 
to provide context to as why its been ignored until that date.    

```
CVN-111 until=2021-01-01
CVN-543 until=2018-02-12 #Waiting on release from third party. Should be out before this date but gives us a little time to fix it. 
```

#### Output

We support multiple different output formats. Examples can be found below for each. [This intentionally vulnerable repo](https://github.com/sonatype-nexus-community/intentionally-vulnerable-golang-project) was used to generate the example output.
Quiet option is supported in text and csv. json formatting will ignore the Quiet option and output the same values if it's passed or not.  

*text (default)*
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

*json*
```json
{"audited":[{"Coordinates":"pkg:golang/github.com/bitly/oauth2_proxy@0.1","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/bitly/oauth2_proxy@0.1","Vulnerabilities":[{"Id":"9eb9a5bc-8310-4104-bf85-3a820d28ba79","Title":"[CVE-2017-1000070]  URL Redirection to Untrusted Site (\"Open Redirect\")","Description":"The Bitly oauth2_proxy in version 2.1 and earlier was affected by an open redirect vulnerability during the start and termination of the 2-legged OAuth flow. This issue was caused by improper input validation and a violation of RFC-6819","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"CVE-2017-1000070","Reference":"https://ossindex.sonatype.org/vuln/9eb9a5bc-8310-4104-bf85-3a820d28ba79","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/cockroachdb/cockroach@2.1.4","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/cockroachdb/cockroach@2.1.4","Vulnerabilities":[],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/ethereum/go-ethereum@1.8.15","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/ethereum/go-ethereum@1.8.15","Vulnerabilities":[{"Id":"4efaed86-e62e-4c0c-b812-36c07e61ede4","Title":"CWE-400: Uncontrolled Resource Consumption ('Resource Exhaustion')","Description":"The software does not properly restrict the size or amount of resources that are requested or influenced by an actor, which can be used to consume more resources than intended.","CvssScore":"7.5","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/4efaed86-e62e-4c0c-b812-36c07e61ede4","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/elastic/beats@5.6.3","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/elastic/beats@5.6.3","Vulnerabilities":[{"Id":"8e4d562d-517b-4d00-a845-a7a3e2be41db","Title":"[CVE-2017-11480]  Improper Access Control","Description":"Packetbeat versions prior to 5.6.4 are affected by a denial of service flaw in the PostgreSQL protocol handler. If Packetbeat is listening for PostgreSQL traffic and a user is able to send arbitrary network traffic to the monitored port, the attacker could prevent Packetbeat from properly logging other PostgreSQL traffic.","CvssScore":"7.5","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H","Cve":"CVE-2017-11480","Reference":"https://ossindex.sonatype.org/vuln/8e4d562d-517b-4d00-a845-a7a3e2be41db","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/etcd-io/etcd@3.3.0","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/etcd-io/etcd@3.3.0","Vulnerabilities":[{"Id":"5c876f5e-2814-4822-baf0-1092fc63ec25","Title":"[CVE-2018-1098]  Cross-Site Request Forgery (CSRF)","Description":"A cross-site request forgery flaw was found in etcd 3.3.1 and earlier. An attacker can set up a website that tries to send a POST request to the etcd server and modify a key. Adding a key is done with PUT so it is theoretically safe (can't PUT from an HTML form or such) but POST allows creating in-order keys that an attacker can send.","CvssScore":"8.8","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H","Cve":"CVE-2018-1098","Reference":"https://ossindex.sonatype.org/vuln/5c876f5e-2814-4822-baf0-1092fc63ec25","Excluded":false},{"Id":"8a190129-526c-4ee0-b663-92f38139c165","Title":"[CVE-2018-1099]  Improper Input Validation","Description":"DNS rebinding vulnerability found in etcd 3.3.1 and earlier. An attacker can control his DNS records to direct to localhost, and trick the browser into sending requests to localhost (or any other address).","CvssScore":"5.5","CvssVector":"CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:H/A:N","Cve":"CVE-2018-1099","Reference":"https://ossindex.sonatype.org/vuln/8a190129-526c-4ee0-b663-92f38139c165","Excluded":false},{"Id":"69b9f08b-8eda-4125-8e84-b7d67a7c9ee5","Title":"[CVE-2018-16886]  Improper Authentication","Description":"etcd versions 3.2.x before 3.2.26 and 3.3.x before 3.3.11 are vulnerable to an improper authentication issue when role-based access control (RBAC) is used and client-cert-auth is enabled. If an etcd client server TLS certificate contains a Common Name (CN) which matches a valid RBAC username, a remote attacker may authenticate as that user with any valid (trusted) client certificate in a REST API request to the gRPC-gateway.","CvssScore":"8.1","CvssVector":"CVSS:3.0/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:H","Cve":"CVE-2018-16886","Reference":"https://ossindex.sonatype.org/vuln/69b9f08b-8eda-4125-8e84-b7d67a7c9ee5","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/github/hub@2.0.0","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/github/hub@2.0.0","Vulnerabilities":[],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/gogs/gogs@0.9.45","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/gogs/gogs@0.9.45","Vulnerabilities":[{"Id":"a4c682fa-9c9f-4e9e-b218-720d5125b17f","Title":"CWE-89: Improper Neutralization of Special Elements used in an SQL Command ('SQL Injection')","Description":"The software constructs all or part of an SQL command using externally-influenced input from an upstream component, but it does not neutralize or incorrectly neutralizes special elements that could modify the intended SQL command when it is sent to a downstream component.","CvssScore":"9.9","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:L/UI:N/S:C/C:H/I:H/A:H","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/a4c682fa-9c9f-4e9e-b218-720d5125b17f","Excluded":false},{"Id":"304fa9e0-012e-4385-88b2-88c0c5ec3247","Title":"[CVE-2018-15192] An SSRF vulnerability in webhooks in Gitea through 1.5.0-rc2 and Gogs through 0....","Description":"An SSRF vulnerability in webhooks in Gitea through 1.5.0-rc2 and Gogs through 0.11.53 allows remote attackers to access intranet services.","CvssScore":"8.6","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:N/A:N","Cve":"CVE-2018-15192","Reference":"https://ossindex.sonatype.org/vuln/304fa9e0-012e-4385-88b2-88c0c5ec3247","Excluded":false},{"Id":"a8c20c84-1f6a-472a-ba1b-3eaedb2a2a14","Title":"[CVE-2018-20303]  Improper Limitation of a Pathname to a Restricted Directory (\"Path Traversal\")","Description":"In pkg/tool/path.go in Gogs before 0.11.82.1218, a directory traversal in the file-upload functionality can allow an attacker to create a file under data/sessions on the server, a similar issue to CVE-2018-18925.","CvssScore":"7.5","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:H/A:N","Cve":"CVE-2018-20303","Reference":"https://ossindex.sonatype.org/vuln/a8c20c84-1f6a-472a-ba1b-3eaedb2a2a14","Excluded":false},{"Id":"bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2","Title":"[CVE-2018-18925] Gogs 0.11.66 allows remote code execution because it does not properly validate ...","Description":"Gogs 0.11.66 allows remote code execution because it does not properly validate session IDs, as demonstrated by a \"..\" session-file forgery in the file session provider in file.go. This is related to session ID handling in the go-macaron/session code for Macaron.","CvssScore":"9.8","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H","Cve":"CVE-2018-18925","Reference":"https://ossindex.sonatype.org/vuln/bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2","Excluded":false},{"Id":"bbbdbb94-f65a-475c-9e9f-6793778fbd9b","Title":"[CVE-2018-15178]  URL Redirection to Untrusted Site (\"Open Redirect\")","Description":"Open redirect vulnerability in Gogs before 0.12 allows remote attackers to redirect users to arbitrary websites and conduct phishing attacks via an initial /\\ substring in the user/login redirect_to parameter, related to the function isValidRedirect in routes/user/auth.go.","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"CVE-2018-15178","Reference":"https://ossindex.sonatype.org/vuln/bbbdbb94-f65a-475c-9e9f-6793778fbd9b","Excluded":false},{"Id":"fc70a115-52cc-44ea-a33d-793267f860dd","Title":"CWE-79: Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')","Description":"The software does not neutralize or incorrectly neutralizes user-controllable input before it is placed in output that is used as a web page that is served to other users.","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/fc70a115-52cc-44ea-a33d-793267f860dd","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/goharbor/harbor@1.7.2","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/goharbor/harbor@1.7.2","Vulnerabilities":[],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/gophish/gophish@0.1.1","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/gophish/gophish@0.1.1","Vulnerabilities":[{"Id":"0416e202-2705-431d-9915-8ed93334ca58","Title":"CWE-79: Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')","Description":"The software does not neutralize or incorrectly neutralizes user-controllable input before it is placed in output that is used as a web page that is served to other users.","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/0416e202-2705-431d-9915-8ed93334ca58","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/ipfs/go-ipfs@0.4.18","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/ipfs/go-ipfs@0.4.18","Vulnerabilities":[],"InvalidSemVer":false}],"exclusions":[],"invalid":[{"Coordinates":"pkg:golang/github.com/go-gitea/gitea@1.3.0.rc1","Reference":"","Vulnerabilities":null,"InvalidSemVer":true}],"num_audited":10,"num_vulnerable":6,"version":"development","vulnerable":[{"Coordinates":"pkg:golang/github.com/bitly/oauth2_proxy@0.1","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/bitly/oauth2_proxy@0.1","Vulnerabilities":[{"Id":"9eb9a5bc-8310-4104-bf85-3a820d28ba79","Title":"[CVE-2017-1000070]  URL Redirection to Untrusted Site (\"Open Redirect\")","Description":"The Bitly oauth2_proxy in version 2.1 and earlier was affected by an open redirect vulnerability during the start and termination of the 2-legged OAuth flow. This issue was caused by improper input validation and a violation of RFC-6819","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"CVE-2017-1000070","Reference":"https://ossindex.sonatype.org/vuln/9eb9a5bc-8310-4104-bf85-3a820d28ba79","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/ethereum/go-ethereum@1.8.15","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/ethereum/go-ethereum@1.8.15","Vulnerabilities":[{"Id":"4efaed86-e62e-4c0c-b812-36c07e61ede4","Title":"CWE-400: Uncontrolled Resource Consumption ('Resource Exhaustion')","Description":"The software does not properly restrict the size or amount of resources that are requested or influenced by an actor, which can be used to consume more resources than intended.","CvssScore":"7.5","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/4efaed86-e62e-4c0c-b812-36c07e61ede4","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/elastic/beats@5.6.3","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/elastic/beats@5.6.3","Vulnerabilities":[{"Id":"8e4d562d-517b-4d00-a845-a7a3e2be41db","Title":"[CVE-2017-11480]  Improper Access Control","Description":"Packetbeat versions prior to 5.6.4 are affected by a denial of service flaw in the PostgreSQL protocol handler. If Packetbeat is listening for PostgreSQL traffic and a user is able to send arbitrary network traffic to the monitored port, the attacker could prevent Packetbeat from properly logging other PostgreSQL traffic.","CvssScore":"7.5","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H","Cve":"CVE-2017-11480","Reference":"https://ossindex.sonatype.org/vuln/8e4d562d-517b-4d00-a845-a7a3e2be41db","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/etcd-io/etcd@3.3.0","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/etcd-io/etcd@3.3.0","Vulnerabilities":[{"Id":"5c876f5e-2814-4822-baf0-1092fc63ec25","Title":"[CVE-2018-1098]  Cross-Site Request Forgery (CSRF)","Description":"A cross-site request forgery flaw was found in etcd 3.3.1 and earlier. An attacker can set up a website that tries to send a POST request to the etcd server and modify a key. Adding a key is done with PUT so it is theoretically safe (can't PUT from an HTML form or such) but POST allows creating in-order keys that an attacker can send.","CvssScore":"8.8","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H","Cve":"CVE-2018-1098","Reference":"https://ossindex.sonatype.org/vuln/5c876f5e-2814-4822-baf0-1092fc63ec25","Excluded":false},{"Id":"8a190129-526c-4ee0-b663-92f38139c165","Title":"[CVE-2018-1099]  Improper Input Validation","Description":"DNS rebinding vulnerability found in etcd 3.3.1 and earlier. An attacker can control his DNS records to direct to localhost, and trick the browser into sending requests to localhost (or any other address).","CvssScore":"5.5","CvssVector":"CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:H/A:N","Cve":"CVE-2018-1099","Reference":"https://ossindex.sonatype.org/vuln/8a190129-526c-4ee0-b663-92f38139c165","Excluded":false},{"Id":"69b9f08b-8eda-4125-8e84-b7d67a7c9ee5","Title":"[CVE-2018-16886]  Improper Authentication","Description":"etcd versions 3.2.x before 3.2.26 and 3.3.x before 3.3.11 are vulnerable to an improper authentication issue when role-based access control (RBAC) is used and client-cert-auth is enabled. If an etcd client server TLS certificate contains a Common Name (CN) which matches a valid RBAC username, a remote attacker may authenticate as that user with any valid (trusted) client certificate in a REST API request to the gRPC-gateway.","CvssScore":"8.1","CvssVector":"CVSS:3.0/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:H","Cve":"CVE-2018-16886","Reference":"https://ossindex.sonatype.org/vuln/69b9f08b-8eda-4125-8e84-b7d67a7c9ee5","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/gogs/gogs@0.9.45","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/gogs/gogs@0.9.45","Vulnerabilities":[{"Id":"a4c682fa-9c9f-4e9e-b218-720d5125b17f","Title":"CWE-89: Improper Neutralization of Special Elements used in an SQL Command ('SQL Injection')","Description":"The software constructs all or part of an SQL command using externally-influenced input from an upstream component, but it does not neutralize or incorrectly neutralizes special elements that could modify the intended SQL command when it is sent to a downstream component.","CvssScore":"9.9","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:L/UI:N/S:C/C:H/I:H/A:H","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/a4c682fa-9c9f-4e9e-b218-720d5125b17f","Excluded":false},{"Id":"304fa9e0-012e-4385-88b2-88c0c5ec3247","Title":"[CVE-2018-15192] An SSRF vulnerability in webhooks in Gitea through 1.5.0-rc2 and Gogs through 0....","Description":"An SSRF vulnerability in webhooks in Gitea through 1.5.0-rc2 and Gogs through 0.11.53 allows remote attackers to access intranet services.","CvssScore":"8.6","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:N/A:N","Cve":"CVE-2018-15192","Reference":"https://ossindex.sonatype.org/vuln/304fa9e0-012e-4385-88b2-88c0c5ec3247","Excluded":false},{"Id":"a8c20c84-1f6a-472a-ba1b-3eaedb2a2a14","Title":"[CVE-2018-20303]  Improper Limitation of a Pathname to a Restricted Directory (\"Path Traversal\")","Description":"In pkg/tool/path.go in Gogs before 0.11.82.1218, a directory traversal in the file-upload functionality can allow an attacker to create a file under data/sessions on the server, a similar issue to CVE-2018-18925.","CvssScore":"7.5","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:H/A:N","Cve":"CVE-2018-20303","Reference":"https://ossindex.sonatype.org/vuln/a8c20c84-1f6a-472a-ba1b-3eaedb2a2a14","Excluded":false},{"Id":"bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2","Title":"[CVE-2018-18925] Gogs 0.11.66 allows remote code execution because it does not properly validate ...","Description":"Gogs 0.11.66 allows remote code execution because it does not properly validate session IDs, as demonstrated by a \"..\" session-file forgery in the file session provider in file.go. This is related to session ID handling in the go-macaron/session code for Macaron.","CvssScore":"9.8","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H","Cve":"CVE-2018-18925","Reference":"https://ossindex.sonatype.org/vuln/bcb0c38d-0d35-44ee-b7a7-8f77183d1ae2","Excluded":false},{"Id":"bbbdbb94-f65a-475c-9e9f-6793778fbd9b","Title":"[CVE-2018-15178]  URL Redirection to Untrusted Site (\"Open Redirect\")","Description":"Open redirect vulnerability in Gogs before 0.12 allows remote attackers to redirect users to arbitrary websites and conduct phishing attacks via an initial /\\ substring in the user/login redirect_to parameter, related to the function isValidRedirect in routes/user/auth.go.","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"CVE-2018-15178","Reference":"https://ossindex.sonatype.org/vuln/bbbdbb94-f65a-475c-9e9f-6793778fbd9b","Excluded":false},{"Id":"fc70a115-52cc-44ea-a33d-793267f860dd","Title":"CWE-79: Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')","Description":"The software does not neutralize or incorrectly neutralizes user-controllable input before it is placed in output that is used as a web page that is served to other users.","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/fc70a115-52cc-44ea-a33d-793267f860dd","Excluded":false}],"InvalidSemVer":false},{"Coordinates":"pkg:golang/github.com/gophish/gophish@0.1.1","Reference":"https://ossindex.sonatype.org/component/pkg:golang/github.com/gophish/gophish@0.1.1","Vulnerabilities":[{"Id":"0416e202-2705-431d-9915-8ed93334ca58","Title":"CWE-79: Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')","Description":"The software does not neutralize or incorrectly neutralizes user-controllable input before it is placed in output that is used as a web page that is served to other users.","CvssScore":"6.1","CvssVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N","Cve":"","Reference":"https://ossindex.sonatype.org/vuln/0416e202-2705-431d-9915-8ed93334ca58","Excluded":false}],"InvalidSemVer":false}]}
```

*json-pretty*
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
    {
      "Coordinates": "pkg:golang/github.com/cockroachdb/cockroach@2.1.4",
      "Reference": "https://ossindex.sonatype.org/component/pkg:golang/github.com/cockroachdb/cockroach@2.1.4",
      "Vulnerabilities": [],
      "InvalidSemVer": false
    },
    {
      "Coordinates": "pkg:golang/github.com/ethereum/go-ethereum@1.8.15",
      "Reference": "https://ossindex.sonatype.org/component/pkg:golang/github.com/ethereum/go-ethereum@1.8.15",
      "Vulnerabilities": [
        {
          "Id": "4efaed86-e62e-4c0c-b812-36c07e61ede4",
          "Title": "CWE-400: Uncontrolled Resource Consumption ('Resource Exhaustion')",
          "Description": "The software does not properly restrict the size or amount of resources that are requested or influenced by an actor, which can be used to consume more resources than intended.",
          "CvssScore": "7.5",
          "CvssVector": "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H",
          "Cve": "",
          "Reference": "https://ossindex.sonatype.org/vuln/4efaed86-e62e-4c0c-b812-36c07e61ede4",
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
  "vulnerable": [
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
    {
      "Coordinates": "pkg:golang/github.com/ethereum/go-ethereum@1.8.15",
      "Reference": "https://ossindex.sonatype.org/component/pkg:golang/github.com/ethereum/go-ethereum@1.8.15",
      "Vulnerabilities": [
        {
          "Id": "4efaed86-e62e-4c0c-b812-36c07e61ede4",
          "Title": "CWE-400: Uncontrolled Resource Consumption ('Resource Exhaustion')",
          "Description": "The software does not properly restrict the size or amount of resources that are requested or influenced by an actor, which can be used to consume more resources than intended.",
          "CvssScore": "7.5",
          "CvssVector": "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H",
          "Cve": "",
          "Reference": "https://ossindex.sonatype.org/vuln/4efaed86-e62e-4c0c-b812-36c07e61ede4",
          "Excluded": false
        }
      ],
      "InvalidSemVer": false
    },
    {
      "Coordinates": "pkg:golang/github.com/elastic/beats@5.6.3",
      "Reference": "https://ossindex.sonatype.org/component/pkg:golang/github.com/elastic/beats@5.6.3",
      "Vulnerabilities": [
        {
          "Id": "8e4d562d-517b-4d00-a845-a7a3e2be41db",
          "Title": "[CVE-2017-11480]  Improper Access Control",
          "Description": "Packetbeat versions prior to 5.6.4 are affected by a denial of service flaw in the PostgreSQL protocol handler. If Packetbeat is listening for PostgreSQL traffic and a user is able to send arbitrary network traffic to the monitored port, the attacker could prevent Packetbeat from properly logging other PostgreSQL traffic.",
          "CvssScore": "7.5",
          "CvssVector": "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H",
          "Cve": "CVE-2017-11480",
          "Reference": "https://ossindex.sonatype.org/vuln/8e4d562d-517b-4d00-a845-a7a3e2be41db",
          "Excluded": false
        }
      ],
      "InvalidSemVer": false
    },
    ...
  ]
}
```

*csv*
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
[3/10],pkg:golang/github.com/ethereum/go-ethereum@1.8.15,true,1,"[{""Id"":""4efaed86-e62e-4c0c-b812-36c07e61ede4"",""Title"":""CWE-400: Uncontrolled Resource Consumption ('Resource Exhaustion')"",""Description"":""The software does not properly restrict the size or amount of resources that are requested or influenced by an actor, which can be used to consume more resources than intended."",""CvssScore"":""7.5"",""CvssVector"":""CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H"",""Cve"":"""",""Reference"":""https://ossindex.sonatype.org/vuln/4efaed86-e62e-4c0c-b812-36c07e61ede4"",""Excluded"":false}]"
...
```
### Nexus IQ Server Options

By default, assuming you have an out of the box Nexus IQ Server running, you can run `nancy` like so:

`go list -json -m all | nancy iq --iq-application public-application-id`

It is STRONGLY suggested that you do not do this, and we will warn you on output if you are.

A more logical use of `nancy` against Nexus IQ Server will look like so:

`go list -json -m all | nancy iq --iq-application public-application-id --iq-username nondefaultuser --iq-token yourtoken --iq-server-url http://adifferentserverurl:port --iq-stage develop`

Options for stage are as follows:

`build, develop, stage-release, release`

By default `--iq-stage` will be `develop`.

Successful submissions to Nexus IQ Server will result in either an OS exit of 0, meaning all is clear and a response akin to:

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

Errors processing in Nexus IQ Server will look like:

```
Uh oh! There was an error with your request to Nexus IQ Server: <error>
```

#### Persistent Nexus IQ Server Config

Nancy lets you set the Nexus IQ Server Address, User and Token as persistent config (application and stage are generally per project so we do not let you set these globally).

To set your Nexus IQ Server config run:

`nancy config`

Choose `iq` as an option and run through the rest of the config. Once you are done, Nancy should use this config for communicating with Nexus IQ, simplifying your use of the tool.

### Usage in CI

You can see an example of using `nancy` in Travis-CI at [this intentionally vulnerable repo we made](https://github.com/sonatype-nexus-community/intentionally-vulnerable-golang-project).

Nancy as well runs on it self (delicious dog food!) in CircleCI, in a myriad of fashions. You can see how we do that here in [our repo's CircleCI config](https://github.com/sonatype-nexus-community/nancy/blob/main/.circleci/config.yml).

### DISCLAIMER

A portion of the golang ecosystem doesn't use proper versions, and instead uses a commit hash to resolve your dependency. Dependencies like this will not work with
`nancy` quite yet, as we don't have a mechanism on OSS Index to lookup vulnerabilities in that manner. 

## Why Nancy?

[Nancy Drew](https://en.wikipedia.org/wiki/Nancy_Drew) was the first female detective used extensively in literature, and gave women across the world a new hero.

This project is called `nancy` as like the great detective herself, it looks for problems you might not be aware of, and gives you the information to help put them to an end!

## Installation

At the current time you have a few options:

* Build from source
* Download release binary from [here on GitHub](https://github.com/sonatype-nexus-community/nancy/releases)

### Build from source

* Clone the project `git clone github.com/sonatype-nexus-community/nancy`
* In the root of the project run `make`
  * This will execute multiple targets so if you want to short circuit some of that process you can also just run `make build` to get the binary without running tests, linting, etc
* Use that binary where ever your heart so desires!

### Download release binary

Each tag pushed to this repo creates a new release binary, and if you'd like to skip building from source, you can download a binary similar to:

```console
$ curl -o /path/where/you/want/nancy \
  https://github.com/sonatype-nexus-community/nancy/releases/download/v0.0.44/nancy-darwin.amd64-v0.0.44
```

## Development

`nancy` is written using Golang 1.13, so it is best you start there.

Tests can be run like this `make test`

Adding new files? Get the license header correct with:

> go get -u github.com/google/addlicense
> addlicense -v -f ./header.txt .

### Release Process

Follow the steps below to release a new version of Nancy. You need to be part of the `deploy from circle ci` group for this to work.

  1. Checkout/pull the latest `main` branch, and create a new tag with the desired semantic version and a helpful note:
  
         git tag -a v1.0.x -m "Helpful message in tag."
         
  2. Push the tag up:
  
         git push origin v1.0.x
         
  3. There is no step 3.
          
## Contributing

We care a lot about making the world a safer place, and that's why we created `nancy`. If you as well want to
speed up the pace of software development by working on this project, jump on in! Before you start work, create
a new issue, or comment on an existing issue, to let others know you are!

## Acknowledgements

The `nancy` logo was created using a combo of [Gopherize.me](https://gopherize.me/) and good ole Photoshop. Thanks to the creators of 
Gopherize for an easy way to make a fun Gopher :)

Original Gopher designed by Renee French.

## The Fine Print

Remember:

* If you are a Sonatype customer, you may file Sonatype support tickets related to `nancy` support in regard to this project
  * We suggest you file issues here on GitHub as well, so that the community can pitch in
* If you are not a Sonatype customer, Do NOT file Sonatype support tickets related to nancy support in regard to this project, file an issue here on GitHub

Have fun creating and using `nancy` and the [Sonatype OSS Index](https://ossindex.sonatype.org/), we are glad to have you here!

## Getting help

Looking to contribute to our code but need some help? There's a few ways to get information:

* Chat with us on [Gitter](https://gitter.im/sonatype-nexus-community/nancy)
