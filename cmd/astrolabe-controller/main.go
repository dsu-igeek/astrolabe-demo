/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	astrolabeiov1 "github.com/dsu-igeek/astrolabe-demo/cmd/astrolabe-controller/api/v1"
	"github.com/dsu-igeek/astrolabe-demo/cmd/astrolabe-controller/controllers"
	"github.com/dsu-igeek/astrolabe-demo/pkg/psql"
	//+kubebuilder:scaffold:imports

	kubernetes "github.com/vmware-tanzu/astrolabe-velero/pkg/k8sns"
	"github.com/vmware-tanzu/astrolabe/pkg/fs"
	"github.com/vmware-tanzu/astrolabe/pkg/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(astrolabeiov1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func configurePEM(confDir string) (astrolabe.ProtectedEntityManager, error) {
	addOnInits := make(map[string]server.InitFunc)
	addOnInits["k8sns"] = kubernetes.NewKubernetesNamespaceProtectedEntityTypeManagerFromConfig
	addOnInits["fss"] = fs.NewFSProtectedEntityTypeManagerFromConfig
	addOnInits["psql"] = psql.NewPSQLProtectedEntityTypeManager

	pem := server.NewProtectedEntityManager(confDir, addOnInits, logrus.New())
	return pem, nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var confDir string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&confDir, "confDir", "", "Configuration directory")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "d78b9fb1.astrolabe.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	pem, err := configurePEM(confDir)
	if err != nil {
		setupLog.Error(err, "unable to configure Protected Entity Manager")
		os.Exit(1)
	}

	if err = (&controllers.PSQLSnapshotReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Pem:    pem,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FileSystemSnapshot")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
