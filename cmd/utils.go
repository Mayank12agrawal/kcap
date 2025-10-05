package cmd

import (
    "encoding/json"
    "os"
)

func printJSON(v interface{}) error {
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    return enc.Encode(v)
}
