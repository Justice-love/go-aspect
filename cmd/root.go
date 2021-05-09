package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path/filepath"
	"runtime"
	"strconv"
)

var rootCmd = &cobra.Command{
	Use:   "xgc",
	Short: "aspect for golang",
	Long:  "aspect for golang",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				dirname, filename := filepath.Split(f.File)
				lastelem := filepath.Base(dirname)
				filename = filepath.Join(lastelem, filename)
				line := strconv.Itoa(f.Line)
				return "", "[" + filename + ":" + line + "]"
			},
		})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
