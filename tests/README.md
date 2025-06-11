# e2e test - device plugin
This test provides end-to-end (E2E) behavior tests for the NVIDIA Device Plugin in a Kubernetes or OpenShift cluster.
The test assumes that the NVIDIA Device Plugin is already deployed and running in your cluster.

Validates-

1. The NVIDIA Device Plugin Pod is running on every worker node.
2. If specified in the provided test configuration:
    - Confirms the quantity of GPUs available on each node.
    - Launches a test Virtual Machine (VM) requesting the specified GPUs.
    - Verifies that the VM successfully transitions to the Running phase with assigned GPUs.

How to use
--
To run the test, you must provide the following:

### Test Configuration File

Use the provided `test_config.yaml` as a template.

#### Required Fields:
```yaml
deviceplugin_name: <the device plugin name>
deviceplugin_namespace: <the device plugin namespace>
```
#### Optional Fields:
Specify nodes and expected GPU devices if you want to validate GPU availability:

```yaml
nodes:
- name: <node name>
  devices:
    - name: <device name>
      number: <expected quantity>
```

You can specify the config file path with the -test-config flag:
```cmd
  -test-config <path to test-config file>
```
If not specified, it defaults to the example path.

```TODO: decide what to do with allocate!```

### Kubeconfig file

The test uses your kubeconfig file to access the cluster.
Default path: `$HOME/.kube/config`
To override, use the -kubeconfig flag:

 ```cmd
  -kubeconfig <path to kubeconfig>
```

## Running Tests

To run tests, run the following command from `e2e` directory.

```bash
  cd tests/e2e
  ginkgo run
```
or

```bash
  cd tests/e2e
  go test
```
