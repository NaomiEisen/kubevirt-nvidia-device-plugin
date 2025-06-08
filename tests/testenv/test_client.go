package testenv

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type TestClient struct {
	ClientSet kubernetes.Interface
	Config    *TestConfig
}

type TestConfig struct {
	DevicePluginName      string     `yaml:"deviceplugin_name"`
	DevicePluginNamespace string     `yaml:"deviceplugin_namespace"`
	Nodes                 []NodeInfo `yaml:"nodes"`
}

type NodeInfo struct {
	Name    string       `yaml:"name"`
	Devices []DeviceInfo `yaml:"devices"`
}

type DeviceInfo struct {
	Name   string `yaml:"name"`
	Number string `yaml:"number"`
}

// ---- Client methods ----

func GetNewClientTest() (*TestClient, error) {
	return NewTestClient()
}

func (t *TestClient) GetNode(nodeName string) (*corev1.Node, error) {
	return t.ClientSet.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
}

func (t *TestClient) GetWorkerNodes() ([]corev1.Node, error) {
	nodes, err := t.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	workerNodes := []corev1.Node{}
	for _, node := range nodes.Items {
		labels := node.Labels

		// Skip nodes labeled as master/control-plane
		if _, isMaster := labels["node-role.kubernetes.io/master"]; isMaster {
			continue
		}
		if _, isControlPlane := labels["node-role.kubernetes.io/control-plane"]; isControlPlane {
			continue
		}

		// Skip explicitly unschedulable nodes
		if node.Spec.Unschedulable {
			continue
		}

		workerNodes = append(workerNodes, node)
	}
	return workerNodes, nil
}

func (t *TestClient) GetAllocatableDeviceQuantity(nodeName string, deviceName string) (string, error) {
	node, err := t.GetNode(nodeName)
	if err != nil {
		return "", err
	}
	quantity, exists := node.Status.Allocatable[corev1.ResourceName(deviceName)]
	if !exists {
		return "", errors.New(fmt.Sprintf("device %s not found in allocatable resources for node %s", deviceName, nodeName))
	}

	return quantity.String(), nil
}

func (t *TestClient) GetPodsList(prefix string, namespace string) ([]corev1.Pod, error) {
	if prefix == "" || namespace == "" {
		return nil, errors.New("prefix or namespace is empty")
	}
	podList, err := t.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var matchedPods []corev1.Pod
	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.Name, prefix) {
			matchedPods = append(matchedPods, pod)
		}
	}
	return matchedPods, nil
}

func (t *TestClient) GetPodsStatusMap(pods []corev1.Pod) (map[string]corev1.PodPhase, error) {
	if pods == nil {
		return nil, errors.New("invalid pod list")
	}

	statusMap := make(map[string]corev1.PodPhase)
	for _, pod := range pods {
		statusMap[pod.Name] = pod.Status.Phase
	}

	return statusMap, nil
}

// --- Archive ---
func (t *TestClient) GetPodOnNode(nodeName string, podName string, namespace string) (*corev1.Pod, error) {
	pods, err := t.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == nodeName && pod.Name == podName {
			return &pod, nil
		}
	}

	return nil, errors.New("pod not found")
}
