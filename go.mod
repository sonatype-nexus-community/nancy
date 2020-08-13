module github.com/sonatype-nexus-community/nancy

require (
	github.com/Flaque/filet v0.0.0-20190209224823-fc4d33cfcf93
	github.com/Masterminds/semver v0.0.0-20180403130225-3c92f33da7a8
	github.com/Masterminds/vcs v1.13.1 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20190529165535-67e0ed34491a
	github.com/golang/dep v0.5.4
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/jedib0t/go-pretty/v6 v6.0.2
	github.com/jmank88/nuts v0.3.0 // indirect
	github.com/logrusorgru/aurora v0.0.0-20190803045625-94edacc10f9b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nightlyone/lockfile v0.0.0-20180618180623-0ad87eef1443 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/sdboyer/constext v0.0.0-20170321163424-836a14457353 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/sonatype-nexus-community/go-sona-types v0.0.6
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.5.1
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a // indirect
	gopkg.in/yaml.v2 v2.3.0
)

// fix vulnerability: sonatype-2019-0666 in v1.4.0
replace github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2

// fix vulnerability: CVE-2019-11840 in v0.0.0-20190308221718-c2843e01d9a2
replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9

// fix vulnerability: CVE-2020-14040 in v0.3.0
replace golang.org/x/text => golang.org/x/text v0.3.3

go 1.13
