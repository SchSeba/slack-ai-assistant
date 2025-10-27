# SR-IOV Configuration for OpenShift 4.16

## Overview

SR-IOV (Single Root I/O Virtualization) allows a PCIe device to present itself as multiple devices, improving performance and resource utilization.

## Installation

To install the SR-IOV Network Operator on OpenShift 4.16:

1. Navigate to OperatorHub in the OpenShift Console
2. Search for "SR-IOV Network Operator"
3. Click Install and follow the prompts

## Configuration

Create an `SriovNetworkNodePolicy` to configure SR-IOV devices:

```yaml
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovNetworkNodePolicy
metadata:
  name: policy-1
  namespace: openshift-sriov-network-operator
spec:
  nodeSelector:
    feature.node.kubernetes.io/network-sriov.capable: "true"
  numVfs: 8
  nicSelector:
    vendor: "8086"
    deviceID: "1572"
  resourceName: intelnics
```

## Troubleshooting

- Check node labels: `oc get nodes --show-labels`
- View operator logs: `oc logs -n openshift-sriov-network-operator deployment/sriov-network-operator`
- Verify device availability: `oc get sriovnetworknodestates -n openshift-sriov-network-operator`

## Performance Tuning

For optimal performance with SR-IOV on OpenShift 4.16:

1. Enable CPU pinning
2. Configure hugepages
3. Use DPDK when appropriate


