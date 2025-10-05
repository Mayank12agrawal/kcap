# kcap - Kubernetes Capacity & Resource Analyzer 🚀

**kcap** is a CLI tool to analyze Kubernetes resource usage in real-time and provide actionable CPU/memory right-sizing recommendations for **nodes**, **pods**, and **deployments**.

---

## Features ⚙️

* 🖥️ Node, Pod, and Deployment resource summaries
* 🎯 Resource right-sizing recommendations
* 📉 Node scale-in candidate identification
* 📄 JSON output for automation
* 🏷️ Namespace filtering (`-n`)
* 🚫 Ignores DaemonSet pods by default
* ⏱️ Real-time metrics (~1-min average)
* 💡 Actionable insights for cost optimization

---

## Installation 🛠️

### 📦 One-line Install

```bash
curl -fsSL https://raw.githubusercontent.com/Mayank12agrawal/kcap/main/install.sh | sh
```

### 🧰 Build from Source

```bash
go build -o kcap
```

**Prerequisites**: Kubernetes cluster with Metrics Server.

---

## Usage 📋

### Nodes

```bash
kcap nodes [-n <namespace>] [--kubeconfig <path>] [--json]
```

### Pods

```bash
kcap pods -n <namespace> [--kubeconfig <path>] [--json]
```

### Deployments

```bash
kcap deploys -n <namespace> [--kubeconfig <path>] [--json]
```

### Recommendations

```bash
kcap recommend -n <namespace> [--json] [--threshold <waste_percentage>]
```

*Default threshold: 80%*

### Report

```bash
kcap report -n <namespace> [--json] [--threshold <waste_percentage>]
```

---

## Concepts 🧠

| Concept  | Explanation            |
| -------- | ---------------------- |
| CPUm     | 1000m = 1 CPU core     |
| MemoryMi | 1 Mi = 1,048,576 bytes |

> Resource requests must be set in pod specs. Recommendations are based on ~1-min Metrics Server snapshots.

---

## Limitations ⚠️

* 📊 Short-term usage snapshots only
* ⚡ Usage spikes outside sampling window may be missed
* ⚠️ Recommendations are **guidelines**, validate with long-term monitoring

---

## Best Practices ✅

* 📝 Set resource requests and limits on pods
* 📈 Use alongside Prometheus or other monitoring tools for historical insights
* 🔄 Run kcap regularly for proactive optimization
* 💰 Use recommendations for cost-saving strategies

---

## Example 💡

```bash
kcap recommend -n default --threshold 80
```

**Output Sample:**

| TYPE               | DETAILS            | SUGGESTION                  |
| ------------------ | ------------------ | --------------------------- |
| Scale-in candidate | ip-10-50-8-114.ec2 | Consider draining this node |
| Right-size pod     | default/myapp-1    | Reduce CPU request          |
| Right-size pod     | default/myapp-1    | Reduce Memory request       |

---

## Additional Tips ✨

* 🤖 Combine kcap with CI/CD pipelines for automated alerts
* 📊 Use `--json` output for custom dashboards
* 🎯 Filter by namespace for focused analysis
* 🧪 Always validate changes in staging before production

## Feedback & Contributing 🤝

Have an idea, found a bug, or want to contribute?  
Open an [issue](https://github.com/Mayank12agrawal/kcap/issues) or [pull request](https://github.com/Mayank12agrawal/kcap/pulls) on the repository.