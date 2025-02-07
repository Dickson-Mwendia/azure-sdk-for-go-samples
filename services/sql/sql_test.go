// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package sql

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	// sql driver
	_ "github.com/denisenkom/go-mssqldb"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/util"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/resources"

	"github.com/marstr/randname"
)

var (
	serverName = generateName("gosdksql")
	dbName     = "sqldb1"
	dbLogin    = "sqldbuser1"
	dbPassword = "sqldbuserpass!1"
)

func addLocalEnvAndParse() error {
	// parse env at top-level (also controls dotenv load)
	err := config.ParseEnvironment()
	if err != nil {
		return fmt.Errorf("failed to add top-level env: %+v", err)
	}
	return nil
}

func addLocalFlagsAndParse() error {
	// add top-level flags
	err := config.AddFlags()
	if err != nil {
		return fmt.Errorf("failed to add top-level flags: %+v", err)
	}

	flag.StringVar(&serverName, "sqlServerName", serverName, "Name for SQL server.")
	flag.StringVar(&dbName, "sqlDbName", dbName, "Name for SQL database.")
	flag.StringVar(&dbLogin, "sqlDbUsername", dbLogin, "Username for SQL login.")
	flag.StringVar(&dbPassword, "sqlDbPassword", dbPassword, "Password for SQL login.")

	// parse all flags
	flag.Parse()
	return nil
}

func setup() error {
	var err error
	err = addLocalEnvAndParse()
	if err != nil {
		return err
	}
	err = addLocalFlagsAndParse()
	if err != nil {
		return err
	}

	return nil
}

func teardown() error {
	if config.KeepResources() == false {
		// does not wait
		_, err := resources.DeleteGroup(context.Background(), config.GroupName())
		if err != nil {
			return err
		}
	}
	return nil
}

// test helpers
func generateName(prefix string) string {
	return strings.ToLower(randname.GenerateWithPrefix(prefix, 5))
}

// TestMain sets up the environment and initiates tests.
func TestMain(m *testing.M) {
	var err error
	var code int

	err = setup()
	if err != nil {
		log.Fatalf("could not set up environment: %+v", err)
	}

	code = m.Run()

	err = teardown()
	if err != nil {
		log.Fatalf(
			"could not tear down environment: %v\n; original exit code: %v\n",
			err, code)
	}

	os.Exit(code)
}

// Example_createDatabase creates a SQL server and database, then creates a table and inserts a record.
func Example_createDatabase() {
	var groupName = config.GenerateGroupName("DatabaseQueries")
	config.SetGroupName(groupName)

	serverName = strings.ToLower(serverName)

	ctx := context.Background()
	defer resources.Cleanup(ctx)

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}

	_, err = CreateServer(ctx, serverName, dbLogin, dbPassword)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot create sql server: %+v", err))
	}
	util.PrintAndLog("sql server created")

	_, err = CreateDB(ctx, serverName, dbName)
	if err != nil {
		util.LogAndPanic(fmt.Errorf("cannot create sql database: %+v", err))
	}
	util.PrintAndLog("database created")

	err = CreateFirewallRules(ctx, serverName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("database firewall rules set")

	err = testSQLDataplane(serverName, dbName, dbLogin, dbPassword)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("database operations performed")

	// Output:
	// sql server created
	// database created
	// database firewall rules set
	// database operations performed
}

// testSQLDataplane executes some simple SQL queries
func testSQLDataplane(server, database, username, password string) error {
	log.Printf("available drivers: %v", sql.Drivers())

	db, err := Open(server, database, username, password)
	if err != nil {
		return err
	}

	err = CreateTable(db)
	if err != nil {
		return err
	}

	err = Insert(db)
	if err != nil {
		return err
	}

	return Query(db)
}
