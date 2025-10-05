package cmd

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/spf13/cobra"
    "github.com/jedib0t/go-pretty/v6/table"
    "kcap/pkg/analysis"
    "kcap/pkg/k8s"
)

var reportCmd = &cobra.Command{
    Use:   "report",
    Short: "Full cluster summary including nodes, deployments, and recommendations",
    Run: func(cmd *cobra.Command, args []string) {
        ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
        defer cancel()

        kube, err := k8s.NewK8sClientWithConfig(flagKubeconfig)
        if err != nil {
            fmt.Println("Error creating kube client:", err)
            os.Exit(1)
        }

        nodes, _ := kube.ListNodes(ctx)
        nodeMetrics, _ := kube.NodeMetrics(ctx)
        pods, _ := kube.ListPods(ctx, flagNamespace)
        podMetrics, _ := kube.PodMetrics(ctx, flagNamespace)

        nodeStats := analysis.NodeStats(nodes, nodeMetrics, pods)
        podRecords := analysis.PodRecords(pods, podMetrics, "")
        deployStats := analysis.DeploymentAggregation(podRecords)
        nodeRecs := analysis.RecommendNodes(nodeStats)
        podRecs := analysis.RecommendPods(podRecords, flagThreshold)
        recs := append(nodeRecs, podRecs...)

        fmt.Println("Cluster Summary:")
        var totalCPUAlloc, totalCPUReq, totalCPUUse int64
        var totalMemAlloc, totalMemReq, totalMemUse int64
        for _, n := range nodeStats {
            totalCPUAlloc += n.CPUAllocMilli
            totalCPUReq += n.CPUReqMilli
            totalCPUUse += n.CPUUsedMilli
            totalMemAlloc += n.MemAllocMi
            totalMemReq += n.MemReqMi
            totalMemUse += n.MemUsedMi
        }
        fmt.Printf("CPU Alloc(m): %d  CPU Req(m): %d  CPU Used(m): %d\n", totalCPUAlloc, totalCPUReq, totalCPUUse)
        fmt.Printf("MEM Alloc(Mi): %d  MEM Req(Mi): %d  MEM Used(Mi): %d\n\n", totalMemAlloc, totalMemReq, totalMemUse)

        fmt.Println("Top Over-provisioned Deployments:")
        t := table.NewWriter()
        t.SetOutputMirror(os.Stdout)
        t.AppendHeader(table.Row{"DEPLOYMENT", "CPU(req/use m)", "MEM(req/use Mi)", "PODS", "WASTE% CPU"})
        for _, d := range deployStats {
            t.AppendRow(table.Row{
                d.Name,
                fmt.Sprintf("%d / %d", d.CPUReqMilli, d.CPUUsedMilli),
                fmt.Sprintf("%d / %d", d.MemReqMi, d.MemUsedMi),
                d.PodCount,
                fmt.Sprintf("%.1f", d.WasteCPU),
            })
        }
        t.Render()

        fmt.Println("\nRecommendations:")
        t2 := table.NewWriter()
        t2.SetOutputMirror(os.Stdout)
        t2.AppendHeader(table.Row{"TYPE", "DETAILS", "SUGGESTION"})
        for _, r := range recs {
            t2.AppendRow(table.Row{r.Type, r.Details, r.Suggestion})
        }
        t2.Render()
    },
}

func init() {
    reportCmd.Flags().StringVar(&flagKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
    reportCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "", "Namespace")
    reportCmd.Flags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
    reportCmd.Flags().Float64Var(&flagThreshold, "threshold", 80.0, "Waste threshold percentage")
}
