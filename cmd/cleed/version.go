package main

import "github.com/spf13/cobra"

func (r *Root) initVersion() {
	r.Cmd.AddCommand(&cobra.Command{
		Run:   r.RunVersion,
		Use:   "version",
		Short: "Display the version of cleed",
	})
}

func (r *Root) RunVersion(cmd *cobra.Command, args []string) {
	r.printer.Printf("cleed version %s\n", r.version)
}
