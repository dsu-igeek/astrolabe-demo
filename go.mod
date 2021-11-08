module github.com/dsu-igeek/astrolabe-demo

go 1.14

require (
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/vmware-tanzu/astrolabe v0.4.1-0.20210813185044-12eb18c3f6d5
	github.com/vmware-tanzu/astrolabe-velero v0.0.0-00010101000000-000000000000
	github.com/vmware-tanzu/velero-plugin-for-vsphere v1.3.0 // indirect
	github.com/zalando/postgres-operator v1.6.3
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.2
)

replace github.com/vmware-tanzu/astrolabe => ../../vmware-tanzu/astrolabe

replace github.com/vmware-tanzu/velero => ../../vmware-tanzu/velero

replace github.com/vmware-tanzu/astrolabe-velero => ../../vmware-tanzu/astrolabe-velero
