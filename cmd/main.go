package main

import (
	"github.com/ketches/ingress-openresty/controller"
	"github.com/ketches/ingress-openresty/pkg/kube"
)

func main() {
	clientset := kube.GetClient()
	c := controller.NewIngressController(clientset)
	c.Run()
}
