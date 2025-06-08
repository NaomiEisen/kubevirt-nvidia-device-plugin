package testenv

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Flags
var (
	testConfigFlag = flag.String("test-config", "../test_config_h100.yaml", "Path to test config file")
	kubeconfigFlag *string
)

// Set kubeconfigFlag path
func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfigFlag = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfigFlag file")
	} else {
		kubeconfigFlag = flag.String("kubeconfig", "", "absolute path to the kubeconfigFlag file")
	}
}

func NewTestClient() (*TestClient, error) {
	flag.Parse()

	testConfig, err := LoadTestConfig(*testConfigFlag)
	if err != nil {
		return nil, err
	}

	configFromFlags, err := clientcmd.BuildConfigFromFlags("", *kubeconfigFlag)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(configFromFlags)
	if err != nil {
		return nil, err
	}

	return &TestClient{
		ClientSet: clientset,
		Config:    testConfig,
	}, nil
}

func LoadTestConfig(path string) (*TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TestConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}
