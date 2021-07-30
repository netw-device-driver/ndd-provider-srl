/*
Copyright 2021 Wim Henderickx.

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

package srl

import (
	"context"
	"strconv"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	ndrv1 "github.com/netw-device-driver/ndd-core/apis/dvr/v1"
	"github.com/netw-device-driver/ndd-grpc/ndd"
	regclient "github.com/netw-device-driver/ndd-grpc/register/client"
	register "github.com/netw-device-driver/ndd-grpc/register/registerpb"
	srlv1 "github.com/netw-device-driver/ndd-provider-srl/apis/srl/v1"
	nddv1 "github.com/netw-device-driver/ndd-runtime/apis/common/v1"
	"github.com/netw-device-driver/ndd-runtime/pkg/event"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	"github.com/netw-device-driver/ndd-runtime/pkg/meta"
	"github.com/netw-device-driver/ndd-runtime/pkg/reconciler/managed"
	"github.com/netw-device-driver/ndd-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	// Finalizer
	RegistrationFinalizer = "Registration.srl.ndd.henderiw.be"

	// Timers
	//reconcileTimeout = 1 * time.Minute
	//shortWait        = 30 * time.Second
	//veryShortWait    = 5 * time.Second

	// Errors
	errUnexpectedObject   = "the managed resource is not a Registration resource"
	errTrackTCUsage       = "cannot track TargetConfig usage"
	errGetTC              = "cannot get TargetConfig"
	errGetNetworkNode     = "cannot get NetworkNode"
	errNewClient          = "cannot create new client"
	errKubeUpdateFailed   = "cannot update Registration custom resource"
	targetNotConfigured   = "target is not configured to proceed"
	errRegistrationRead   = "cannot read Registration"
	errRegistrationCreate = "cannot create Registration"
	errRegistrationUpdate = "cannot update Registration"
	errRegistrationDelete = "cannot delete Registration"

	// old erros
	//errGetRegistration             = "cannot get Registration"
	//errAddRegistrationFinalizer    = "cannot add Registration finalizer"
	//errRemoveRegistrationFinalizer = "cannot remove Registration finalizer"
	//errUpdateRegistrationStatus    = "cannot update Registration status"
	//errRegistrationFailed          = "cannot register to the device driver"
	//errDeRegistrationFailed        = "cannot desregister to the device driver"
	//errCacheStatusFailed           = "cannot get cache status from the device driver"

	// Event reasons
	//reasonSync event.Reason = "SyncRegistration"
)

/*
// RegistrationReconcilerOption is used to configure the RegistrationReconciler.
type RegistrationReconcilerOption func(*RegistrationReconciler)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.log = log
	}
}

// WithRecorder specifies how the Reconciler should record Kubernetes events.
func WithRecorder(er event.Recorder) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.record = er
	}
}

// WithValidator specifies how the Reconciler should perform object
// validation.
func WithValidator(v Validator) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.validator = v
	}
}

func WithGrpcApplicator(a gclient.Applicator) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.applicator = a
	}
}

// RegistrationReconciler reconciles a Registration object
type RegistrationReconciler struct {
	client     client.Client
	finalizer  resource.Finalizer
	validator  Validator
	log        logging.Logger
	record     event.Recorder
	applicator gclient.Applicator
}
*/

