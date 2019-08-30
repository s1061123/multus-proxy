package proxy

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"
	informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller"
)

// ServiceHandler is an abstract interface of objects which receive
// notifications about service object changes.
type ServiceHandler interface {
    // OnServiceAdd is called whenever creation of new service object
    // is observed.
    OnServiceAdd(service *corev1.Service)
    // OnServiceUpdate is called whenever modification of an existing
    // service object is observed.
    OnServiceUpdate(oldService, service *corev1.Service)
    // OnServiceDelete is called whenever deletion of an existing service
    // object is observed.
    OnServiceDelete(service *corev1.Service)
    // OnServiceSynced is called once all the initial event handlers were
    // called and the state is fully propagated to local cache.
    OnServiceSynced()
}

// EndpointsHandler is an abstract interface of objects which receive
// notifications about endpoints object changes.
type EndpointsHandler interface {
    // OnEndpointsAdd is called whenever creation of new endpoints object
    // is observed.
    OnEndpointsAdd(endpoints *corev1.Endpoints)
    // OnEndpointsUpdate is called whenever modification of an existing
    // endpoints object is observed.
    OnEndpointsUpdate(oldEndpoints, endpoints *corev1.Endpoints)
    // OnEndpointsDelete is called whenever deletion of an existing endpoints
    // object is observed.
    OnEndpointsDelete(endpoints *corev1.Endpoints)
    // OnEndpointsSynced is called once all the initial event handlers were
    // called and the state is fully propagated to local cache.
    OnEndpointsSynced()
}

// EndpointsConfig tracks a set of endpoints configurations.
type EndpointsConfig struct {
    listerSynced  cache.InformerSynced
    eventHandlers []EndpointsHandler
}

// NewEndpointsConfig creates a new EndpointsConfig.
func NewEndpointsConfig(endpointsInformer informers.EndpointsInformer, resyncPeriod time.Duration) *EndpointsConfig {
    result := &EndpointsConfig{
        listerSynced: endpointsInformer.Informer().HasSynced,
    }

    endpointsInformer.Informer().AddEventHandlerWithResyncPeriod(
        cache.ResourceEventHandlerFuncs{
            AddFunc:    result.handleAddEndpoints,
            UpdateFunc: result.handleUpdateEndpoints,
            DeleteFunc: result.handleDeleteEndpoints,
        },
        resyncPeriod,
    )

    return result
}

// RegisterEventHandler registers a handler which is called on every endpoints change.
func (c *EndpointsConfig) RegisterEventHandler(handler EndpointsHandler) {
    c.eventHandlers = append(c.eventHandlers, handler)
}

// Run waits for cache synced and invokes handlers after syncing.
func (c *EndpointsConfig) Run(stopCh <-chan struct{}) {
    klog.Info("Starting endpoints config controller")

    if !controller.WaitForCacheSync("endpoints config", stopCh, c.listerSynced) {
        return
    }

    for i := range c.eventHandlers {
        klog.V(3).Infof("Calling handler.OnEndpointsSynced()")
        c.eventHandlers[i].OnEndpointsSynced()
    }
}

func (c *EndpointsConfig) handleAddEndpoints(obj interface{}) {
    endpoints, ok := obj.(*corev1.Endpoints)
    if !ok {
        utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
        return
    }
    for i := range c.eventHandlers {
        klog.V(4).Infof("Calling handler.OnEndpointsAdd")
        c.eventHandlers[i].OnEndpointsAdd(endpoints)
    }
}

func (c *EndpointsConfig) handleUpdateEndpoints(oldObj, newObj interface{}) {
    oldEndpoints, ok := oldObj.(*corev1.Endpoints)
    if !ok {
        utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
        return
    }
    endpoints, ok := newObj.(*corev1.Endpoints)
    if !ok {
        utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
        return
    }
    for i := range c.eventHandlers {
        klog.V(4).Infof("Calling handler.OnEndpointsUpdate")
        c.eventHandlers[i].OnEndpointsUpdate(oldEndpoints, endpoints)
    }
}

func (c *EndpointsConfig) handleDeleteEndpoints(obj interface{}) {
    endpoints, ok := obj.(*corev1.Endpoints)
    if !ok {
        tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
        if !ok {
            utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
            return
        }
        if endpoints, ok = tombstone.Obj.(*corev1.Endpoints); !ok {
            utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
            return
        }
    }
    for i := range c.eventHandlers {
        klog.V(4).Infof("Calling handler.OnEndpointsDelete")
        c.eventHandlers[i].OnEndpointsDelete(endpoints)
    }
}

// ServiceConfig tracks a set of service configurations.
type ServiceConfig struct {
    listerSynced  cache.InformerSynced
    eventHandlers []ServiceHandler
}

// NewServiceConfig creates a new ServiceConfig.
func NewServiceConfig(serviceInformer informers.ServiceInformer, resyncPeriod time.Duration) *ServiceConfig {
    result := &ServiceConfig{
        listerSynced: serviceInformer.Informer().HasSynced,
    }

    serviceInformer.Informer().AddEventHandlerWithResyncPeriod(
        cache.ResourceEventHandlerFuncs{
            AddFunc:    result.handleAddService,
            UpdateFunc: result.handleUpdateService,
            DeleteFunc: result.handleDeleteService,
        },
        resyncPeriod,
    )

    return result
}

// RegisterEventHandler registers a handler which is called on every service change.
func (c *ServiceConfig) RegisterEventHandler(handler ServiceHandler) {
    c.eventHandlers = append(c.eventHandlers, handler)
}

// Run waits for cache synced and invokes handlers after syncing.
func (c *ServiceConfig) Run(stopCh <-chan struct{}) {
    klog.Info("Starting service config controller")

    if !controller.WaitForCacheSync("endpoints config", stopCh, c.listerSynced) {
        return
    }
    /*
    */

    for i := range c.eventHandlers {
        klog.V(3).Info("Calling handler.OnServiceSynced()")
        c.eventHandlers[i].OnServiceSynced()
    }
}

func (c *ServiceConfig) handleAddService(obj interface{}) {
    service, ok := obj.(*corev1.Service)
    if !ok {
        utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
        return
    }
    for i := range c.eventHandlers {
        klog.V(4).Info("Calling handler.OnServiceAdd")
        c.eventHandlers[i].OnServiceAdd(service)
    }
}

func (c *ServiceConfig) handleUpdateService(oldObj, newObj interface{}) {
    oldService, ok := oldObj.(*corev1.Service)
    if !ok {
        utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
        return
    }
    service, ok := newObj.(*corev1.Service)
    if !ok {
        utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
        return
    }
    for i := range c.eventHandlers {
        klog.V(4).Info("Calling handler.OnServiceUpdate")
        c.eventHandlers[i].OnServiceUpdate(oldService, service)
    }
}

func (c *ServiceConfig) handleDeleteService(obj interface{}) {
    service, ok := obj.(*corev1.Service)
    if !ok {
        tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
        if !ok {
            utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
            return
        }
        if service, ok = tombstone.Obj.(*corev1.Service); !ok {
            utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
            return
        }
    }
    for i := range c.eventHandlers {
        klog.V(4).Info("Calling handler.OnServiceDelete")
        c.eventHandlers[i].OnServiceDelete(service)
    }
}
