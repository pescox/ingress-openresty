package ingress

type Info struct {
	Namespace          string `json:"namespace"`
	Name               string `json:"name"`
	Host               string `json:"host"`
	Path               string `json:"path"`
	ServiceName        string `json:"serviceName"`
	ServicePort        int    `json:"servicePort"`
	PathRewriteEnabled bool   `json:"pathRewriteEnabled"`
}
