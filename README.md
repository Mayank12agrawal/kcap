kcap - Kubernetes Capacity and Resource Analyzer
kcap is a command-line tool designed to analyze Kubernetes cluster resource utilization in real-time, providing insights and actionable recommendations for CPU and memory right-sizing at node, pod, and deployment levels.

Features
Node summary: Displays allocatable, requested, and usage metrics of cluster nodes with health status and scale-in candidate identification.

Pod summary: Shows pod-level CPU and memory requests vs usage, including waste percentages.

Deployment summary: Aggregates pod metrics by deployment to identify over-provisioned deployments.

Resource recommendations: Suggests nodes to drain (scale-in candidates) and pods for resource request reduction with configurable thresholds.

JSON output: Machine-readable output for integration into automation pipelines.

namespace filtering with shorthand -n flag similar to kubectl.

Ignores DaemonSet pods by default to focus on user workloads.

Installation
Build from source:
```
go build -o kcap
```
Ensure you have a working Kubernetes cluster with Metrics Server installed.
Usage
Use --help for detailed flags for each command.

Nodes
Show node resource summary and status:
```
kcap nodes [-n <namespace>] [--kubeconfig <path>] [--json]
```
Note: Nodes are cluster-wide; namespace flag filters pods for usage computations.

Pods
Show pod CPU/memory requests vs usage with waste %:
```
kcap pods -n <namespace> [--kubeconfig <path>] [--json]
```
Deployments
Aggregated deployment resource usage and waste summary:
```
kcap deploys -n <namespace> [--kubeconfig <path>] [--json]
```
Recommendations
Suggest resource right-sizing and node scale-in candidates. Customize threshold:
```
kcap recommend -n <namespace> [--kubeconfig <path>] [--json] [--threshold <waste_percentage>]
```
Default threshold is 80%.

Report
Comprehensive cluster or namespace-level summary report:
```
kcap report -n <namespace> [--kubeconfig <path>] [--json] [--threshold <waste_percentage>]
```
Concepts
CPU m (millicores): 1000m = 1 CPU core.

Memory Mi (Mebibytes): 1 Mi = 1,048,576 bytes.

Resource requests must be set in pod specs to avoid zero request readings.

Recommendations are based on recent Metrics Server data (~1-minute average).

Right-sizing suggestions apply a configurable buffer (default 30-70%) above current usage.

Limitations
Metrics come from Metrics Server, representing short-term usage snapshots.

Usage spikes outside the sampling window may not be captured.

Recommendations are guidelines and should be validated with longer-term monitoring.

DaemonSet pods are excluded from pod and deployment analysis to reduce noise.

Best Practices
Always set resource requests and limits on your pods for effective scheduling and resource management.

Use kcap recommendations as part of a broader capacity planning and cost optimization strategy.

Combine with Prometheus or other monitoring solutions for historical usage insights.

Example
```
kcap recommend -n default --threshold 80
```
Produces output like:
TYPE                |  DETAILS             |  SUGGESTION                 
--------------------+----------------------+-----------------------------
Scale-in candidate  |  ip-10-50-8-114.ec2  |  Consider draining this node
Right-size pod      |  default/myapp-1     |  Reduce CPU request         
Right-size pod      |  default/myapp-1     |  Reduce Memory request      

