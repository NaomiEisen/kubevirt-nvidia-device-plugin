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
	kubecli "kubevirt.io/client-go/kubecli"
)

const (
	// TODO: change the default to something normal
	DefaultTestConfigPath = "../test_config_h100.yaml"
)

// Flags
var (
	testConfigFlag = flag.String("test-config", DefaultTestConfigPath, "Path to test config file")
	kubeconfigFlag *string
)

// Set kubeconfigFlag path
func init() {
	usage := "absolute path to the kubeconfigFlag file"

	if home := homedir.HomeDir(); home != "" {
		kubeconfigFlag = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), usage)
	} else {
		kubeconfigFlag = flag.String("kubeconfig", "", usage)
	}
}

func NewTestClient() (*TestClient, error) {
	flag.Parse()

	testConfig, err := LoadTestConfig(*testConfigFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to load kube config: %w", err)
	}

	configFromFlags, err := clientcmd.BuildConfigFromFlags("", *kubeconfigFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to build kube client from flags: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(configFromFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kube client: %w", err)
	}

	virtClient, err := kubecli.GetKubevirtClientFromRESTConfig(configFromFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize KubeVirt client: %w", err)
	}

	return &TestClient{
		ClientSet:      clientset,
		KubeVirtClient: virtClient,
		Config:         testConfig,
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
