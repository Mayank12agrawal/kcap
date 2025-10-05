package cmd

import (
    "github.com/spf13/cobra"
)

var (
    flagKubeconfig string
    flagNamespace  string
    flagJSON       bool
    flagThreshold  float64
)

var rootCmd = &cobra.Command{
    Use:   "kcap",
    Short: "Kubernetes Capacity Analyzer CLI",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(deploysCmd)
    rootCmd.AddCommand(nodesCmd)
    rootCmd.AddCommand(podsCmd)
    rootCmd.AddCommand(recommendCmd)
    rootCmd.AddCommand(reportCmd)
}
