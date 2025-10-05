package cmd

import (
    "context"
    "fmt"
    "os"
    "strconv"
    "time"

    "github.com/spf13/cobra"
    "github.com/jedib0t/go-pretty/v6/table"
    "kcap/pkg/analysis"
    "kcap/pkg/k8s"
)

var podsCmd = &cobra.Command{
    Use:   "pods",
    Short: "Show pods request vs usage. Use --namespace to limit.",
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

        list := analysis.PodRecords(pods, podMetrics, "")

        if flagJSON {
            printJSON(list)
            return
        }

        t := table.NewWriter()
        t.SetOutputMirror(os.Stdout)
        t.AppendHeader(table.Row{
            "NAMESPACE", "POD", "NODE", "CPU(REQ/USE M)",
            "MEM(REQ/USE MI)", "OWNER", "DAEMONSET", "WASTE% (CPU)", "WASTE% (MEM)",
        })

        for _, p := range list {
            cpu := fmt.Sprintf("%d / %d", p.CPUReqMilli, p.CPUUsedMilli)
            mem := fmt.Sprintf("%d / %d", p.MemReqMi, p.MemUsedMi)

            cpuWaste := "N/A"
            memWaste := "N/A"
            if p.CPUReqMilli > 0 {
                cpuWaste = fmt.Sprintf("%.1f", (1.0 - float64(p.CPUUsedMilli)/float64(p.CPUReqMilli))*100.0)
            }
            if p.MemReqMi > 0 {
                memWaste = fmt.Sprintf("%.1f", (1.0 - float64(p.MemUsedMi)/float64(p.MemReqMi))*100.0)
            }

            t.AppendRow(table.Row{
                p.Namespace, p.Name, p.NodeName,
                cpu, mem, p.Owner, strconv.FormatBool(p.IsDaemonSet),
                cpuWaste, memWaste,
            })
        }
        t.Render()
    },
}

func init() {
    podsCmd.Flags().StringVar(&flagKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
    podsCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "", "Namespace")
    podsCmd.Flags().BoolVar(&flagJSON, "json", false, "Print output as JSON")
}
