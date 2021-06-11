module github.com/sonatype-nexus-community/nancy

require (
	github.com/Flaque/filet v0.0.0-20190209224823-fc4d33cfcf93
	github.com/Masterminds/semver v0.0.0-20190925130524-317e8cce5480
	github.com/Masterminds/vcs v1.13.1 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20200609044655-c4b36f998cf2
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/dep v0.5.4
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/jedib0t/go-pretty/v6 v6.0.4
	github.com/jmank88/nuts v0.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/nightlyone/lockfile v1.0.0 // indirect
	github.com/owenrumney/go-sarif v1.0.4
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rhysd/go-github-selfupdate v1.2.3
	github.com/sdboyer/constext v0.0.0-20170321163424-836a14457353 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/sonatype-nexus-community/go-sona-types v0.0.12
	github.com/spf13/afero v1.3.4 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/ini.v1 v1.60.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

// fix vulnerability: CVE-2020-15114 in etcd v3.3.13+incompatible
replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

// fix vulnerability: CVE-2021-3121 in github.com/gogo/protobuf v1.2.1
replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

// fix vulnerability: SONATYPE-2019-0890 in github.com/pkg/sftp v1.10.1
replace github.com/pkg/sftp => github.com/pkg/sftp v1.13.0

go 1.13
