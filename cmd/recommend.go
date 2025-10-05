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

var recommendCmd = &cobra.Command{
    Use:   "recommend",
    Short: "Provide actionable recommendations for nodes and pods",
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

        podMetrics, _ := kube.PodMetrics(ctx, flagNamespace)

        nodeStats := analysis.NodeStats(nodes, nodeMetrics, pods)
        podRecords := analysis.PodRecords(pods, podMetrics, "")

        nodeRecs := analysis.RecommendNodes(nodeStats)
        podRecs := analysis.RecommendPods(podRecords, flagThreshold)

        recs := append(nodeRecs, podRecs...)

        if flagJSON {
            printJSON(recs)
            return
        }

        t := table.NewWriter()
        t.SetOutputMirror(os.Stdout)
        t.AppendHeader(table.Row{"TYPE", "DETAILS", "SUGGESTION"})

        for _, r := range recs {
            t.AppendRow(table.Row{r.Type, r.Details, r.Suggestion})
        }
        t.Render()
    },
}

func init() {
    recommendCmd.Flags().StringVar(&flagKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
    recommendCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "", "Namespace")
    recommendCmd.Flags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
    recommendCmd.Flags().Float64Var(&flagThreshold, "threshold", 80.0, "Waste threshold percentage")
}
