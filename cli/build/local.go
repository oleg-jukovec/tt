package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/tt/cli/cmdcontext"
	"github.com/tarantool/tt/cli/rocks"
	"github.com/tarantool/tt/cli/util"
)

// getPreBuildScripts returns a slice of supported pre-build executables.
func getPreBuildScripts() []string {
	return []string{"tt.pre-build", "cartridge.pre-build"}
}

// getPostBuildScripts returns a slice of supported post-build executables.
func getPostBuildScripts() []string {
	return []string{"tt.post-build", "cartridge.post-build"}
}

// runBuildHook runs first existing executable from hookNames list.
func runBuildHook(buildCtx *cmdcontext.BuildCtx, hookNames []string) error {
	for _, hookName := range hookNames {
		buildHookPath := filepath.Join(buildCtx.BuildDir, hookName)

		if _, err := os.Stat(buildHookPath); err == nil {
			log.Infof("Running `%s`", buildHookPath)
			err = util.RunHook(buildHookPath, false)
			if err != nil {
				return fmt.Errorf("Failed to run build hook: %s", err)
			}
			break
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("Failed to run build hook: %s", err)
		}
	}

	return nil
}

// buildLocal builds an application locally.
func buildLocal(cmdCtx *cmdcontext.CmdCtx, buildCtx *cmdcontext.BuildCtx) error {
	cwd, err := util.Chdir(buildCtx.BuildDir)
	if err != nil {
		return err
	}
	defer util.Chdir(cwd)

	// Run Pre-build.
	if err := runBuildHook(buildCtx, getPreBuildScripts()); err != nil {
		return fmt.Errorf("Run pre-build hook failed: %s", err)
	}

	// Run rocks make.
	rocksMakeCmd := []string{"make"}
	if buildCtx.SpecFile != "" {
		rocksMakeCmd = append(rocksMakeCmd, buildCtx.SpecFile)
	}
	if err := rocks.Exec(cmdCtx, rocksMakeCmd); err != nil {
		return err
	}

	// Run Post-build.
	if err := runBuildHook(buildCtx, getPostBuildScripts()); err != nil {
		return fmt.Errorf("Run post-build hook failed: %s", err)
	}

	return nil
}
