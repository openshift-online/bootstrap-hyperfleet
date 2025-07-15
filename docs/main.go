package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func main() {
	// This is needed to make `glog` believe that the flags have already been parsed, otherwise
	// every log messages is prefixed by an error message stating the flags haven't been
	// parsed.
	_ = flag.CommandLine.Parse([]string{})

	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Infof("Unable to set logtostderr to true")
	}

	rootCmd := &cobra.Command{
		Use:  "acme",
		Long: "acme",
		Run:  acme,
	}

	statusCommand := NewStatusCommand()
	rootCmd.AddCommand(statusCommand)

	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}
}

func acme(_ *cobra.Command, _ []string) {
	fmt.Println("hello, miniocm. use status subcmd")
}
