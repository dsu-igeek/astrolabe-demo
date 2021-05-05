module github.com/dsu-igeek/astrolabe-demo/cmd/astrolabe

go 1.14

require (
	github.com/aws/aws-sdk-go v1.36.3
	github.com/dsu-igeek/astrolabe-kopia v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/urfave/cli/v2 v2.2.0
	github.com/vmware-tanzu/astrolabe v0.8.1
)

replace github.com/vmware-tanzu/astrolabe => ../../../../vmware-tanzu/astrolabe

replace github.com/vmware-tanzu/velero => ../../../../vmware-tanzu/velero

replace github.com/dsu-igeek/astrolabe-kopia => ../../../astrolabe-kopia

replace github.com/kopia/kopia => ../../../../kopia/kopia