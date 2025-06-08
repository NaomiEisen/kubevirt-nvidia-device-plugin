package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	// This is probably wrong
	testclient "kubevirt-nvidia-device-plugin/tests/testenv"
)

var _ = Describe("GPU Device Plugin Test", Ordered, func() {
	var client *testclient.TestClient

	// Create test client before executing the tests
	BeforeAll(func() {
		var err error
		client, err = testclient.GetNewClientTest()
		Expect(err).ToNot(HaveOccurred(), "Failed to create test client")
	})

	var _ = Describe("Setup Validation", func() {
		Context("Test Device Plugin Deployment", func() {
			When("Deploying device plugin", func() {
				It("Should be running on each worker node", func() {
					// TODO: maybe should explicitly check each node for one corresponding pod
					validateNumPods(client)
					validatePodsStatus(client)
				})
			})
		})

		Context("Test Device Allocation", func() {
			It("Should allocate specified devices", func() {
				if client.Config.Nodes == nil || len(client.Config.Nodes) == 0 {
					Skip("User did not provide nodes and devices to check")
				}
				validateAllocatableDevicesQuantity(client)
			})
		})
	})

	var _ = Describe("GPU Device Plugin Functional Test", func() {
		When("Creating a new VM with GPUs", func() {
			It("Should be running", func() {
				// Create VM

				// Check status
			})
		})

	})
})

// Helper function - gets pods safely with error assertion
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

func validatePodsStatus(client *testclient.TestClient) {
	pods := getPods(client)
	podsStatusMap, err := client.GetPodsStatusMap(pods)
	Expect(err).ToNot(HaveOccurred())

	for podName, podStatus := range podsStatusMap {
		// TODO: delete later
		//log.Printf("Pod Name: %s, Status: %s", podName, podStatus)
		Expect(podStatus).To(Equal(corev1.PodRunning),
			fmt.Sprintf("pod %s is %s", podName, podStatus))
	}
}

func validateAllocatableDevicesQuantity(client *testclient.TestClient) {
	nodesToCheck := client.Config.Nodes
	for _, node := range nodesToCheck {
		for _, dev := range node.Devices {
			quantity, err := client.GetAllocatableDeviceQuantity(node.Name, dev.Name)
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get allocatable device %s", dev.Name))
			Expect(quantity).To(BeEquivalentTo(dev.Number),
				fmt.Sprintf("Number of device %s is incorrect.", dev.Name))
		}
	}
}

func createVirtualMachine(client *testclient.TestClient) {

}
