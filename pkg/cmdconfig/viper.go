package cmdconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"os"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
)

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

// BootstrapViper sets up viper with the essential path config (mod-location and install-dir)
func BootstrapViper() {
	viper.SetDefault(constants.EnvInstallDir, filepaths.DefaultInstallDir)
	bootstrapEnvMappings := map[string]envMapping{
		constants.EnvInstallDir:     {constants.ArgInstallDir, "string"},
		constants.EnvWorkspaceChDir: {constants.ArgModLocation, "string"},
		constants.EnvModLocation:    {constants.ArgModLocation, "string"},
	}
	setViperDefaultFromEnvMapping(bootstrapEnvMappings)
}

// SetViperDefaults sets up viper with default values for all config which is set via hcl config or env vars
func SetViperDefaults(configMap map[string]interface{}, workspaceProfileMap map[string]*modconfig.WorkspaceProfile) {
	// least precedence is default workspace profile
	overrideDefaultsFromWorkspaceProfile(workspaceProfileMap["default"])
	setBaseDefaults()
	if configMap != nil {
		overrideDefaultsFromConfig(configMap)
	}
	overrideDefaultsFromEnv()

	// if an explicit workspace profile was set, it has highest priority (apart from command line args)
	if viper.IsSet(constants.ArgWorkspaceProfile) {
		overrideDefaultsFromWorkspaceProfile(steampipeconfig.GlobalWorkspaceProfile)
	}
	// todo KAI override options defaults from workspace options
}

func overrideDefaultsFromWorkspaceProfile(profile *modconfig.WorkspaceProfile) {
	if profile.CloudHost != "" {
		viper.SetDefault(constants.ArgCloudHost, profile.CloudHost)
	}
	if profile.CloudToken != "" {
		viper.SetDefault(constants.ArgCloudToken, profile.CloudToken)
	}
	if profile.ModLocation != "" {
		viper.SetDefault(constants.ArgModLocation, profile.ModLocation)
	}
	if profile.CloudHost != "" {
		viper.SetDefault(constants.ArgCloudHost, profile.CloudHost)
	}
}

// for keys which do not have a corresponding command flag, we need a separate defaulting mechanism
func setBaseDefaults() {
	defaults := map[string]interface{}{
		constants.ArgUpdateCheck:    true,
		constants.ArgTelemetry:      constants.TelemetryInfo,
		constants.ArgInstallDir:     filepaths.DefaultInstallDir,
		constants.ArgDatabasePort:   constants.DatabaseDefaultPort,
		constants.ArgMaxCacheSizeMb: constants.DefaultMaxCacheSizeMb,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

// set default values from hcl config
func overrideDefaultsFromConfig(configMap map[string]interface{}) {
	for k, v := range configMap {
		viper.SetDefault(k, v)
	}
}

type envMapping struct {
	configVar string
	// "string", "int", "bool"
	varType string
}

// set default values from env vars
func overrideDefaultsFromEnv() {
	// a map of known environment variables to map to viper keys
	envMappings := map[string]envMapping{
		constants.EnvIntrospection:     {constants.ArgIntrospection, "string"},
		constants.EnvTelemetry:         {constants.ArgTelemetry, "string"},
		constants.EnvUpdateCheck:       {constants.ArgUpdateCheck, "bool"},
		constants.EnvCloudHost:         {constants.ArgCloudHost, "string"},
		constants.EnvCloudToken:        {constants.ArgCloudToken, "string"},
		constants.EnvSnapshotLocation:  {constants.ArgSnapshotLocation, "string"},
		constants.EnvWorkspaceDatabase: {constants.ArgWorkspaceDatabase, "string"},
		constants.EnvWorkspaceProfile:  {constants.ArgWorkspaceProfile, "string"},
		constants.EnvServicePassword:   {constants.ArgServicePassword, "string"},
		constants.EnvCheckDisplayWidth: {constants.ArgCheckDisplayWidth, "int"},
		constants.EnvMaxParallel:       {constants.ArgMaxParallel, "int"},
	}
	setViperDefaultFromEnvMapping(envMappings)
}

func setViperDefaultFromEnvMapping(envMappings map[string]envMapping) {
	for k, v := range envMappings {
		if val, ok := os.LookupEnv(k); ok {
			switch v.varType {
			case "string":
				viper.SetDefault(v.configVar, val)
			case "bool":
				if boolVal, err := types.ToBool(val); err == nil {
					viper.SetDefault(v.configVar, boolVal)
				}
			case "int":
				if intVal, err := types.ToInt64(val); err == nil {
					viper.SetDefault(v.configVar, intVal)
				}
			default:
				// must be an invalid value in the map above
				panic(fmt.Sprintf("invalid env var mapping type: %s", v.varType))
			}
		}
	}
}
