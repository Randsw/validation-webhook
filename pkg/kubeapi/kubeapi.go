package kubeapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/randsw/validationwebhook/pkg/logger"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetKubeConfig() (*rest.Config, error) {

	var err error
	var config *rest.Config

	// Try reading kubeconfig file form the environment variable
	if envVar := os.Getenv("KUBECONFIG"); len(envVar) > 0 {
		if config, err = clientcmd.BuildConfigFromFlags("", envVar); err == nil {
			return config, nil
		}
	}

	errorMsg := "error loading kubeconfig from environment variable KUBECONFIG"

	// Try getting kube config from the home directory
	if home := homedir.HomeDir(); home != "" {

		kubeconfig := filepath.Join(home, ".kube", "config")
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err == nil {
			return config, nil
		}

	}

	errorMsg = errorMsg + "\n" + err.Error()

	// finally try to get in-cluster config via service account
	if config, err = rest.InClusterConfig(); err == nil {
		return config, nil
	}

	errorMsg = errorMsg + "\n" + err.Error()

	return config, fmt.Errorf("failed to authenticate with the Kubernets cluster - \n%v", errorMsg)
}

// NewKubeClient - returns a new kuberenets client set
func NewKubeClient(config *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

func InitKubeApiConnection() (kubernetes.Interface, error) {
	config, err := GetKubeConfig()

	if err != nil {
		logger.Fatal("Fail to get kubernetes config", zap.Error(err))
	}

	return NewKubeClient(config)
}

func CheckNamespaceAnnotationTrue(client kubernetes.Interface, annotation, namespace string) (bool, error) {

	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})

	if err != nil {
		nsErr := fmt.Errorf("error checking annotations on the namespace %v - %v", namespace, err)
		logger.Error("error checking annotations on the namespace", zap.String("namespace", namespace), zap.String("err", err.Error()))
		return false, nsErr
	}

	for key, val := range ns.GetAnnotations() {
		if key == annotation && strings.ToLower(val) == "true" {
			logger.Info("Found VALIDATE annotations on the namespace", zap.String("annotation", annotation), zap.String("value", val), zap.String("namespace", namespace))
			return true, nil
		}
	}

	return false, nil
}