// SetupRegistration adds a controller that reconciles Registrations.
func SetupRegistration(mgr ctrl.Manager, o controller.Options, l logging.Logger, poll time.Duration, namespace string) error {

	name := managed.ControllerName(srlv1.RegistrationGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(srlv1.RegistrationGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			log:         l,
			kube:        mgr.GetClient(),
			usage:       resource.NewNetworkNodeUsageTracker(mgr.GetClient(), &ndrv1.NetworkNodeUsage{}),
			newClientFn: regclient.NewClient},
		),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&srlv1.Registration{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		//Watches(
		//	&source.Kind{Type: &ndrv1.NetworkNode{}},
		//	handler.EnqueueRequestsFromMapFunc(r.NetworkNodeMapFunc),
		//).
		Complete(r)

	/*
		r := NewRegistrationReconciler(mgr,
			WithLogger(l.WithValues("controller", name)),
			WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			WithValidator(NewTargetValidator(resource.ClientApplicator{
				Client:     mgr.GetClient(),
				Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
			}, l)),
			WithGrpcApplicator(gclient.NewClientApplicator(
				gclient.WithLogger(l.WithValues("gclient", name)),
				gclient.WithInsecure(true),
				gclient.WithSkipVerify(true),
			)),
		)

		return ctrl.NewControllerManagedBy(mgr).
			Named(name).
			For(&srlv1.Registration{}).
			WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
			WithOptions(option).
			Watches(
				&source.Kind{Type: &ndrv1.NetworkNode{}},
				handler.EnqueueRequestsFromMapFunc(r.NetworkNodeMapFunc),
			).
			Complete(r)
	*/
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	log         logging.Logger
	kube        client.Client
	usage       resource.Tracker
	newClientFn func(ctx context.Context, cfg ndd.Config) (register.RegistrationClient, error)
}

// Connect produces an ExternalClient by:
// 1. Tracking that the managed resource is using a TargetConfig.
// 2. Getting the managed resource's TargetConfig with connection details
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	log := c.log.WithValues("resosurce", mg.GetName())
	log.Debug("Connect")
	_, ok := mg.(*srlv1.Registration)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackTCUsage)
	}

	selectors := []client.ListOption{}
	nnl := &ndrv1.NetworkNodeList{}
	if err := c.kube.List(ctx, nnl, selectors...); err != nil {
		return nil, errors.Wrap(err, errGetNetworkNode)
	}

	// find all targets that have are in configured status
	var ts []*nddv1.Target
	for _, nn := range nnl.Items {
		log.Debug("Network Node", "Name", nn.GetName(), "Status", nn.GetCondition(ndrv1.ConditionKindDeviceDriverConfigured).Status)
		if nn.GetCondition(ndrv1.ConditionKindDeviceDriverConfigured).Status == corev1.ConditionTrue {
			t := &nddv1.Target{
				Name: nn.GetName(),
				Cfg: ndd.Config{
					SkipVerify: true,
					Insecure:   true,
					Target:     ndrv1.PrefixService + "-" + nn.Name + "." + ndrv1.NamespaceLocalK8sDNS + strconv.Itoa(*nn.Spec.GrpcServerPort),
				},
			}
			ts = append(ts, t)
		}
	}
	log.Debug("Active targets", "targets", ts)

	// We dont have to update the deletion since the network device driver would have lost
	// its state already
	/*
		// check if a target got deleted and if so delete it
		for _, t := range ts {
			if IsTargetDeleted(t.Name, cr.Status.AtProvider.Targets) {
				cl, err := c.newClientFn(ctx, t.Cfg)
				if err != nil {
					return nil, errors.Wrap(err, errNewClient)
				}
				cl.Delete(ctx, &register.DeviceType{
					DeviceType: string(srlv1.DeviceTypeSRL),
				})
			}
		}
	*/

	//get clients for each target
	cls := make([]register.RegistrationClient, 0)
	tns := make([]string, 0)
	for _, t := range ts {
		cl, err := c.newClientFn(ctx, t.Cfg)
		if err != nil {
			return nil, errors.Wrap(err, errNewClient)
		}
		cls = append(cls, cl)
		tns = append(tns, t.Name)
	}

	log.Debug("Connect info", "clients", cls, "targets", tns)

	return &external{clients: cls, targets: tns, log: log}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	clients []register.RegistrationClient
	targets []string
	log     logging.Logger
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, ok := mg.(*srlv1.Registration)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(o) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	for _, cl := range e.clients {
		r, err := cl.Read(ctx, &register.DeviceType{
			DeviceType: string(srlv1.DeviceTypeSRL),
		})
		if err != nil {
			// if a single network device driver reports an error this is applicable to all
			// network devices
			return managed.ExternalObservation{}, errors.New(errRegistrationRead)
		}
		// if a network device driver reports a different device type we trigger
		// a recreation of the configuration on all devices by returning
		// Exists = false and
		if r.DeviceType != string(srlv1.DeviceTypeSRL) {
			return managed.ExternalObservation{
				ResourceExists:    false,
				ResourceUpToDate:  false,
				ConnectionDetails: managed.ConnectionDetails{},
			}, nil
		}
	}

	// when all network device driver reports the proper device type
	// we return exists and up to date
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil

}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*srlv1.Registration)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	e.log.Debug("creating", "object", cr)

	for _, cl := range e.clients {
		_, err := cl.Create(ctx, &register.RegistrationInfo{
			DeviceType:             string(srlv1.DeviceTypeSRL),
			MatchString:            srlv1.DeviceMatch,
			Subscriptions:          cr.GetSubscriptions(),
			ExceptionPaths:         cr.GetExceptionPaths(),
			ExplicitExceptionPaths: cr.GetExplicitExceptionPaths(),
		})
		if err != nil {
			return managed.ExternalCreation{}, errors.New(errRegistrationCreate)
		}
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*srlv1.Registration)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	e.log.Debug("creating", "object", cr)

	for _, cl := range e.clients {
		_, err := cl.Update(ctx, &register.RegistrationInfo{
			DeviceType:             string(srlv1.DeviceTypeSRL),
			MatchString:            srlv1.DeviceMatch,
			Subscriptions:          cr.GetSubscriptions(),
			ExceptionPaths:         cr.GetExceptionPaths(),
			ExplicitExceptionPaths: cr.GetExplicitExceptionPaths(),
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.New(errRegistrationUpdate)
		}
	}

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*srlv1.Registration)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	e.log.Debug("deleting", "object", cr)

	for _, cl := range e.clients {
		_, err := cl.Delete(ctx, &register.DeviceType{
			DeviceType: string(srlv1.DeviceTypeSRL),
		})
		if err != nil {
			return errors.New(errRegistrationDelete)
		}
	}

	return nil
}

