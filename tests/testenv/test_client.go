package testenv

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
	"strings"
)

const (
	LabelMaster       = "node-role.kubernetes.io/master"
	LabelControlPlane = "node-role.kubernetes.io/control-plane"
)

type TestClient struct {
	ClientSet      kubernetes.Interface
	KubeVirtClient kubecli.KubevirtClient
	Config         *TestConfig
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
	Name     string `yaml:"name"`
	Number   string `yaml:"number"`
	Allocate bool   `yaml:"allocate"`
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
		if _, isMaster := labels[LabelMaster]; isMaster {
			continue
		}
		if _, isControlPlane := labels[LabelControlPlane]; isControlPlane {
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

// TODO: make one function
func (t *TestClient) GetDeviceCapacity(nodeName string, deviceName string) (string, error) {
	node, err := t.GetNode(nodeName)
	if err != nil {
		return "", err
	}
	quantity, exists := node.Status.Capacity[corev1.ResourceName(deviceName)]
	if !exists {
		return "", fmt.Errorf("device %s not found in 'Capacity' section of node %s", deviceName, nodeName)
	}

	return quantity.String(), nil
}

func (t *TestClient) GetAllocatableDeviceQuantity(nodeName string, deviceName string) (string, error) {
	node, err := t.GetNode(nodeName)
	if err != nil {
		return "", err
	}
	quantity, exists := node.Status.Allocatable[corev1.ResourceName(deviceName)]
	if !exists {
		return "", fmt.Errorf("device %s not found in allocatable resources for node %s", deviceName, nodeName)
	}

	return quantity.String(), nil
}

// ---- end TODO

func (t *TestClient) GetPodOnNode(nodeName string, podNamePrefix string, namespace string) (*corev1.Pod, error) {
	pods, err := t.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == nodeName && strings.HasPrefix(pod.Name, podNamePrefix) {
			return &pod, nil
		}
	}

	return nil, fmt.Errorf(fmt.Sprintf("pod \"%s\" on node \"%s\" not found", podNamePrefix, nodeName))
}

func (t *TestClient) GetPodsList(prefix string, namespace string) ([]corev1.Pod, error) {
	if prefix == "" || namespace == "" {
		return nil, fmt.Errorf("prefix or namespace is empty")
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
		return nil, fmt.Errorf("invalid pod list")
	}

	statusMap := make(map[string]corev1.PodPhase)
	for _, pod := range pods {
		statusMap[pod.Name] = pod.Status.Phase
	}

	return statusMap, nil
}

func (t *TestClient) GetVirtualMachine() *v1.VirtualMachine {
	var gpus []v1.GPU

	// Get devices to allocate
	gpu_index := 0
	for _, node := range t.Config.Nodes {
		for _, device := range node.Devices {
			if device.Allocate {
				// One of each device is enough
				gpus = append(gpus, v1.GPU{
					Name:       fmt.Sprintf("GPU-%d", gpu_index),
					DeviceName: device.Name,
				})
				gpu_index++
			}
		}
	}

	vm := &v1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gpu-test-vm",
			Namespace: "default", // TODO: maybe change?
		},
		Spec: v1.VirtualMachineSpec{
			Template: &v1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubevirt.io/domain": "gpu-test-vm",
					},
				},
				Spec: v1.VirtualMachineInstanceSpec{
					Domain: v1.DomainSpec{
						Resources: v1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
								corev1.ResourceMemory: *resource.NewQuantity(4*1024*1024*1024, resource.BinarySI), // 4Gi
							},
						},
						//Devices: v1.Devices{
						//	GPUs: []v1.GPU{
						//		{
						//			Name:       "nvidia-NVSwitch",
						//			DeviceName: "nvidia.com/GH100_H100_NVSwitch",
						//		},
						//		{
						//			Name:       "nvidia-GPU",
						//			DeviceName: "nvidia.com/GH100_H100_SXM5_80GB",
						//		},
						//	},
						//},
					},
					Volumes: []v1.Volume{
						{
							Name: "containerdisk",
							VolumeSource: v1.VolumeSource{
								ContainerDisk: &v1.ContainerDiskSource{
									Image: "quay.io/kubevirt/cirros-container-disk-demo",
								},
							},
						},
					},
				},
			},
		},
	}

	if len(gpus) > 0 {
		vm.Spec.Template.Spec.Domain.Devices.GPUs = gpus
	}

	return vm
}
