package kube

import (
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

func TestDefaultKubePath(t *testing.T) {
	file := clientcmd.NewDefaultPathOptions().GetDefaultFilename()
	t.Log(file)
}
