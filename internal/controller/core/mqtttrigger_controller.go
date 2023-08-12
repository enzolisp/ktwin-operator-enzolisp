package core

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev0 "ktwin/operator/api/core/v0"
	"ktwin/operator/pkg/event/rabbitmq"

	rabbitmqv1beta1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CLOUD_EVENT_DISPATCHER = "cloud-event-dispatcher"
	MQTT_DISPATCHER        = "mqtt-dispatcher"

	MQTT_DISPATCHER_QUEUE        = "mqtt-dispatcher-queue"
	CLOUD_EVENT_DISPATCHER_QUEUE = "cloud-event-dispatcher-queue"
)

// MQTTTriggerReconciler reconciles a MQTTTrigger object
type MQTTTriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.ktwin,resources=mqtttriggers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.ktwin,resources=mqtttriggers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.ktwin,resources=mqtttriggers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MQTTTrigger object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *MQTTTriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mqttTrigger := corev0.MQTTTrigger{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &mqttTrigger)

	// Delete scenario
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		logger.Error(err, fmt.Sprintf("Unexpected error while deleting TwinInstance %s", req.Name))
		return ctrl.Result{}, err
	}

	return r.createOrUpdateMQTTTrigger(ctx, req, mqttTrigger)
}

func (r *MQTTTriggerReconciler) createOrUpdateMQTTTrigger(ctx context.Context, req ctrl.Request, mqttTrigger corev0.MQTTTrigger) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// RabbitMQ Broker Secret
	rabbitMQSecret := v1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "rabbitmq-default-user",
		Namespace: "default",
	}, &rabbitMQSecret)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while getting rabbitmq default user secret %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	brokerCloudEventExchange := rabbitmqv1beta1.ExchangeList{}
	listOption := []client.ListOption{
		client.InNamespace("default"),
		client.MatchingLabels(client.MatchingFields{
			"eventing.knative.dev/broker": "default",
		}),
	}
	err = r.List(ctx, &brokerCloudEventExchange, listOption...)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while getting rabbitmq broker default exchange %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	if len(brokerCloudEventExchange.Items) == 0 {
		logger.Error(err, fmt.Sprintf("No rabbitmq broker default exchange %s found", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	defaultBrokerExchange := brokerCloudEventExchange.Items[0]

	// Create MQTT Dispatcher dependencies
	mqttDispatcherQueue := r.getMQQTDispatcherQueue(mqttTrigger)
	mqttDispacherDeployment := r.getMQQTDispatcherDeployment(mqttTrigger, rabbitMQSecret, defaultBrokerExchange)
	mqttDispacherService := r.getMQQTDispatcherService(mqttTrigger)

	err = r.Create(ctx, mqttDispatcherQueue, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating mqtt dispatcher queue %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, &mqttDispacherDeployment, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating mqtt dispatcher deployment %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, &mqttDispacherService, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating mqtt dispatcher service %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	// Create Cloud Event Dispatcher
	ceDispatcherQueue := r.getCloudEventDispatcherQueue(mqttTrigger)
	ceDispacherDeployment := r.getCloudEventDispatcherDeployment(mqttTrigger, rabbitMQSecret)
	ceDispacherService := r.geCloudEventDispatcherService(mqttTrigger)

	err = r.Create(ctx, ceDispatcherQueue, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating cloud event dispatcher queue %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, &ceDispacherDeployment, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating cloud event dispatcher deployment %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, &ceDispacherService, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating cloud event dispatcher service %s", mqttTrigger.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MQTTTriggerReconciler) getMQQTDispatcherQueue(mqttTrigger corev0.MQTTTrigger) *rabbitmqv1beta1.Queue {
	args := &rabbitmq.QueueArgs{
		Name:          MQTT_DISPATCHER_QUEUE,
		Namespace:     mqttTrigger.Namespace,
		QueueName:     MQTT_DISPATCHER_QUEUE,
		RabbitMQVhost: "/",
		RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
			Name:      "rabbitmq",
			Namespace: mqttTrigger.Namespace,
		},
		Owner: metav1.OwnerReference{
			APIVersion: mqttTrigger.APIVersion,
			Kind:       mqttTrigger.Kind,
			Name:       mqttTrigger.ObjectMeta.Name,
			UID:        mqttTrigger.ObjectMeta.UID,
		},
		Labels: map[string]string{},
	}
	return rabbitmq.NewQueue(args)
}

func (r *MQTTTriggerReconciler) getMQQTDispatcherDeployment(mqttTrigger corev0.MQTTTrigger, rabbitMQSecret v1.Secret, defaultBrokerExchange rabbitmqv1beta1.Exchange) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mqtt-dispatcher",
			Namespace: mqttTrigger.Namespace,
			Labels: map[string]string{
				"ktwin/trigger": "mqtt-dispatcher",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: mqttTrigger.APIVersion,
					Kind:       mqttTrigger.Kind,
					Name:       mqttTrigger.ObjectMeta.Name,
					UID:        mqttTrigger.ObjectMeta.UID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"ktwin/trigger": "mqtt-dispatcher",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"ktwin/trigger": "mqtt-dispatcher",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "mqtt-dispatcher",
							Image:           "dev.local/ktwin/mqtt-dispatcher:0.1",
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 5672,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "SERVICE_NAME",
									Value: MQTT_DISPATCHER + "-1",
								},
								{
									Name:  "PROTOCOL",
									Value: "amqp",
								},
								{
									Name:  "SERVER_URL",
									Value: string(rabbitMQSecret.Data["host"]),
								},
								{
									Name:  "SERVER_PORT",
									Value: string(rabbitMQSecret.Data["port"]),
								},
								{
									Name:  "USERNAME",
									Value: string(rabbitMQSecret.Data["username"]),
								},
								{
									Name:  "PASSWORD",
									Value: string(rabbitMQSecret.Data["password"]),
								},
								{
									Name:  "DECLARE_QUEUE",
									Value: "false",
								},
								{
									Name:  "DECLARE_EXCHANGE",
									Value: "false",
								},
								{
									Name:  "PUBLISHER_EXCHANGE",
									Value: defaultBrokerExchange.Spec.Name,
								},
								{
									Name:  "SUBSCRIBER_QUEUE",
									Value: MQTT_DISPATCHER_QUEUE,
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *MQTTTriggerReconciler) getCloudEventDispatcherQueue(mqttTrigger corev0.MQTTTrigger) *rabbitmqv1beta1.Queue {
	args := &rabbitmq.QueueArgs{
		Name:          CLOUD_EVENT_DISPATCHER_QUEUE,
		Namespace:     mqttTrigger.Namespace,
		QueueName:     CLOUD_EVENT_DISPATCHER_QUEUE,
		RabbitMQVhost: "/",
		RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
			Name:      "rabbitmq",
			Namespace: mqttTrigger.Namespace,
		},
		Owner: metav1.OwnerReference{
			APIVersion: mqttTrigger.APIVersion,
			Kind:       mqttTrigger.Kind,
			Name:       mqttTrigger.ObjectMeta.Name,
			UID:        mqttTrigger.ObjectMeta.UID,
		},
		Labels: map[string]string{},
	}
	return rabbitmq.NewQueue(args)
}

func (r *MQTTTriggerReconciler) getMQQTDispatcherService(mqttTrigger corev0.MQTTTrigger) v1.Service {
	return v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MQTT_DISPATCHER,
			Namespace: mqttTrigger.Namespace,
			Labels: map[string]string{
				"ktwin/trigger": MQTT_DISPATCHER,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: mqttTrigger.APIVersion,
					Kind:       mqttTrigger.Kind,
					Name:       mqttTrigger.ObjectMeta.Name,
					UID:        mqttTrigger.ObjectMeta.UID,
				},
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"ktwin/trigger": MQTT_DISPATCHER,
			},
			Ports: []v1.ServicePort{
				{
					Port:       5672,
					TargetPort: intstr.FromInt(5672),
					Protocol:   "TCP",
				},
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
}

func (r *MQTTTriggerReconciler) getCloudEventDispatcherDeployment(mqttTrigger corev0.MQTTTrigger, rabbitMQSecret v1.Secret) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CLOUD_EVENT_DISPATCHER,
			Namespace: mqttTrigger.Namespace,
			Labels: map[string]string{
				"ktwin/trigger": CLOUD_EVENT_DISPATCHER,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: mqttTrigger.APIVersion,
					Kind:       mqttTrigger.Kind,
					Name:       mqttTrigger.ObjectMeta.Name,
					UID:        mqttTrigger.ObjectMeta.UID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"ktwin/trigger": CLOUD_EVENT_DISPATCHER,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"ktwin/trigger": CLOUD_EVENT_DISPATCHER,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            CLOUD_EVENT_DISPATCHER,
							Image:           "dev.local/ktwin/cloud-event-dispatcher:0.1",
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 5672,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "SERVICE_NAME",
									Value: CLOUD_EVENT_DISPATCHER + "-1",
								},
								{
									Name:  "PROTOCOL",
									Value: "amqp",
								},
								{
									Name:  "SERVER_URL",
									Value: string(rabbitMQSecret.Data["host"]),
								},
								{
									Name:  "SERVER_PORT",
									Value: string(rabbitMQSecret.Data["port"]),
								},
								{
									Name:  "USERNAME",
									Value: string(rabbitMQSecret.Data["username"]),
								},
								{
									Name:  "PASSWORD",
									Value: string(rabbitMQSecret.Data["password"]),
								},
								{
									Name:  "DECLARE_QUEUE",
									Value: "false",
								},
								{
									Name:  "DECLARE_EXCHANGE",
									Value: "false",
								},
								{
									Name:  "PUBLISHER_EXCHANGE",
									Value: "amqp.topic",
								},
								{
									Name:  "SUBSCRIBER_QUEUE",
									Value: CLOUD_EVENT_DISPATCHER_QUEUE,
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *MQTTTriggerReconciler) geCloudEventDispatcherService(mqttTrigger corev0.MQTTTrigger) v1.Service {
	return v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CLOUD_EVENT_DISPATCHER,
			Namespace: mqttTrigger.Namespace,
			Labels: map[string]string{
				"ktwin/trigger": CLOUD_EVENT_DISPATCHER,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: mqttTrigger.APIVersion,
					Kind:       mqttTrigger.Kind,
					Name:       mqttTrigger.ObjectMeta.Name,
					UID:        mqttTrigger.ObjectMeta.UID,
				},
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"ktwin/trigger": CLOUD_EVENT_DISPATCHER,
			},
			Ports: []v1.ServicePort{
				{
					Port:       5672,
					TargetPort: intstr.FromInt(5672),
					Protocol:   "TCP",
				},
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
}

func int32Ptr(i int32) *int32 { return &i }

// SetupWithManager sets up the controller with the Manager.
func (r *MQTTTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev0.MQTTTrigger{}).
		Complete(r)
}