func (e *external) GetTarget() []string {
	return e.targets
}

/*
// NewRegistrationReconciler creates a new package revision reconciler.
func NewRegistrationReconciler(mgr manager.Manager, opts ...RegistrationReconcilerOption) *RegistrationReconciler {
	r := &RegistrationReconciler{
		client:    mgr.GetClient(),
		finalizer: resource.NewAPIFinalizer(mgr.GetClient(), RegistrationFinalizer),
		log:       logging.NewNopLogger(),
		record:    event.NewNopRecorder(),
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

// NetworkNodeMapFunc is a handler.ToRequestsFunc to be used to enqeue
// request for reconciliation of Registration.
func (r *RegistrationReconciler) NetworkNodeMapFunc(o client.Object) []ctrl.Request {
	log := r.log.WithValues("NetworkNode Object", o)
	result := []ctrl.Request{}

	nn, ok := cr.(*ndrv1.NetworkNode)
	if !ok {
		panic(fmt.Sprintf("Expected a NodeTopology but got a %T", o))
	}
	log.WithValues(nn.GetName(), nn.GetNamespace()).Info("NetworkNode MapFunction")

	selectors := []client.ListOption{
		client.InNamespace(nn.Namespace),
		client.MatchingLabels{},
	}
	ss := &srlv1.RegistrationList{}
	if err := r.client.List(context.TODO(), ss, selectors...); err != nil {
		return result
	}

	for _, o := range ss.Items {
		name := client.ObjectKey{
			Namespace: cr.GetNamespace(),
			Name:      cr.GetName(),
		}
		log.WithValues(cr.GetName(), cr.GetNamespace()).Info("NetworkNodeMapFunc Registration ReQueue")
		result = append(result, ctrl.Request{NamespacedName: name})
	}
	return result
}

//+kubebuilder:rbac:groups=dvr.ndd.henderiw.be,resources=networknodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=srl.ndd.henderiw.be,resources=registrations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=srl.ndd.henderiw.be,resources=registrations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=srl.ndd.henderiw.be,resources=registrations/finalizers,verbs=update

func (r *RegistrationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := r.log.FromContext(ctx)
	log := r.log.WithValues("Registration", req.NamespacedName)
	log.Debug("reconciling Registration")

	o := &srlv1.Registration{}
	if err := r.client.Get(ctx, req.NamespacedName, o); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug(errGetRegistration, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetRegistration)
	}
	log.Debug("Registration", "Sub", o)

	if meta.WasDeleted(o) {
		// find the available targets and deregister the provider
		targets, err := r.validator.FindConfigured(ctx, cr.Spec.TargetConfigReference.Name)
		if err != nil {
			log.Debug(nddv1.ErrFindingTargets, "error", err)
			r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, nddv1.ErrFindingTargets)))
			return reconcile.Result{RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
		}
		for _, target := range targets {
			// register the provider to the device driver
			_, err := r.applicator.DeRegister(ctx, target.DNS,
				&netwdevpb.RegistrationRequest{
					DeviceType:  string(srlv1.DeviceTypeSRL),
					MatchString: srlv1.DeviceMatch,
				})
			if err != nil {
				log.Debug(errDeRegistrationFailed)
				r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, fmt.Sprintf("%s, target, %s", errDeRegistrationFailed, target))))
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(err, fmt.Sprintf("%s, target, %s", errDeRegistrationFailed, target))
			}
		}

		// Delete finalizer after all device drivers are deregistered
		if err := r.finalizer.RemoveFinalizer(ctx, o); err != nil {
			log.Debug(errRemoveRegistrationFinalizer, "error", err)
			r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, errRemoveRegistrationFinalizer)))
			return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(err, errRemoveRegistrationFinalizer)
		}
		return reconcile.Result{Requeue: false}, nil
	}

	if err := r.finalizer.AddFinalizer(ctx, o); err != nil {
		log.Debug(errAddRegistrationFinalizer, "error", err)
		r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, errAddRegistrationFinalizer)))
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(err, errAddRegistrationFinalizer)
	}

	log = log.WithValues(
		"uid", cr.GetUID(),
		"version", cr.GetResourceVersion(),
		"name", cr.GetName(),
	)

	// find targets the resource should be applied to
	targets, err := r.validator.FindConfigured(ctx, cr.Spec.TargetConfigReference.Name)
	if err != nil {
		log.Debug(nddv1.ErrTargetNotFound, "error", err)
		r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, nddv1.ErrTargetNotFound)))
		cr.SetConditions(nddv1.TargetNotFound())
		return reconcile.Result{RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
	}

	// check if targets got deleted and if so delete them from the status
	for targetName := range cr.Status.TargetConditions {
		if r.validator.IfDeleted(ctx, targets, targetName) {
			cr.DeleteTargetCondition(targetName)
			log.Debug(nddv1.InfoTargetDeleted, "target", targetName)
			r.record.Event(o, event.Normal(reasonSync, nddv1.InfoTargetDeleted, "target", targetName))
		}
	}

	// if no targets are found we return and update object status with no target found
	if len(targets) == 0 {
		log.Debug(nddv1.ErrTargetNotFound)
		r.record.Event(o, event.Warning(reasonSync, errors.New(nddv1.ErrTargetNotFound)))
		cr.SetConditions(nddv1.TargetNotFound())
		return reconcile.Result{RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
	}
	// otherwise register the provider to the device driver
	r.record.Event(o, event.Normal(reasonSync, nddv1.InfoTargetFound))
	cr.SetConditions(nddv1.TargetFound())

	for _, target := range targets {
		log.Debug("RegistrationTargets", "target", target)
		// initialize the targets ConditionedStatus
		if len(cr.Status.TargetConditions) == 0 {
			cr.InitializeTargetConditions()
		}

		// set condition if not yet set
		if _, ok := cr.Status.TargetConditions[target.Name]; !ok {
			log.Debug("Initialize target condition")
			cr.SetTargetConditions(target.Name, nddv1.Unknown())
		}

		// update the device driver with the object info when the condition is not met yet
		log.Debug("targetconditionStatus", "status", cr.GetTargetCondition(target.Name, nddv1.ConditionKindConfiguration).Status)
		if cr.GetTargetCondition(target.Name, nddv1.ConditionKindConfiguration).Status == corev1.ConditionFalse ||
			cr.GetTargetCondition(target.Name, nddv1.ConditionKindConfiguration).Status == corev1.ConditionUnknown {

			// register the provider to the device driver
			rsp, err := r.applicator.Register(ctx, target.DNS,
				&netwdevpb.RegistrationRequest{
					DeviceType:             string(srlv1.DeviceTypeSRL),
					MatchString:            srlv1.DeviceMatch,
					Subscriptions:          cr.GetSubscriptions(),
					ExcpetionPaths:         cr.GetExceptionPaths(),
					ExplicitExceptionPaths: cr.GetExplicitExceptionPaths(),
				})
			if err != nil {
				log.Debug(errRegistrationFailed)
				r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, errRegistrationFailed)))
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)

			}
			log.Debug("Register", "response", rsp)
			log.Debug("Object status", "status", cr.Status)
			cr.SetTargetConditions(target.Name, nddv1.ReconcileSuccess())
		}
	}
	// update the status and return
	return reconcile.Result{Requeue: false, RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
}
*/
