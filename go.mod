module gitlab.eng.vmware.com/dsmithuchida/astrolabe-demo

go 1.14

require (
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmware-tanzu/astrolabe v0.3.0
)

replace astrolabe => ../../vmware-tanzu/astrolabe
