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
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	schedulerpluginsv1alpha1 "sigs.k8s.io/scheduler-plugins/apis/scheduling/v1alpha1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	volcanoclient "volcano.sh/apis/pkg/client/clientset/versioned"

	"github.com/kubeflow/common/pkg/controller.v1/common"
	commonutil "github.com/kubeflow/common/pkg/util"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/kubeflow/training-operator/pkg/config"
	controllerv1 "github.com/kubeflow/training-operator/pkg/controller.v1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kubeflowv1.AddToScheme(scheme))
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(schedulerpluginsv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var leaderElectionID string
	var probeAddr string
	var enabledSchemes controllerv1.EnabledSchemes
	var gangSchedulerName string
	var namespace string
	var monitoringPort int
	var controllerThreads int
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&leaderElectionID, "leader-election-id", "1ca428e5.training-operator.kubeflow.org", "The ID for leader election.")
	flag.Var(&enabledSchemes, "enable-scheme", "Enable scheme(s) as --enable-scheme=tfjob --enable-scheme=pytorchjob, case insensitive."+
		" Now supporting TFJob, PyTorchJob, MXNetJob, XGBoostJob, PaddleJob. By default, all supported schemes will be enabled.")
	flag.StringVar(&gangSchedulerName, "gang-scheduler-name", "none", "The scheduler to gang-schedule kubeflow jobs, defaults to none")
	flag.StringVar(&namespace, "namespace", os.Getenv(commonutil.EnvKubeflowNamespace), "The namespace to monitor kubeflow jobs. If unset, it monitors all namespaces cluster-wide."+
		"If set, it only monitors kubeflow jobs in the given namespace.")
	flag.IntVar(&monitoringPort, "monitoring-port", 9443, "Endpoint port for displaying monitoring metrics. "+
		"It can be set to \"0\" to disable the metrics serving.")
	flag.IntVar(&controllerThreads, "controller-threads", 1, "Number of worker threads used by the controller.")

	// PyTorch related flags
	flag.StringVar(&config.Config.PyTorchInitContainerImage, "pytorch-init-container-image",
		config.PyTorchInitContainerImageDefault, "The image for pytorch init container")
	flag.StringVar(&config.Config.PyTorchInitContainerTemplateFile, "pytorch-init-container-template-file",
		config.PyTorchInitContainerTemplateFileDefault, "The template file for pytorch init container")

	// MPI related flags
	flag.StringVar(&config.Config.MPIKubectlDeliveryImage, "mpi-kubectl-delivery-image",
		config.MPIKubectlDeliveryImageDefault, "The image for mpi launcher init container")

	opts := zap.Options{
		Development:     true,
		StacktraceLevel: zapcore.DPanicLevel,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   monitoringPort,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       leaderElectionID,
		Namespace:              namespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Prepare GangSchedulingSetupFunc
	gangSchedulingSetupFunc := common.GenNonGangSchedulerSetupFunc()
	if strings.EqualFold(gangSchedulerName, string(common.GangSchedulerVolcano)) {
		cfg := mgr.GetConfig()
		volcanoClientSet := volcanoclient.NewForConfigOrDie(cfg)
		gangSchedulingSetupFunc = common.GenVolcanoSetupFunc(volcanoClientSet)
	} else if strings.EqualFold(gangSchedulerName, string(common.GangSchedulerSchedulerPlugins)) {
		gangSchedulingSetupFunc = common.GenSchedulerPluginsSetupFunc(mgr.GetClient())
	}

	// TODO: We need a general manager. all rest reconciler addsToManager
	// Based on the user configuration, we start different controllers
	if enabledSchemes.Empty() {
		enabledSchemes.FillAll()
	}
	for _, s := range enabledSchemes {
		setupFunc, supported := controllerv1.SupportedSchemeReconciler[s]
		if !supported {
			setupLog.Error(fmt.Errorf("cannot find %s in supportedSchemeReconciler", s),
				"scheme not supported", "scheme", s)
			os.Exit(1)
		}
		if err = setupFunc(mgr, gangSchedulingSetupFunc, controllerThreads); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", s)
			os.Exit(1)
		}
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
