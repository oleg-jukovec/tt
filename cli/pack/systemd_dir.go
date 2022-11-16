package pack

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/tt/cli/config"
	"github.com/tarantool/tt/cli/util"
	"gopkg.in/yaml.v2"
)

const (
	defaultInstanceFdLimit = 65535
)

// initSystemdDir generates systemd unit files for every application in the current bundle.
// pathToEnv is a path to environment in the target system.
// baseDirPath is a root of the directory which will get packed.
func initSystemdDir(packCtx *PackCtx, opts *config.CliOpts,
	baseDirPath, pathToEnv string) error {
	log.Infof("Initializing systemd directory.")

	packageName, err := getPackageName(packCtx, opts, "", false)
	if err != nil {
		return err
	}

	systemdBaseDir := filepath.Join(baseDirPath, "etc", "systemd", "system")
	err = os.MkdirAll(systemdBaseDir, dirPermissions|5)
	if err != nil {
		return err
	}

	appUnitTemplate := appUnitContentTemplate
	appInstUnitTemplate := appInstUnitContentTemplate

	contentParams, err := getUnitParams(packCtx, pathToEnv, packageName)
	if err != nil {
		return err
	}

	fileNameCtx := map[string]interface{}{
		"EnvName": packageName,
	}

	appUnitPathTempl := filepath.Join(systemdBaseDir, "{{ .EnvName }}.service")
	appUnitPath, err := util.GetTextTemplatedStr(&appUnitPathTempl, fileNameCtx)
	if err != nil {
		return err
	}
	err = util.CheckTemplateCompletion(appUnitTemplate)
	if err != nil {
		return err
	}
	err = util.InstantiateFileFromTemplate(appUnitPath, appUnitTemplate, contentParams)
	if err != nil {
		return err
	}

	appInstUnitPathTempl := filepath.Join(systemdBaseDir,
		"{{ .EnvName }}%.service")
	appInstUnitPath, err := util.GetTextTemplatedStr(&appInstUnitPathTempl, fileNameCtx)
	if err != nil {
		return err
	}
	err = util.CheckTemplateCompletion(appUnitTemplate)
	if err != nil {
		return err
	}
	err = util.InstantiateFileFromTemplate(appInstUnitPath, appInstUnitTemplate, contentParams)
	if err != nil {
		return err
	}

	return nil
}

// getUnitParams checks if there is a passed unit params file in context and
// returns its content. Otherwise, it returns the default params.
func getUnitParams(packCtx *PackCtx, pathToEnv,
	envName string) (map[string]interface{}, error) {
	pathToEnvFile := filepath.Join(pathToEnv, "tarantool.yaml")
	ttBinary := getTTBinary(packCtx, pathToEnv)

	referenceParams := map[string]interface{}{
		"TT":         ttBinary,
		"ConfigPath": pathToEnvFile,
		"FdLimit":    defaultInstanceFdLimit,
		"EnvName":    envName,
	}

	contentParams := make(map[string]interface{})

	if packCtx.RpmDeb.SystemdUnitParamsFile != "" {
		unitTemplFile, err := os.Open(packCtx.RpmDeb.SystemdUnitParamsFile)
		if err != nil {
			return nil, err
		}

		err = yaml.NewDecoder(unitTemplFile).Decode(&contentParams)
		if err != nil {
			return nil, err
		}
	}
	for key := range referenceParams {
		if _, ok := contentParams[key]; !ok {
			contentParams[key] = referenceParams[key]
		}
	}
	return contentParams, nil
}

func getTTBinary(packCtx *PackCtx, envSystemPath string) string {
	if (!packCtx.TarantoolIsSystem && !packCtx.WithoutBinaries) ||
		packCtx.WithBinaries {
		return filepath.Join(envSystemPath, envBinPath, "tt")
	}
	return "tt"
}

const (
	appUnitContentTemplate = `[Unit]
Description=Tarantool app {{ .EnvName }}.default
After=network.target

[Service]
Type=forking
ExecStart={{ .TT }} -L {{ .ConfigPath }} start
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE={{ .FdLimit }}

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias={{ .EnvName }}
`

	//nolint
	appInstUnitContentTemplate = `[Unit]
Description=Tarantool app {{ .EnvName }}@%i
After=network.target

[Service]
Type=forking
ExecStart={{ .TT }} -L {{ .ConfigPath }} start %i
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE={{ .FdLimit }}

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=%i
`
)
