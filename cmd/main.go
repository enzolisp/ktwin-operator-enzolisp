/*
Copyright 2023.

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

	corev0 "github.com/Open-Digital-Twin/ktwin-operator/api/core/v0"
	dtdv0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
	corecontroller "github.com/Open-Digital-Twin/ktwin-operator/internal/controller/core"
	dtdcontroller "github.com/Open-Digital-Twin/ktwin-operator/internal/controller/dtd"
	"github.com/Open-Digital-Twin/ktwin-operator/pkg/event"
	eventStore "github.com/Open-Digital-Twin/ktwin-operator/pkg/event-store"
	"github.com/Open-Digital-Twin/ktwin-operator/pkg/graph"
	"github.com/Open-Digital-Twin/ktwin-operator/pkg/service"

	rabbitmqv1beta1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	keventing "knative.dev/eventing/pkg/apis/eventing/v1"
	kserving "knative.dev/serving/pkg/apis/serving/v1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(dtdv0.AddToScheme(scheme))
	utilruntime.Must(corev0.AddToScheme(scheme))

	// Third party
	utilruntime.Must(kserving.AddToScheme(scheme))
	utilruntime.Must(keventing.AddToScheme(scheme))
	utilruntime.Must(rabbitmqv1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// ClusterRole permissions

// Kubernetes resources
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// KNative resources
//+kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=triggers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers,verbs=get;list;watch;create;update;patch;delete

// RabbitMQ Resources
// +kubebuilder:rbac:groups=rabbitmq.com,resources=exchanges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rabbitmq.com,resources=queues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rabbitmq.com,resources=bindings,verbs=get;list;watch;create;update;patch;delete

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
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

	// Define TwinGraph Server
	twinGraphServer := graph.NewTwinGraphServer()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "cf3f9c51.ktwin",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&dtdcontroller.TwinInterfaceReconciler{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		TwinService: service.NewTwinService(),
		TwinEvent:   event.NewTwinEvent(),
		EventStore:  eventStore.NewEventStore(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TwinInterface")
		os.Exit(1)
	}
	if err = (&dtdcontroller.TwinInstanceReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		TwinService:     service.NewTwinService(),
		TwinEvent:       event.NewTwinEvent(),
		EventStore:      eventStore.NewEventStore(),
		TwinGraphServer: twinGraphServer,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TwinInstance")
		os.Exit(1)
	}
	if err = (&corecontroller.GatewayReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Gateway")
		os.Exit(1)
	}
	if err = (&corecontroller.MQTTTriggerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MQTTTrigger")
		os.Exit(1)
	}
	if err = (&corecontroller.EventStoreReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		EventStore: eventStore.NewEventStore(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "EventStore")
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

	// Expose the Twin Graph to be consumed by services
	mgr.AddMetricsExtraHandler("/twin-graph", twinGraphServer.HandleGraphFunc())

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
