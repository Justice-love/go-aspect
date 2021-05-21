package cmd

import (
	"github.com/Justice-love/go-aspect/build"
	"github.com/Justice-love/go-aspect/file"
	"github.com/Justice-love/go-aspect/inject"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "add aspect and build",
	Long:  "add aspect and build",
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Fail to build: %v", err)
		}
		runBuild(args, wd)
	},
}

func runBuild(args []string, wd string) {
	root := file.SourceCopy(file.TempDir(), file.SourceDir())
	defer func() {
		build.Clean(root)
	}()
	inspect := build.NewInspect(root)
	points := inject.AllEndpoints(inspect.EndpointPath())
	x := file.X{
		RootPath: root,
		Points:   points,
	}
	advices := x.IteratorSource(root)
	inject.DoInjectCode(advices)
	inspect.Build(root, inspect.BuildTags(wd), strings.Join(args, " "))
}

func init() {
	buildCmd.Flags().BoolVar(&build.BuildI, "i", false, "")
	buildCmd.Flags().StringVar(&build.BuildO, "o", "", "output file or directory")
	AddBuildFlags(buildCmd)
	rootCmd.AddCommand(buildCmd)
}

func AddBuildFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&build.BuildA, "a", false, "")
	cmd.Flags().BoolVar(&build.BuildN, "n", false, "")
	//cmd.Flags().IntVar(&build.BuildP, "p", build.BuildP, "")
	//cmd.Flags().BoolVar(&build.BuildV, "v", false, "")
	//cmd.Flags().BoolVar(&build.BuildX, "x", false, "")

	//cmd.Flags().StringVar(&build.BuildBuildmode, "buildmode", "default", "")
	cmd.Flags().BoolVar(&build.BuildLinkshared, "linkshared", false, "")
	//cmd.Flags().StringVar(&build.BuildPkgdir, "pkgdir", "", "")
	//cmd.Flags().BoolVar(&build.BuildRace, "race", false, "")
	cmd.Flags().BoolVar(&build.BuildMSan, "msan", false, "")
	//cmd.Flags().BoolVar(&build.BuildTrimpath, "trimpath", false, "")
	//cmd.Flags().BoolVar(&build.BuildWork, "work", false, "")
}
