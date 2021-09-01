module github.com/dsu-igeek/astrolabe-demo/cmd/astrolabe

go 1.14

require (
	github.com/aws/aws-sdk-go v1.36.3
	github.com/dsu-igeek/astrolabe-demo v0.0.0-00010101000000-000000000000
	github.com/dsu-igeek/astrolabe-kopia v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli/v2 v2.2.0
	github.com/vmware-tanzu/astrolabe v0.8.1
	github.com/vmware-tanzu/astrolabe-velero v0.0.0-00010101000000-000000000000
	github.com/vmware-tanzu/velero-plugin-for-aws v0.0.0-00010101000000-000000000000
	github.com/vmware-tanzu/velero-plugin-for-vsphere v1.2.1 // indirect
)

replace github.com/vmware-tanzu/astrolabe => ../../../../vmware-tanzu/astrolabe

replace github.com/vmware-tanzu/velero => ../../../../vmware-tanzu/velero

replace github.com/dsu-igeek/astrolabe-kopia => ../../../astrolabe-kopia

replace github.com/kopia/kopia => ../../../../kopia/kopia

replace github.com/dsu-igeek/astrolabe-demo => ../..

replace github.com/vmware-tanzu/velero-plugin-for-aws => ../../../../vmware-tanzu/velero-plugin-for-aws

replace github.com/vmware-tanzu/velero-plugin-for-vsphere => ../../../../vmware-tanzu/velero-plugin-for-vsphere

replace github.com/vmware-tanzu/astrolabe-velero => ../../../../vmware-tanzu/astrolabe-velero
