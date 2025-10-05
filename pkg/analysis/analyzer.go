package analysis

import (
    "fmt"
    "strings"

    v1 "k8s.io/api/core/v1"
)

const (
    CPUScaleInThresholdPercent = 30.0 // Scale-in candidate threshold (%)
    MemScaleInThresholdPercent = 30.0
)

type NodeStat struct {
    Name          string
    CPUAllocMilli int64
    CPUReqMilli   int64
    CPUUsedMilli  int64
    MemAllocMi    int64
    MemReqMi      int64
    MemUsedMi     int64
    UserPodCount  int
    Status        string
}

type DeploymentStat struct {
    Namespace    string
    Name         string
    CPUReqMilli  int64
    CPUUsedMilli int64
    MemReqMi     int64
    MemUsedMi    int64
    PodCount     int
    WasteCPU     float64
    WasteMem     float64
}

type Recommendation struct {
    Type       string
    Details    string
    Suggestion string
    Severity   string
}

type PodRecord struct {
    Namespace    string
    Name         string
    NodeName     string
    CPUReqMilli  int64
    CPUUsedMilli int64
    MemReqMi     int64
    MemUsedMi    int64
    Owner        string
    Deployment   string
    IsDaemonSet  bool
}

func ResolveDeploymentName(pod v1.Pod) string {
    for _, ownerRef := range pod.OwnerReferences {
        if ownerRef.Kind == "Deployment" {
            return ownerRef.Name
        }
        if ownerRef.Kind == "ReplicaSet" {
            rsName := ownerRef.Name
            idx := strings.LastIndex(rsName, "-")
            if idx > 0 {
                return rsName[:idx]
            }
            return rsName
        }
    }
    if appName, ok := pod.Labels["app.kubernetes.io/name"]; ok {
        return appName
    }
    if app, ok := pod.Labels["app"]; ok {
        return app
    }
    return pod.Name
}

func getNodeCondition(conditions []v1.NodeCondition, condType v1.NodeConditionType) *v1.NodeCondition {
    for i, condition := range conditions {
        if condition.Type == condType {
            return &conditions[i]
        }
    }
    return nil
}

func NodeStats(nodes []v1.Node, nodeMetrics map[string]v1.ResourceList, pods []v1.Pod) []NodeStat {
    var stats []NodeStat

    for _, n := range nodes {
        readyCondition := getNodeCondition(n.Status.Conditions, v1.NodeReady)
        status := "Unknown"
        cpuAlloc := n.Status.Allocatable.Cpu().MilliValue()
        memAlloc := n.Status.Allocatable.Memory().Value() / 1024 / 1024

        var cpuUsed int64 = 0
        var memUsed int64 = 0
        if usage, ok := nodeMetrics[n.Name]; ok {
            cpuUsed = usage.Cpu().MilliValue()
            memUsed = usage.Memory().Value() / 1024 / 1024
        }

        cpuUsagePercent := 0.0
        memUsagePercent := 0.0
        if cpuAlloc > 0 {
            cpuUsagePercent = float64(cpuUsed) / float64(cpuAlloc) * 100.0
        }
        if memAlloc > 0 {
            memUsagePercent = float64(memUsed) / float64(memAlloc) * 100.0
        }

        if readyCondition != nil && readyCondition.Status == v1.ConditionTrue {
            if cpuUsagePercent < CPUScaleInThresholdPercent && memUsagePercent < MemScaleInThresholdPercent {
                status = "Scale-in candidate"
            } else {
                status = "Healthy"
            }
        } else {
            status = "NotReady"
        }

        var cpuReqTotal int64 = 0
        var memReqTotal int64 = 0
        podCount := 0

        for _, pod := range pods {
            if pod.Spec.NodeName == n.Name {
                podCount++
                for _, c := range pod.Spec.Containers {
                    cpuReqTotal += c.Resources.Requests.Cpu().MilliValue()
                    memReqTotal += c.Resources.Requests.Memory().Value() / (1024 * 1024)
                }
            }
        }

        stats = append(stats, NodeStat{
            Name:          n.Name,
            Status:        status,
            CPUAllocMilli: cpuAlloc,
            CPUReqMilli:   cpuReqTotal,
            CPUUsedMilli:  cpuUsed,
            MemAllocMi:    memAlloc,
            MemReqMi:      memReqTotal,
            MemUsedMi:     memUsed,
            UserPodCount:  podCount,
        })
    }
    return stats
}

