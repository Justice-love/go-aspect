package cmd

import (
	"github.com/Justice-love/go-aspect/build"
	"github.com/Justice-love/go-aspect/file"
	"github.com/Justice-love/go-aspect/inject"
	"github.com/spf13/cobra"
	"os"
)

var debug = &cobra.Command{
	Use:   "debug",
	Short: "generate source code to ~/.xgc",
	Long:  "add aspect and build",
	Run: func(cmd *cobra.Command, args []string) {
		_ = os.RemoveAll(file.DebugDir())
		root := file.SourceCopy(file.DebugDir(), file.SourceDir())
		inspect := build.NewInspect(root)
		points := inject.AllEndpoints(inspect.EndpointPath())
		x := file.X{
			RootPath: root,
			Points:   points,
		}
		advices := x.IteratorSource(root)
		inject.DoInjectCode(advices)
	},
}

func init() {
	rootCmd.AddCommand(debug)
}
