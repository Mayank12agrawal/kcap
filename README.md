
# 🌱 kcap - Kubernetes Capacity & Resource Analyzer

![Go Version](https://img.shields.io/badge/Go-1.20+-blue)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Compatible-326CE5?logo=kubernetes)
![License](https://img.shields.io/badge/License-MIT-yellow)

`kcap` is a lightweight CLI tool that analyzes **Kubernetes cluster resource utilization** in real-time and provides actionable recommendations for **CPU and memory right-sizing** across nodes, pods, and deployments.

---

## 🚀 Features

- 📊 **Node summary:** Shows allocatable, requested, and usage metrics with health status and scale-in candidate detection.
- 📦 **Pod summary:** Displays CPU and memory requests vs usage, including waste percentage.
- 🧬 **Deployment summary:** Aggregates pod metrics by deployment to highlight over-provisioned workloads.
- 🧠 **Resource recommendations:** Suggests nodes to drain and pods to right-size based on configurable thresholds.
- 📤 **JSON output:** Machine-readable for integration into automation pipelines.
- 🧹 **Namespace filtering:** Filter resources with `-n` flag like `kubectl`.
- 🚫 **DaemonSet exclusion:** Ignores DaemonSet pods by default to reduce noise.

---

## ⚙️ Installation

### 🌀 Quick Install (Recommended)

Run this one-liner to install the latest release:
```bash
curl -fsSL https://raw.githubusercontent.com/Mayank12agrawal/kcap/main/install.sh | bash
```

### 🛠️ Build from Source

```bash
git clone https://github.com/Mayank12agrawal/kcap.git
cd kcap
go build -o kcap
```

✅ **Prerequisites:**
- Go 1.20+  
- A running Kubernetes cluster  
- Metrics Server installed  

---

## 📘 Commands

### 🖥️ `kcap nodes`
Show cluster node resource summary and identify scale-in candidates.
```bash
kcap nodes [--kubeconfig <path>] [--json]
```
📌 Use `-n <namespace>` to filter pods for usage calculation.

### 📦 `kcap pods`
Display pod-level CPU and memory requests vs usage with waste percentage.
```bash
kcap pods -n <namespace> [--kubeconfig <path>] [--json]
```

### 🏗️ `kcap deploys`
Aggregate pod metrics by deployment to identify over-provisioned workloads.
```bash
kcap deploys -n <namespace> [--kubeconfig <path>] [--json]
```

### 🧠 `kcap recommend`
Suggest nodes for scale-in and pods for right-sizing based on a configurable threshold.
```bash
kcap recommend -n <namespace> [--threshold <waste_percentage>] [--json]
```
Default threshold: `80%`

📌 Example:
```bash
kcap recommend -n default --threshold 80
```

### 📊 `kcap report`
Generate a full summary of nodes, pods, deployments, and recommendations.
```bash
kcap report -n <namespace> [--threshold <waste_percentage>] [--json]
```

---

## 🧪 Example Workflow

```bash
# View node utilization
kcap nodes

# See pod-level resource usage in a namespace
kcap pods -n default

# Get right-sizing recommendations with an 80% threshold
kcap recommend -n default --threshold 80

# Generate a complete cluster report in JSON
kcap report -n default --threshold 80 --json
```

---

## 📈 Sample Recommendation Output

```text
TYPE                 | DETAILS                       | SUGGESTION
---------------------+------------------------------+-----------------------------
Scale-in candidate   | ip-10-50-8-114.ec2           | Consider draining this node
Right-size pod       | default/myapp-1              | Reduce CPU request
Right-size pod       | default/myapp-1              | Reduce Memory request
```

---

## 📌 Notes & Limitations

- Metrics are from **Metrics Server**, representing ~1-minute averages.  
- Usage spikes outside this window may not be captured.  
- Recommendations are **guidelines** — validate them with historical metrics.  
- DaemonSet pods are excluded from analysis.

---

## ✅ Best Practices

- Always define **resource requests and limits** on pods.
- Use `kcap` as part of a broader **capacity planning** and **cost optimization** strategy.
- Combine with Prometheus or other monitoring tools for long-term visibility.

---

## 🤝 Contributing

Found a bug 🐛 or have a feature idea 💡?  
We welcome contributions!

- Open an [issue](https://github.com/Mayank12agrawal/kcap/issues)
- Submit a [pull request](https://github.com/Mayank12agrawal/kcap/pulls)

---

⭐ **If you find `kcap` useful, give it a star on [GitHub](https://github.com/Mayank12agrawal/kcap)!**
