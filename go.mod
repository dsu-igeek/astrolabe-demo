module gitlab.eng.vmware.com/dsmithuchida/astrolabe-demo

go 1.14

require (
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/vmware-tanzu/astrolabe v0.3.0
	github.com/zalando/postgres-operator v1.6.3
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)

replace github.com/vmware-tanzu/astrolabe => ../../vmware-tanzu/astrolabe
