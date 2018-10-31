package kube

import (
	"k8s.io/helm/pkg/kube"
)

type KubeLayer struct {
	tillerTunnel *kube.Tunnel
}

func New() *KubeLayer {
	k := &Kube{}
	k.Meta = make(map[string]string)
	return k
}

func (k *KubeLayer) PrintMeta() {
	fmt.Printf("tillerTunnel: %v", k.tillerTunnel)
}
