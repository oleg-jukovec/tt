package steps

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/tarantool/tt/cli/cmdcontext"
)

func TestTemplateRender(t *testing.T) {
	workDir, err := ioutil.TempDir("", testWorkDirName)
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	require.NoError(t, copy.Copy("testdata/cartridge", workDir))

	var createCtx cmdcontext.CreateCtx
	templateCtx := NewTemplateContext()
	templateCtx.AppPath = workDir
	templateCtx.Vars = map[string]string{
		"cluster_cookie": "cookie",
		"user_name":      "admin",
		"app_name":       "app1",
	}

	renderTemplate := RenderTemplate{}
	require.NoError(t, renderTemplate.Run(&createCtx, &templateCtx))

	configFileName := filepath.Join(workDir, "config.lua")
	require.FileExists(t, configFileName)
	require.FileExists(t, filepath.Join(workDir, "app1.yml"))

	buf, err := os.ReadFile(configFileName)
	require.NoError(t, err)
	const expectedText = `cluster_cookie = cookie
login = admin
`
	require.Equal(t, expectedText, string(buf))
}

func TestTemplateRenderMissingVar(t *testing.T) {
	workDir, err := ioutil.TempDir("", testWorkDirName)
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	require.NoError(t, copy.Copy("testdata/cartridge", workDir))

	var createCtx cmdcontext.CreateCtx
	templateCtx := NewTemplateContext()
	templateCtx.AppPath = workDir

	renderTemplate := RenderTemplate{}
	require.EqualError(t, renderTemplate.Run(&createCtx, &templateCtx), "Template instantiation "+
		"error: Template execution failed: template: "+
		"config.lua.tt.template:1:19: executing \"config.lua.tt.template\" "+
		"at <.cluster_cookie>: map has no entry for key \"cluster_cookie\"")
}

func TestTemplateRenderMissingVarInFileName(t *testing.T) {
	workDir, err := ioutil.TempDir("", testWorkDirName)
	require.NoError(t, err)
	defer os.RemoveAll(workDir)

	require.NoError(t, copy.Copy("testdata/cartridge", workDir))

	var createCtx cmdcontext.CreateCtx
	templateCtx := NewTemplateContext()
	templateCtx.AppPath = workDir
	templateCtx.Vars = map[string]string{
		"cluster_cookie": "cookie",
		"user_name":      "admin",
	}

	renderTemplate := RenderTemplate{}
	require.Error(t, renderTemplate.Run(&createCtx, &templateCtx))
}