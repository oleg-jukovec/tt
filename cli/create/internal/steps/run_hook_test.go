package steps

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tarantool/tt/cli/cmdcontext"
)

func TestRunHooks(t *testing.T) {
	workDir, err := ioutil.TempDir("", testWorkDirName)
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	var createCtx cmdcontext.CreateCtx
	templateCtx := NewTemplateContext()
	templateCtx.AppPath = workDir
	templateCtx.IsManifestPresent = true
	templateCtx.Manifest.PreHook = "pre-gen.sh"
	templateCtx.Manifest.PostHook = "post-gen.sh"

	require.NoError(t, copy.Copy("testdata/hooks", workDir))

	runPreHook := RunHook{HookType: "pre"}
	runPostHook := RunHook{HookType: "post"}
	assert.NoError(t, runPreHook.Run(&createCtx, &templateCtx))
	assert.NoError(t, runPostHook.Run(&createCtx, &templateCtx))
	assert.FileExists(t, filepath.Join(templateCtx.AppPath, "pre-script-invoked"))
	assert.FileExists(t, filepath.Join(templateCtx.AppPath, "post-script-invoked"))

	// Check if scripts are removed.
	assert.NoFileExists(t, filepath.Join(workDir, templateCtx.Manifest.PreHook))
	assert.NoFileExists(t, filepath.Join(workDir, templateCtx.Manifest.PostHook))
}

func TestRunHooksMissingScript(t *testing.T) {
	workDir, err := ioutil.TempDir("", testWorkDirName)
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	var createCtx cmdcontext.CreateCtx
	templateCtx := NewTemplateContext()
	templateCtx.AppPath = workDir
	templateCtx.IsManifestPresent = true
	templateCtx.Manifest.PreHook = "pre-gen.sh"
	templateCtx.Manifest.PostHook = "post-gen.sh"

	runPreHook := RunHook{HookType: "pre"}
	runPostHook := RunHook{HookType: "post"}
	require.EqualError(t, runPreHook.Run(&createCtx, &templateCtx),
		fmt.Sprintf("Error access to %[1]s: stat %[1]s: no such file or directory",
			filepath.Join(workDir, "pre-gen.sh")))

	require.EqualError(t, runPostHook.Run(&createCtx, &templateCtx),
		fmt.Sprintf("Error access to %[1]s: stat %[1]s: no such file or directory",
			filepath.Join(workDir, "post-gen.sh")))
}