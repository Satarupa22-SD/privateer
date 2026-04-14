package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/privateerproj/privateer-sdk/command"
)

func (c *CLI) addRunCmd() {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run plugins that have been specified in the config.",
		Long: `
When everything is battoned down, it is time to run forth.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			c.logger.Trace("run called")
			exitCode := c.run()
			os.Exit(int(exitCode))
		},
	}
	c.rootCmd.AddCommand(runCmd)
}

// run executes all plugins, prints a JSON/YAML summary to stdout, and returns an exit code.
func (c *CLI) run() (exitCode int) {
	c.setupCloseHandler()
	exitCode, results := command.Run(c.logger, command.GetPlugins)
	c.printSummary(command.RunSummary{Results: results})
	return exitCode
}

// printSummary marshals the run summary to stdout in the format specified by
// the "output" config key (json by default, yaml if set to "yaml").
func (c *CLI) printSummary(summary command.RunSummary) {
	var (
		data []byte
		err  error
	)
	if viper.GetString("output") == "yaml" {
		data, err = yaml.Marshal(summary)
	} else {
		data, err = json.MarshalIndent(summary, "", "  ")
	}
	if err != nil {
		c.logger.Error("failed to marshal run summary", "error", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}

// setupCloseHandler creates a signal listener on a new goroutine which will notify
// the program if it receives an interrupt from the OS (SIGINT or SIGTERM).
func (c *CLI) setupCloseHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		c.logger.Error("Test execution was aborted by user")
		os.Exit(int(command.Aborted))
	}()
}
