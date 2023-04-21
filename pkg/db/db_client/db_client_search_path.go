package db_client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

// GetCurrentSearchPath implements Client
// query the database to get the current session search path
func (c *DbClient) GetCurrentSearchPath(ctx context.Context) ([]string, error) {
	res, err := c.ExecuteSync(ctx, "show search_path")
	if err != nil {
		return nil, err
	}
	pathAsString, ok := res.Rows[0].(*queryresult.RowResult).Data[0].(string)
	if !ok {
		return nil, fmt.Errorf("failed to read the current search path: %s", err.Error())
	}
	return db_common.BuildSearchPathResult(pathAsString)
}

// SetRequiredSessionSearchPath implements Client
// if either a search-path or search-path-prefix is set in config, set the search path
// (otherwise fall back to user search path)
// this just sets the required search path for this client
// - when creating a database session, we will actually set the searchPath
func (c *DbClient) SetRequiredSessionSearchPath(ctx context.Context) error {
	configuredSearchPath := viper.GetStringSlice(constants.ArgSearchPath)
	searchPathPrefix := viper.GetStringSlice(constants.ArgSearchPathPrefix)

	// strip empty elements from search path and prefix
	configuredSearchPath = helpers.RemoveFromStringSlice(configuredSearchPath, "")
	searchPathPrefix = helpers.RemoveFromStringSlice(searchPathPrefix, "")

	requiredSearchPath, err := db_common.GetUserSearchPath(ctx, c.pool)
	if err != nil {
		return err
	}

	// store custom search path and search path prefix
	c.searchPathPrefix = searchPathPrefix

	// if a search path was passed, add 'internal' to the end
	if len(configuredSearchPath) > 0 {
		// add 'internal' schema as last schema in the search path
		requiredSearchPath = append(configuredSearchPath, constants.InternalSchema)
	}

	// add in the prefix if present
	requiredSearchPath = db_common.AddSearchPathPrefix(searchPathPrefix, requiredSearchPath)

	// if either configuredSearchPath or searchPathPrefix are set, store requiredSearchPath as customSearchPath
	if len(configuredSearchPath)+len(searchPathPrefix) > 0 {
		c.customSearchPath = requiredSearchPath
	} else {
		// otherwise clear it
		c.customSearchPath = nil
	}

	return nil
}

// GetRequiredSessionSearchPath implements Client
func (c *DbClient) GetRequiredSessionSearchPath(ctx context.Context) ([]string, error) {
	if c.customSearchPath != nil {
		return c.customSearchPath, nil
	}

	return db_common.GetUserSearchPath(ctx, c.pool)
}

// reload Steampipe config, update viper and re-set required search path
func (c *DbClient) updateRequiredSearchPath(ctx context.Context) error {
	config, errorsAndWarnings := steampipeconfig.LoadSteampipeConfig(viper.GetString(constants.ArgModLocation), "dashboard")
	if errorsAndWarnings.GetError() != nil {
		return errorsAndWarnings.GetError()
	}
	steampipeconfig.GlobalConfig = config
	cmdconfig.SetDefaultsFromConfig(steampipeconfig.GlobalConfig.ConfigMap())
	return c.SetRequiredSessionSearchPath(ctx)
}

// ensure the search path for the database session is as required
func (c *DbClient) ensureSessionSearchPath(ctx context.Context, session *db_common.DatabaseSession) error {
	log.Printf("[TRACE] ensureSessionSearchPath")
	// if we are NOT using a custom search path, nothing to do (we will fall back on the user search path)
	if len(c.customSearchPath) == 0 {
		log.Printf("[TRACE] no custom search path - fall back on the user search path")
		return nil
	}

	// now determine whether the session search path is the same as the required search path
	// if so, return
	if strings.Join(session.SearchPath, ",") == strings.Join(c.customSearchPath, ",") {
		log.Printf("[TRACE] session search path is already correct - nothing to do")
		return nil
	}

	// so we need to set the search path
	log.Printf("[TRACE] session search path will be updated to  %s", strings.Join(c.customSearchPath, ","))

	// TODO KAI USE PARAMS
	_, err := session.Connection.Exec(ctx, fmt.Sprintf("set search_path to %s", strings.Join(db_common.PgEscapeSearchPath(c.customSearchPath), ",")))
	if err == nil {
		// update the session search path property
		session.SearchPath = c.customSearchPath
	}
	return err
}
