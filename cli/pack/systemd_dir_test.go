package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tarantool/tt/cli/config"
	"github.com/tarantool/tt/cli/pack/test_helpers"
)

func Test_initSystemdDir(t *testing.T) {
	baseTestDir := t.TempDir()
	defer os.RemoveAll(baseTestDir)

	prefixToUnit := filepath.Join("etc", "systemd", "system")
	fakeCfgPath := "/path/to/cfg"

	var (
		test1Dir = "test_default_template_values"
		test2Dir = "test_default_template_partly_defined_values"
		test3Dir = "test_default_template_fully_defined_values"
		appDir   = "app"
	)
	testDirs := []string{
		filepath.Join(test1Dir, appDir),
		filepath.Join(test2Dir, appDir),
		filepath.Join(test3Dir, appDir),
	}

	err := test_helpers.CreateDirs(baseTestDir, testDirs)
	require.NoError(t, err)

	type args struct {
		baseDirPath string
		pathToEnv   string
		opts        *config.CliOpts
		packCtx     *PackCtx
	}
	tests := []struct {
		name    string
		prepare func() error
		args    args
		wantErr assert.ErrorAssertionFunc
		check   func() error
	}{
		{
			name: "Default template and values test 1",
			args: args{
				baseDirPath: filepath.Join(baseTestDir, test1Dir),
				pathToEnv:   fakeCfgPath,
				opts: &config.CliOpts{
					App: &config.AppOpts{
						InstancesEnabled: filepath.Join(baseTestDir, test1Dir),
					},
				},
				packCtx: &PackCtx{
					Name:    "pack",
					AppList: []string{appDir},
				},
			},
			prepare: func() error {
				return nil
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			check: func() error {
				content, err := ioutil.ReadFile(filepath.Join(baseTestDir,
					test1Dir, prefixToUnit, "pack.service"))
				if err != nil {
					return err
				}
				contentStr := string(content)
				if contentStr != expectedContentInAppUnitTest1 {
					return fmt.Errorf("the unit file doesn't contain the passed value")
				}
				return nil
			},
		},
		{
			name: "Default template and partly defined values test 2",
			args: args{
				baseDirPath: filepath.Join(baseTestDir, test2Dir),
				pathToEnv:   fakeCfgPath,
				opts: &config.CliOpts{
					App: &config.AppOpts{
						InstancesEnabled: filepath.Join(baseTestDir, test2Dir),
					},
				},
				packCtx: &PackCtx{
					Name:    "pack",
					AppList: []string{appDir},
					RpmDeb: RpmDebCtx{
						SystemdUnitParamsFile: filepath.Join(baseTestDir,
							test2Dir, "custom-template.txt"),
					},
				},
			},
			prepare: func() error {
				return os.WriteFile(filepath.Join(baseTestDir,
					test2Dir, "custom-template.txt"),
					[]byte(testUnitFilePartlyDefinedParamsContent), 0666)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			check: func() error {
				content, err := ioutil.ReadFile(filepath.Join(baseTestDir,
					test2Dir, prefixToUnit, "pack.service"))
				if err != nil {
					return err
				}
				contentStr := string(content)
				if contentStr != expectedContentInAppUnitTest2 {
					return fmt.Errorf("the unit file doesn't contain the passed value")
				}
				return nil
			},
		},
		{
			name: "Default template and fully defined values test 3",
			args: args{
				baseDirPath: filepath.Join(baseTestDir, test3Dir),
				pathToEnv:   fakeCfgPath,
				opts: &config.CliOpts{
					App: &config.AppOpts{
						InstancesEnabled: filepath.Join(baseTestDir, test3Dir),
					},
				},
				packCtx: &PackCtx{
					Name:    "pack",
					AppList: []string{appDir},
					RpmDeb: RpmDebCtx{
						SystemdUnitParamsFile: filepath.Join(baseTestDir,
							test3Dir, "custom-template.txt"),
					},
				},
			},
			prepare: func() error {
				return os.WriteFile(filepath.Join(baseTestDir,
					test3Dir, "custom-template.txt"),
					[]byte(testUnitFileFullyDefinedParamsContent), 0666)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			check: func() error {
				content, err := ioutil.ReadFile(filepath.Join(baseTestDir,
					test3Dir, prefixToUnit, "pack.service"))
				if err != nil {
					return err
				}
				contentStr := string(content)
				if contentStr != expectedContentInAppUnitTest3 {
					return fmt.Errorf("the unit file doesn't contain the passed value")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.prepare())
			tt.wantErr(t, initSystemdDir(tt.args.packCtx, tt.args.opts,
				tt.args.baseDirPath, tt.args.pathToEnv),
				fmt.Sprintf("initSystemdDir(%v, %v, %v, %v)",
					tt.args.baseDirPath, tt.args.pathToEnv, tt.args.opts, tt.args.packCtx))

			assert.NoError(t, tt.check())
		})
	}
}

const (
	testUnitFilePartlyDefinedParamsContent = `FdLimit: 2048`

	testUnitFileFullyDefinedParamsContent = `TT: /usr/bin/tt
ConfigPath: /path/cfg/tarantool.yaml
EnvName: customEnvName
FdLimit: 2048
`

	expectedContentInAppUnitTest1 = `[Unit]
Description=Tarantool app pack.default
After=network.target

[Service]
Type=forking
ExecStart=/path/to/cfg/env/bin/tt -L /path/to/cfg/tarantool.yaml start
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=65535

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=pack
`
	expectedContentInAppUnitTest2 = `[Unit]
Description=Tarantool app pack.default
After=network.target

[Service]
Type=forking
ExecStart=/path/to/cfg/env/bin/tt -L /path/to/cfg/tarantool.yaml start
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=2048

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=pack
`
	expectedContentInAppUnitTest3 = `[Unit]
Description=Tarantool app customEnvName.default
After=network.target

[Service]
Type=forking
ExecStart=/usr/bin/tt -L /path/cfg/tarantool.yaml start
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=2048

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=customEnvName
`
)

func Test_getUnitParams(t *testing.T) {
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	type args struct {
		packCtx   *PackCtx
		pathToEnv string
		envName   string
	}
	tests := []struct {
		name    string
		args    args
		prepare func() error
		want    map[string]interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "default parameters",
			args: args{
				envName:   "envName",
				pathToEnv: "/path/to/env",
				packCtx: &PackCtx{
					WithoutBinaries: true,
					RpmDeb: RpmDebCtx{
						SystemdUnitParamsFile: "",
					},
				},
			},
			want: map[string]interface{}{
				"TT":         "tt",
				"ConfigPath": "/path/to/env",
				"FdLimit":    defaultInstanceFdLimit,
				"EnvName":    "envName",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err != nil
			},
			prepare: func() error {
				return nil
			},
		},
		{
			name: "partly defined parameters",
			args: args{
				envName:   "envName",
				pathToEnv: "/path/to/env",
				packCtx: &PackCtx{
					WithoutBinaries: true,
					RpmDeb: RpmDebCtx{
						SystemdUnitParamsFile: filepath.Join(testDir, "partly-params.yaml"),
					},
				},
			},
			want: map[string]interface{}{
				"TT":         "tt",
				"ConfigPath": "/path/to/env",
				"FdLimit":    1024,
				"EnvName":    "envName",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err != nil
			},
			prepare: func() error {
				err := ioutil.WriteFile(filepath.Join(testDir, "partly-params.yaml"),
					[]byte("FdLimit: 1024\n"), 0666)
				return err
			},
		},
		{
			name: "fully defined parameters",
			args: args{
				envName:   "envName",
				pathToEnv: "/path/to/env",
				packCtx: &PackCtx{
					WithoutBinaries: true,
					RpmDeb: RpmDebCtx{
						SystemdUnitParamsFile: filepath.Join(testDir, "fully-params.yaml"),
					},
				},
			},
			want: map[string]interface{}{
				"TT":         "/usr/bin/tt",
				"ConfigPath": "/test/path",
				"FdLimit":    1024,
				"EnvName":    "testEnv",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err != nil
			},
			prepare: func() error {
				err := ioutil.WriteFile(filepath.Join(testDir, "fully-params.yaml"),
					[]byte("FdLimit: 1024\n"+
						"TT: /usr/bin/tt\n"+
						"ConfigPath: /test/path\n"+
						"EnvName: testEnv\n"), 0666)
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.prepare())
			got, err := getUnitParams(tt.args.packCtx, tt.args.pathToEnv, tt.args.envName)
			if !tt.wantErr(t, err, fmt.Sprintf("getUnitParams(%v, %v, %v)",
				tt.args.packCtx, tt.args.pathToEnv, tt.args.envName)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getUnitParams(%v, %v, %v)",
				tt.args.packCtx, tt.args.pathToEnv, tt.args.envName)
		})
	}
}
