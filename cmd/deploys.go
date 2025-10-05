package cmd

import (
    "context"
    "fmt"
    "os"
    "sort"
    "time"

    "github.com/spf13/cobra"
    "github.com/jedib0t/go-pretty/v6/table"
    "kcap/pkg/analysis"
    "kcap/pkg/k8s"
)

var deploysCmd = &cobra.Command{
    Use:   "deploys",
    Short: "Aggregated deployment CPU/memory request vs usage summary",
    Run: func(cmd *cobra.Command, args []string) {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        kube, err := k8s.NewK8sClientWithConfig(flagKubeconfig)
        if err != nil {
            fmt.Println("Error creating kube client:", err)
            os.Exit(1)
        }

        pods, err := kube.ListPods(ctx, flagNamespace)
        if err != nil {
            fmt.Println("Error listing pods:", err)
            os.Exit(1)
        }

        podMetrics, err := kube.PodMetrics(ctx, flagNamespace)
        if err != nil {
            fmt.Println("Warning: Metrics-server not available, usage values will be zero")
            podMetrics = nil
        }

        podRecords := analysis.PodRecords(pods, podMetrics, "")
        deployStats := analysis.DeploymentAggregation(podRecords)

        // Sort by CPU waste descending
        sort.Slice(deployStats, func(i, j int) bool {
            return deployStats[i].WasteCPU > deployStats[j].WasteCPU
        })

        if flagJSON {
            printJSON(deployStats)
            return
        }

        t := table.NewWriter()
        t.SetOutputMirror(os.Stdout)
        t.AppendHeader(table.Row{"NAMESPACE", "DEPLOYMENT", "CPU(REQ/USE m)", "MEM(REQ/USE Mi)", "PODS", "WASTE% CPU", "WASTE% MEM"})

        for _, d := range deployStats {
            cpu := fmt.Sprintf("%d / %d", d.CPUReqMilli, d.CPUUsedMilli)
            mem := fmt.Sprintf("%d / %d", d.MemReqMi, d.MemUsedMi)
            wasteCPU := fmt.Sprintf("%.1f", d.WasteCPU)
            wasteMem := fmt.Sprintf("%.1f", d.WasteMem)
            t.AppendRow(table.Row{d.Namespace, d.Name, cpu, mem, d.PodCount, wasteCPU, wasteMem})
        }
        t.Render()
    },
}

func init() {
    deploysCmd.Flags().StringVar(&flagKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
    deploysCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "", "Namespace")
    deploysCmd.Flags().BoolVar(&flagJSON, "json", false, "Print output as JSON")
}
