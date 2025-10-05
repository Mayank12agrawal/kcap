package cmd

import (
    "context"
    "fmt"
    "os"
    "sort"
    "strconv"
    "time"

    "github.com/spf13/cobra"
    "github.com/jedib0t/go-pretty/v6/table"
    "kcap/pkg/analysis"
    "kcap/pkg/k8s"
)

var nodesCmd = &cobra.Command{
    Use:   "nodes",
    Short: "Show per-node allocatable, requested and usage summary",
    Run: func(cmd *cobra.Command, args []string) {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        kube, err := k8s.NewK8sClientWithConfig(flagKubeconfig)
        if err != nil {
            fmt.Println("Error creating kube client:", err)
            os.Exit(1)
        }

        nodes, err := kube.ListNodes(ctx)
        if err != nil {
            fmt.Println("Error listing nodes:", err)
            os.Exit(1)
        }

        nodeMetrics, err := kube.NodeMetrics(ctx)
        if err != nil {
            fmt.Println("Warning: Metrics-server not available, usage values will be zero")
            nodeMetrics = nil
        }

        pods, err := kube.ListPods(ctx, flagNamespace)
        if err != nil {
            fmt.Println("Error listing pods:", err)
            os.Exit(1)
        }

        stats := analysis.NodeStats(nodes, nodeMetrics, pods)
        sort.Slice(stats, func(i, j int) bool {
            return stats[i].Name < stats[j].Name
        })

        if flagJSON {
            printJSON(stats)
            return
        }

        t := table.NewWriter()
        t.SetOutputMirror(os.Stdout)
        t.AppendHeader(table.Row{"NODE", "CPU(Alloc/Req/Use m)", "MEM(Alloc/Req/Use Mi)", "WORKLOADPODS", "STATUS"})

        for _, s := range stats {
            cpuField := fmt.Sprintf("%d / %d / %d", s.CPUAllocMilli, s.CPUReqMilli, s.CPUUsedMilli)
            memField := fmt.Sprintf("%d / %d / %d", s.MemAllocMi, s.MemReqMi, s.MemUsedMi)
            t.AppendRow(table.Row{s.Name, cpuField, memField, strconv.Itoa(s.UserPodCount), s.Status})
        }
        t.Render()
    },
}

func init() {
    nodesCmd.Flags().StringVar(&flagKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
    nodesCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "", "Namespace")
    nodesCmd.Flags().BoolVar(&flagJSON, "json", false, "Print output as JSON")
}
