package openresty

import "github.com/ketches/ingress-openresty/internal/ingress"

type HttpServerConfig struct {
	Name         string
	Host         string
	Path         string
	Port         string
	Upstream     []string
	HttpsEnabled bool
}

func BuildHttpServerConfig(ing ingress.Info) *HttpServerConfig {
	return &HttpServerConfig{}
}

func UpdateConfigAndReload(conf *HttpServerConfig) {

}

func DeleteConfigAndReload(conf *HttpServerConfig) {

}
