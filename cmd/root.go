package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fractalx/fractalx-init/internal/generator"
	"github.com/fractalx/fractalx-init/internal/model"
	"github.com/fractalx/fractalx-init/internal/spec"
	"github.com/fractalx/fractalx-init/internal/validate"
	"github.com/fractalx/fractalx-init/internal/wizard"
	"github.com/spf13/cobra"
)

var (
	fromFile string
	output   string
	noZip    bool
)

var rootCmd = &cobra.Command{
	Use:   "fractalx-init",
	Short: "FractalX Initializr — generate a Spring Boot monolith ready for decomposition",
	Long: `fractalx-init generates a Spring Boot monolith pre-annotated with
FractalX decomposition markers. Run it interactively (default) or
supply a fractalx.yaml spec with --from to skip the wizard.

After generation, run:
  mvn fractalx:decompose

to split your monolith into production-ready microservices.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n  \033[31merror:\033[0m %s\n\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&fromFile, "from", "", "Path to fractalx.yaml spec file (skips the wizard)")
	rootCmd.Flags().StringVar(&output, "output", ".", "Output directory for the generated project")
	rootCmd.Flags().BoolVar(&noZip, "no-zip", false, "Write files directly to disk instead of a ZIP archive")
}

func run(_ *cobra.Command, _ []string) error {
	ps, err := obtainSpec()
	if err != nil {
		return err
	}

	// Validate
	result := validate.Validate(ps)
	if result.HasErrors() || len(result.Warnings) > 0 {
		fmt.Println()
		result.Print()
		if result.HasErrors() {
			return fmt.Errorf("validation failed — fix the errors above and try again")
		}
	}

	// Generate
	fmt.Printf("\n  Generating \033[1m%s\033[0m...\n", ps.ArtifactID)

	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if noZip {
		dest := filepath.Join(output, ps.ArtifactID)
		if err := generator.GenerateDir(ps, dest); err != nil {
			return fmt.Errorf("generate project: %w", err)
		}
		fmt.Printf("\n  \033[32m✔\033[0m  Project written to: \033[1m%s\033[0m\n\n", dest)
		printNextSteps(ps, output, true)
	} else {
		zipPath := filepath.Join(output, ps.ArtifactID+".zip")
		if err := generator.GenerateZip(ps, zipPath); err != nil {
			return fmt.Errorf("generate zip: %w", err)
		}
		fmt.Printf("\n  \033[32m✔\033[0m  Archive created: \033[1m%s\033[0m\n", zipPath)
		fmt.Printf("      Unzip with: \033[2munzip %s\033[0m\n\n", zipPath)
		printNextSteps(ps, output, false)
	}

	return nil
}

func obtainSpec() (*model.ProjectSpec, error) {
	if fromFile != "" {
		fmt.Printf("  Loading spec from \033[1m%s\033[0m...\n", fromFile)
		return spec.FromFile(fromFile)
	}
	return wizard.Run()
}

func printNextSteps(ps *model.ProjectSpec, outDir string, isDir bool) {
	fmt.Println("  Next steps:")
	if isDir {
		fmt.Printf("    cd %s/%s\n", outDir, ps.ArtifactID)
	} else {
		fmt.Printf("    unzip %s/%s.zip && cd %s\n", outDir, ps.ArtifactID, ps.ArtifactID)
	}
	fmt.Println("    mvn spring-boot:run -Dspring-boot.run.profiles=dev")
	fmt.Println("    # later: mvn fractalx:decompose")
	fmt.Println()
}
