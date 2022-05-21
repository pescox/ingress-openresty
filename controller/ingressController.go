package controller

import (
	"context"
	"log"
	"time"

	"github.com/ketches/ingress-openresty/internal/ingress"
	"github.com/ketches/ingress-openresty/internal/openresty"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	informersnetworkingv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	listersnetworkingv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type ingressController struct {
	clientset          *kubernetes.Clientset
	ingressCacheSynced cache.InformerSynced
	ingressInformer    informersnetworkingv1.IngressInformer
	ingressLister      listersnetworkingv1.IngressLister
	queue              workqueue.RateLimitingInterface
	stopCh             chan struct{}
}

func NewIngressController(clientset *kubernetes.Clientset) *ingressController {
	informers := informers.NewSharedInformerFactoryWithOptions(clientset, time.Minute*30, informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
		// lo.LabelSelector = "ingressClass=ingress-openresty"
	}))

	stopCh := make(chan struct{})
	informers.Start(stopCh)
	ingInformer := informers.Networking().V1().Ingresses()

	c := &ingressController{
		clientset:          clientset,
		ingressCacheSynced: ingInformer.Informer().HasSynced,
		ingressInformer:    ingInformer,
		ingressLister:      ingInformer.Lister(),
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress-openresty"),
		stopCh:             stopCh,
	}

	ingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) { log.Println("add"); c.queue.Add(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldIng, ok1 := oldObj.(networkingv1.Ingress)
			newIng, ok2 := newObj.(networkingv1.Ingress)
			if ok1 && ok2 && oldIng.ResourceVersion < newIng.ResourceVersion {
				c.queue.Add(newObj)
			}
		},
		DeleteFunc: func(obj interface{}) { log.Println("delete"); c.queue.Add(obj) },
	})

	klog.Infoln("ingres-openresty controller start.")
	informers.Start(stopCh)
	return c
}

func (c *ingressController) Run() {
	if !cache.WaitForCacheSync(c.stopCh, c.ingressCacheSynced) {
		klog.Infoln("waiting cache to be synced.")
	}
	go wait.Until(c.worker, time.Second, c.stopCh)
	<-c.stopCh
}

func (c *ingressController) worker() {
	for c.processItem() {

	}
}

func (c *ingressController) processItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	defer c.queue.Forget(item)
	if ok := c.syncItem(item); !ok {
		klog.Infoln("syncing ingress failed.")
		return false
	}
	return true
}

func (c *ingressController) syncItem(obj interface{}) bool {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return false
	}

	apiserverIngress, ok := c.getIngressFromApiServer(ing)
	if !ok {
		// delete ingress, then delete openresty config & reload
		ingressInfo := ingress.Info{}
		serverConf := openresty.BuildHttpServerConfig(ingressInfo)
		openresty.DeleteConfigAndReload(serverConf)
	} else {
		if apiserverIngress != nil {
			// add or update ingress, then update openresty config & reload
			ingressInfo := ingress.Info{}
			serverConf := openresty.BuildHttpServerConfig(ingressInfo)
			openresty.UpdateConfigAndReload(serverConf)
		} else {
			return false
		}
	}
	return true
}

func (c *ingressController) getIngressFromApiServer(cacheIngress *networkingv1.Ingress) (*networkingv1.Ingress, bool) {
	ingress, err := c.clientset.NetworkingV1().Ingresses(cacheIngress.Namespace).Get(context.Background(), cacheIngress.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// not fount, ingress is deleted.
			return nil, false
		}
		return nil, true
	}
	return ingress, true
}
