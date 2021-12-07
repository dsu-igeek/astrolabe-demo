module gitlab.eng.vmware.com/dsmithuchida/astrolabe-demo

go 1.14

require (
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/vmware-tanzu/astrolabe v0.3.0
	github.com/vmware-tanzu/velero v0.0.0-00010101000000-000000000000 // indirect
	github.com/zalando/postgres-operator v1.6.0
	k8s.io/api v0.19.7 // indirect
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
)

replace github.com/vmware-tanzu/astrolabe => ../../vmware-tanzu/astrolabe

replace github.com/vmware-tanzu/velero => ../../vmware-tanzu/velero
