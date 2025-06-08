package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	testclient "kubevirt-nvidia-device-plugin/tests/testenv"
	"log"
)

var _ = Describe("GPU Device Plugin Test", Ordered, func() {
	var client *testclient.TestClient

	// Create test client before executing the tests
	BeforeAll(func() {
		var err error
		client, err = testclient.GetNewClientTest()
		Expect(err).ToNot(HaveOccurred(), "Failed to create test client")
	})

	var _ = Describe("Device Plugin Setup Validation", func() {
		Context("Test Device Plugin Deployment", func() {
			When("Deploying device plugin", func() {
				It("Should be running on each worker node", func() {
					pods, err := client.GetPodsList(client.Config.DevicePluginName, client.Config.DevicePluginNamespace)
					Expect(err).ToNot(HaveOccurred())
					Expect(pods).ToNot(BeEmpty(), "No device plugin pods found")

					workerNodes, err := client.GetWorkerNodes()
					Expect(err).ToNot(HaveOccurred())
					Expect(len(workerNodes)).To(Equal(len(pods)))

					statusMap, err := client.GetPodsStatusMap(pods)
					Expect(err).ToNot(HaveOccurred())

					for podName, podStatus := range statusMap {
						// TODO: delete later
						log.Printf("Pod Name: %s, Status: %s", podName, podStatus)
						Expect(podStatus).To(Equal(corev1.PodRunning), fmt.Sprintf("pod %s is %s", podName, podStatus))
					}
				})
			})
		})

		Context("Test Device Allocation", func() {
			It("Should allocate specified devices", func() {
				nodesToCheck := client.Config.Nodes
				if nodesToCheck == nil || len(nodesToCheck) == 0 {
					Skip("User did not provide nodes and devices to check")
				}

				for _, node := range nodesToCheck {
					for _, dev := range node.Devices {
						quantity, err := client.GetAllocatableDeviceQuantity(node.Name, dev.Name)
						Expect(err).ToNot(HaveOccurred())
						Expect(quantity).Should(BeEquivalentTo(dev.Number))
					}
				}

			})
		})
	})
})
