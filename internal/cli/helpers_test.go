package cli

import (
	"bytes"

	"github.com/spf13/cobra"
)

func captureOutput(cmd *cobra.Command) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.Execute()
	return buf.String(), err
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
