package e2e

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	v1 "kubevirt.io/api/core/v1"
	"time"

	// This is probably wrong
	testclient "kubevirt-nvidia-device-plugin/tests/testenv"
)

const (
	WaitingTimeout time.Duration = 1 * time.Minute
	CheckInterval  time.Duration = 2 * time.Second
)

var _ = Describe("GPU Device Plugin Test", Ordered, func() {
	var client *testclient.TestClient
	var vm v1.VirtualMachine

	// Create test client before executing the tests
	BeforeAll(func() {
		var err error
		client, err = testclient.GetNewClientTest()
		Expect(err).ToNot(HaveOccurred(), "Failed to create test client")
	})

	var _ = Describe("Setup Validation", func() {
		Context("Device Plugin Deployment", func() {
			When("Deploying device plugin", func() {
				It("is present on each worker node", func() {
					// TODO: Choose one of them
					validateNumPods(client)
					validateDPOnEachNode(client)
				})
				It("has all pods in the RUNNING state", func() {
					validatePodsStatus(client)
				})
			})
		})

		Context("Device allocation", func() {
			It("should allocate the specified devices", func() {
				if client.Config.Nodes == nil || len(client.Config.Nodes) == 0 {
					Skip("Skipping device allocation check: no nodes or devices specified")
				}
				validateDevicesCapacity(client)
			})
		})
	})

	var _ = Describe("GPU Device Plugin Functional Test", func() {
		When("Creating a new VM with GPUs", func() {
			It("Should successfully create new VM with GPUs", func() {
				vm = createVirtualMachine(client)
			})

			It("Should be running", func() {
				validateRunningVM(client, vm)
			})

			It("Should be successfully deleted", func() {
				deleteVirtualMachine(client, vm)
			})
		})

	})
})

// Helper function – safely gets pods with error assertion
func getPods(client *testclient.TestClient) []corev1.Pod {
	pods, err := client.GetPodsList(client.Config.DevicePluginName, client.Config.DevicePluginNamespace)
	Expect(err).ToNot(HaveOccurred())
	Expect(pods).ToNot(BeEmpty(),
		fmt.Sprintf("No device plugin pods with the name \"%s\" are found in namespace \"%s\"",
			client.Config.DevicePluginName, client.Config.DevicePluginNamespace))
	return pods
}

func validateNumPods(client *testclient.TestClient) {
	pods := getPods(client)

	workerNodes, err := client.GetWorkerNodes()
	Expect(err).ToNot(HaveOccurred())
	Expect(len(workerNodes)).To(Equal(len(pods)),
		"Number of device plugin pods is not aligned with the number of available worker nodes")
}

func validateDPOnEachNode(client *testclient.TestClient) {
	workerNodes, err := client.GetWorkerNodes()
	Expect(err).ToNot(HaveOccurred())

	for _, node := range workerNodes {
		_, err = client.GetPodOnNode(node.Name, client.Config.DevicePluginName, client.Config.DevicePluginNamespace)
		Expect(err).ToNot(HaveOccurred(),
			fmt.Sprintf("Failed to get pod with prefix \"%s\" on node \"%s\"", client.Config.DevicePluginName, node.Name))
	}

}

func validatePodsStatus(client *testclient.TestClient) {
	pods := getPods(client)
	podsStatusMap, err := client.GetPodsStatusMap(pods)
	Expect(err).ToNot(HaveOccurred())

	for podName, podStatus := range podsStatusMap {
		Expect(podStatus).To(Equal(corev1.PodRunning),
			fmt.Sprintf("pod %s is %s", podName, podStatus))
	}
}

func validateDevicesCapacity(client *testclient.TestClient) {
	nodesToCheck := client.Config.Nodes
	for _, node := range nodesToCheck {
		for _, dev := range node.Devices {
			quantity, err := client.GetDeviceCapacity(node.Name, dev.Name)
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get allocatable device %s", dev.Name))
			Expect(quantity).To(BeEquivalentTo(dev.Number),
				fmt.Sprintf("Number of device %s is incorrect.", dev.Name))
		}
	}
}

func createVirtualMachine(client *testclient.TestClient) v1.VirtualMachine {
	vm := client.GetVirtualMachine()

	// Set vm to run
	vm.Spec.RunStrategy = ptr.To(v1.RunStrategyAlways)

	vmInterface := client.KubeVirtClient.VirtualMachine("default")
	createdVM, err := vmInterface.Create(context.TODO(), vm, metav1.CreateOptions{})
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to create virtual machine %s", vm.Name))

	return *createdVM
}

func validateRunningVM(client *testclient.TestClient, vm v1.VirtualMachine) {
	var vmi *v1.VirtualMachineInstance
	var err error

	vmiInterface := client.KubeVirtClient.VirtualMachineInstance("default")

	Eventually(func() v1.VirtualMachineInstancePhase {
		vmi, err = vmiInterface.Get(context.TODO(), vm.Name, metav1.GetOptions{})
		if err != nil {
			return ""
		}
		return vmi.Status.Phase
	}, WaitingTimeout, CheckInterval).Should(Equal(v1.Running))

}

func deleteVirtualMachine(client *testclient.TestClient, vm v1.VirtualMachine) {
	vmInterface := client.KubeVirtClient.VirtualMachine(vm.Namespace)

	err := vmInterface.Delete(context.TODO(), vm.Name, metav1.DeleteOptions{})
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to delete virtual machine %s", vm.Name))
}
