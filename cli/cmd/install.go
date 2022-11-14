package cmd

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/tt/cli/cmdcontext"
	"github.com/tarantool/tt/cli/configure"
	"github.com/tarantool/tt/cli/install"
	"github.com/tarantool/tt/cli/modules"
	"github.com/tarantool/tt/cli/search"
)

var installFlags install.InstallationFlag

// NewInstallCmd creates install command.
func NewInstallCmd() *cobra.Command {
	var installCmd = &cobra.Command{
		Use:   "install <PROGRAM> [flags]",
		Short: "Install program",
		Long: "Install program\n\n" +
			"Available programs:\n" +
			"tt - Tarantool CLI\n" +
			"tarantool - Tarantool\n" +
			"tarantool-ee - Tarantool enterprise edition\n" +
			"Example: tt install tarantool | tarantool" + search.VersionCliSeparator + "version",
		Run: func(cmd *cobra.Command, args []string) {
			err := modules.RunCmd(&cmdCtx, cmd.Name(), &modulesInfo, internalInstallModule, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}
	installCmd.Flags().BoolVarP(&installFlags.Force, "force", "f", false,
		"force requirements errors")
	installCmd.Flags().BoolVarP(&installFlags.Noclean, "no-clean", "", false,
		"don't delete temporary files")
	installCmd.Flags().BoolVarP(&installFlags.Reinstall, "reinstall", "", false,
		"reinstall program")
	installCmd.Flags().BoolVarP(&installFlags.Local, "local-repo", "", false,
		"install from local files")
	installCmd.Flags().BoolVarP(&installFlags.BuildInDocker, "use-docker", "", false,
		"build tarantool in Ubuntu 16.04 docker container")
	return installCmd
}

// internalInstallModule is a default install module.
func internalInstallModule(cmdCtx *cmdcontext.CmdCtx, args []string) error {
	cliOpts, err := configure.GetCliOpts(cmdCtx.Cli.ConfigPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(cmdCtx.Cli.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("There is no tarantool.yaml found, please create one")
	}

	installFlags.Verbose = cmdCtx.Cli.Verbose
	err = install.Install(args, cliOpts.App.BinDir, cliOpts.App.IncludeDir+"/include",
		installFlags, cliOpts.Repo.Install, cliOpts)
	return err
}