func PodRecords(pods []v1.Pod, podMetrics map[string]v1.ResourceList, filter string) []PodRecord {
    var records []PodRecord
    for _, p := range pods {
        isDaemon := false
        for _, ownerRef := range p.OwnerReferences {
            if ownerRef.Kind == "DaemonSet" {
                isDaemon = true
                break
            }
        }
        if isDaemon {
            continue // Ignore DaemonSet pods
        }

        owner := "None"
        for _, ownerRef := range p.OwnerReferences {
            owner = ownerRef.Kind
            break
        }
        cpuReq := int64(0)
        memReq := int64(0)
        for _, c := range p.Spec.Containers {
            if r := c.Resources.Requests.Cpu(); r != nil {
                cpuReq += r.MilliValue()
            }
            if r := c.Resources.Requests.Memory(); r != nil {
                memReq += r.Value() / 1024 / 1024
            }
        }

        var cpuUsed int64 = 0
        var memUsed int64 = 0
        if usage, ok := podMetrics[p.Name]; ok {
            cpuUsed = usage.Cpu().MilliValue()
            memUsed = usage.Memory().Value() / 1024 / 1024
        }

        records = append(records, PodRecord{
            Namespace:    p.Namespace,
            Name:         p.Name,
            NodeName:     p.Spec.NodeName,
            CPUReqMilli:  cpuReq,
            CPUUsedMilli: cpuUsed,
            MemReqMi:     memReq,
            MemUsedMi:    memUsed,
            Owner:        owner,
            Deployment:   ResolveDeploymentName(p),
            IsDaemonSet:  isDaemon,
        })
    }
    return records
}

func DeploymentAggregation(pods []PodRecord) []DeploymentStat {
    m := make(map[string]*DeploymentStat)
    for _, p := range pods {
        key := p.Namespace + "/" + p.Deployment
        d, ok := m[key]
        if !ok {
            d = &DeploymentStat{
                Namespace: p.Namespace,
                Name:      p.Deployment,
            }
            m[key] = d
        }
        d.PodCount++
        d.CPUReqMilli += p.CPUReqMilli
        d.CPUUsedMilli += p.CPUUsedMilli
        d.MemReqMi += p.MemReqMi
        d.MemUsedMi += p.MemUsedMi
    }
    var deployments []DeploymentStat
    for _, d := range m {
        if d.CPUReqMilli > 0 {
            d.WasteCPU = (1.0 - float64(d.CPUUsedMilli)/float64(d.CPUReqMilli)) * 100
        }
        if d.MemReqMi > 0 {
            d.WasteMem = (1.0 - float64(d.MemUsedMi)/float64(d.MemReqMi)) * 100
        }
        deployments = append(deployments, *d)
    }
    return deployments
}

func severityLevel(waste float64) string {
    switch {
    case waste >= 90:
        return "High"
    case waste >= 70:
        return "Medium"
    case waste >= 50:
        return "Low"
    default:
        return "Info"
    }
}

func RecommendNodes(nodes []NodeStat) []Recommendation {
    var recs []Recommendation
    for _, n := range nodes {
        switch n.Status {
        case "Scale-in candidate":
            recs = append(recs, Recommendation{
                Type:       "Scale-in candidate",
                Details:    n.Name,
                Suggestion: "Consider draining this node",
                Severity:   "Medium",
            })
        case "NotReady":
            recs = append(recs, Recommendation{
                Type:       "Node",
                Details:    n.Name + " is NotReady",
                Suggestion: "Check node health and connectivity",
                Severity:   "High",
            })
        }
    }
    return recs
}

func RecommendPods(pods []PodRecord, threshold float64) []Recommendation {
    var recs []Recommendation
    for _, p := range pods {
        if p.CPUReqMilli > 0 {
            cpuWaste := 100 * (1.0 - float64(p.CPUUsedMilli)/float64(p.CPUReqMilli))
            if cpuWaste >= threshold {
                recs = append(recs, Recommendation{
                    Type:       "Pod (CPU)",
                    Details:    fmt.Sprintf("%s/%s", p.Namespace, p.Name),
                    Suggestion: "Consider reducing CPU requests",
                    Severity:   severityLevel(cpuWaste),
                })
            }
        }
        if p.MemReqMi > 0 {
            memWaste := 100 * (1.0 - float64(p.MemUsedMi)/float64(p.MemReqMi))
            if memWaste >= threshold {
                recs = append(recs, Recommendation{
                    Type:       "Pod (Memory)",
                    Details:    fmt.Sprintf("%s/%s", p.Namespace, p.Name),
                    Suggestion: "Consider reducing Memory requests",
                    Severity:   severityLevel(memWaste),
                })
            }
        }
    }
    return recs
}


