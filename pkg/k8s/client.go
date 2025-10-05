package k8s

import (
    "context"
    "os"
    "path/filepath"

    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    metrics "k8s.io/metrics/pkg/client/clientset/versioned"
    metricsv "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type K8sClient struct {
    Clientset     *kubernetes.Clientset
    MetricsClient *metrics.Clientset
}

// NewK8sClientWithConfig creates a Kubernetes clientset and a Metrics client,
// loading configuration from a kubeconfig path or in-cluster config.
func NewK8sClientWithConfig(kubeconfig string) (*K8sClient, error) {
    var cfg *rest.Config
    var err error

    if kubeconfig == "" {
        kubeconfig = os.Getenv("KUBECONFIG")
    }

    if kubeconfig == "" {
        if home, err := os.UserHomeDir(); err == nil {
            candidate := filepath.Join(home, ".kube", "config")
            if _, err := os.Stat(candidate); err == nil {
                kubeconfig = candidate
            }
        }
    }

    if kubeconfig != "" {
        cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            return nil, err
        }
    } else {
        cfg, err = rest.InClusterConfig()
        if err != nil {
            return nil, err
        }
    }

    clientset, err := kubernetes.NewForConfig(cfg)
    if err != nil {
        return nil, err
    }

    metricsClient, err := metrics.NewForConfig(cfg)
    if err != nil {
        return nil, err
    }

    return &K8sClient{Clientset: clientset, MetricsClient: metricsClient}, nil
}

// ListNodes lists all nodes in the cluster.
func (k *K8sClient) ListNodes(ctx context.Context) ([]v1.Node, error) {
    nodes, err := k.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    return nodes.Items, nil
}

// ListPods lists all pods in a given namespace. Passing empty string lists all pods.
func (k *K8sClient) ListPods(ctx context.Context, namespace string) ([]v1.Pod, error) {
    pods, err := k.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    return pods.Items, nil
}

// NodeMetrics fetches metrics usage for all nodes, keyed by node name.
func (k *K8sClient) NodeMetrics(ctx context.Context) (map[string]v1.ResourceList, error) {
    nodeMetricsList, err := k.MetricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    metricsMap := make(map[string]v1.ResourceList)
    for _, nm := range nodeMetricsList.Items {
        metricsMap[nm.Name] = nm.Usage
    }
    return metricsMap, nil
}

// PodMetrics fetches metrics usage for pods in the given namespace, keyed by pod name.
func (k *K8sClient) PodMetrics(ctx context.Context, namespace string) (map[string]v1.ResourceList, error) {
    podMetricsList, err := k.MetricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    metricsMap := make(map[string]v1.ResourceList)
    for _, pm := range podMetricsList.Items {
        metricsMap[pm.Name] = aggregatePodContainerUsage(pm)
    }
    return metricsMap, nil
}

// aggregatePodContainerUsage sums CPU and memory usage of all containers in a pod metrics item.
func aggregatePodContainerUsage(pm metricsv.PodMetrics) v1.ResourceList {
    cpuTotal := int64(0)
    memTotal := int64(0)
    for _, c := range pm.Containers {
        cpuTotal += c.Usage.Cpu().MilliValue()
        memTotal += c.Usage.Memory().Value()
    }
    return v1.ResourceList{
        v1.ResourceCPU:    *resource.NewMilliQuantity(cpuTotal, resource.DecimalSI),
        v1.ResourceMemory: *resource.NewQuantity(memTotal, resource.BinarySI),
    }
}
