package database

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <type>",
		Short: "Create a new database",
		Long: `Create a new database of the specified type.

Supported types: postgresql, mysql, mariadb, mongodb, redis, keydb, clickhouse, dragonfly

Examples:
  coolify databases create postgresql --server-uuid=<uuid> --project-uuid=<uuid> --environment-name=production
  coolify databases create mysql --server-uuid=<uuid> --project-uuid=<uuid> --environment-name=production --name="My MySQL"`,
		Args: cli.ExactArgs(1, "<type>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			dbType := args[0]
			validTypes := []string{"postgresql", "mysql", "mariadb", "mongodb", "redis", "keydb", "clickhouse", "dragonfly"}
			isValid := false
			for _, t := range validTypes {
				if t == dbType {
					isValid = true
					break
				}
			}
			if !isValid {
				return fmt.Errorf("invalid database type '%s'. Valid types: %s", dbType, strings.Join(validTypes, ", "))
			}

			serverUUID, _ := cmd.Flags().GetString("server-uuid")
			projectUUID, _ := cmd.Flags().GetString("project-uuid")
			environmentName, _ := cmd.Flags().GetString("environment-name")
			environmentUUID, _ := cmd.Flags().GetString("environment-uuid")

			if serverUUID == "" || projectUUID == "" {
				return fmt.Errorf("--server-uuid and --project-uuid are required")
			}

			if environmentName == "" && environmentUUID == "" {
				return fmt.Errorf("either --environment-name or --environment-uuid must be provided")
			}

			req := &models.DatabaseCreateRequest{
				ServerUUID:  serverUUID,
				ProjectUUID: projectUUID,
			}

			if environmentName != "" {
				req.EnvironmentName = &environmentName
			}
			if environmentUUID != "" {
				req.EnvironmentUUID = &environmentUUID
			}

			// Common flags
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				desc, _ := cmd.Flags().GetString("description")
				req.Description = &desc
			}
			if cmd.Flags().Changed("image") {
				image, _ := cmd.Flags().GetString("image")
				req.Image = &image
			}
			if cmd.Flags().Changed("destination-uuid") {
				dest, _ := cmd.Flags().GetString("destination-uuid")
				req.DestinationUUID = &dest
			}
			if cmd.Flags().Changed("instant-deploy") {
				instant, _ := cmd.Flags().GetBool("instant-deploy")
				req.InstantDeploy = &instant
			}
			if cmd.Flags().Changed("is-public") {
				isPublic, _ := cmd.Flags().GetBool("is-public")
				req.IsPublic = &isPublic
			}
			if cmd.Flags().Changed("public-port") {
				port, _ := cmd.Flags().GetInt("public-port")
				req.PublicPort = &port
			}

			// Resource limits
			if cmd.Flags().Changed("limits-memory") {
				mem, _ := cmd.Flags().GetString("limits-memory")
				req.LimitsMemory = &mem
			}
			if cmd.Flags().Changed("limits-cpus") {
				cpus, _ := cmd.Flags().GetString("limits-cpus")
				req.LimitsCpus = &cpus
			}

			// PostgreSQL specific
			if dbType == "postgresql" {
				if cmd.Flags().Changed("postgres-user") {
					user, _ := cmd.Flags().GetString("postgres-user")
					req.PostgresUser = &user
				}
				if cmd.Flags().Changed("postgres-password") {
					pass, _ := cmd.Flags().GetString("postgres-password")
					req.PostgresPassword = &pass
				}
				if cmd.Flags().Changed("postgres-db") {
					db, _ := cmd.Flags().GetString("postgres-db")
					req.PostgresDB = &db
				}
			}

			// MySQL specific
			if dbType == "mysql" {
				if cmd.Flags().Changed("mysql-root-password") {
					pass, _ := cmd.Flags().GetString("mysql-root-password")
					req.MysqlRootPassword = &pass
				}
				if cmd.Flags().Changed("mysql-user") {
					user, _ := cmd.Flags().GetString("mysql-user")
					req.MysqlUser = &user
				}
				if cmd.Flags().Changed("mysql-password") {
					pass, _ := cmd.Flags().GetString("mysql-password")
					req.MysqlPassword = &pass
				}
				if cmd.Flags().Changed("mysql-database") {
					db, _ := cmd.Flags().GetString("mysql-database")
					req.MysqlDatabase = &db
				}
			}

			// MariaDB specific
			if dbType == "mariadb" {
				if cmd.Flags().Changed("mariadb-root-password") {
					pass, _ := cmd.Flags().GetString("mariadb-root-password")
					req.MariadbRootPassword = &pass
				}
				if cmd.Flags().Changed("mariadb-user") {
					user, _ := cmd.Flags().GetString("mariadb-user")
					req.MariadbUser = &user
				}
				if cmd.Flags().Changed("mariadb-password") {
					pass, _ := cmd.Flags().GetString("mariadb-password")
					req.MariadbPassword = &pass
				}
				if cmd.Flags().Changed("mariadb-database") {
					db, _ := cmd.Flags().GetString("mariadb-database")
					req.MariadbDatabase = &db
				}
			}

			// MongoDB specific
			if dbType == "mongodb" {
				if cmd.Flags().Changed("mongo-root-username") {
					user, _ := cmd.Flags().GetString("mongo-root-username")
					req.MongoInitdbRootUsername = &user
				}
				if cmd.Flags().Changed("mongo-root-password") {
					pass, _ := cmd.Flags().GetString("mongo-root-password")
					req.MongoInitdbRootPassword = &pass
				}
				if cmd.Flags().Changed("mongo-database") {
					db, _ := cmd.Flags().GetString("mongo-database")
					req.MongoInitdbDatabase = &db
				}
			}

			// Redis specific
			if dbType == "redis" {
				if cmd.Flags().Changed("redis-password") {
					pass, _ := cmd.Flags().GetString("redis-password")
					req.RedisPassword = &pass
				}
			}

			// KeyDB specific
			if dbType == "keydb" {
				if cmd.Flags().Changed("keydb-password") {
					pass, _ := cmd.Flags().GetString("keydb-password")
					req.KeydbPassword = &pass
				}
			}

			// Clickhouse specific
			if dbType == "clickhouse" {
				if cmd.Flags().Changed("clickhouse-admin-user") {
					user, _ := cmd.Flags().GetString("clickhouse-admin-user")
					req.ClickhouseAdminUser = &user
				}
				if cmd.Flags().Changed("clickhouse-admin-password") {
					pass, _ := cmd.Flags().GetString("clickhouse-admin-password")
					req.ClickhouseAdminPassword = &pass
				}
			}

			// Dragonfly specific
			if dbType == "dragonfly" {
				if cmd.Flags().Changed("dragonfly-password") {
					pass, _ := cmd.Flags().GetString("dragonfly-password")
					req.DragonflyPassword = &pass
				}
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			database, err := dbService.Create(ctx, dbType, req)
			if err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(database)
		},
	}

	// Common flags
	cmd.Flags().String("server-uuid", "", "Server UUID (required)")
	cmd.Flags().String("project-uuid", "", "Project UUID (required)")
	cmd.Flags().String("environment-name", "", "Environment name")
	cmd.Flags().String("environment-uuid", "", "Environment UUID")
	cmd.Flags().String("destination-uuid", "", "Destination UUID if server has multiple destinations")
	cmd.Flags().String("name", "", "Database name")
	cmd.Flags().String("description", "", "Database description")
	cmd.Flags().String("image", "", "Docker image")
	cmd.Flags().Bool("instant-deploy", false, "Deploy immediately after creation")
	cmd.Flags().Bool("is-public", false, "Make database publicly accessible")
	cmd.Flags().Int("public-port", 0, "Public port")
	cmd.Flags().String("limits-memory", "", "Memory limit (e.g., '512m', '2g')")
	cmd.Flags().String("limits-cpus", "", "CPU limit (e.g., '0.5', '2')")

	// PostgreSQL flags
	cmd.Flags().String("postgres-user", "", "PostgreSQL user")
	cmd.Flags().String("postgres-password", "", "PostgreSQL password")
	cmd.Flags().String("postgres-db", "", "PostgreSQL database name")

	// MySQL flags
	cmd.Flags().String("mysql-root-password", "", "MySQL root password")
	cmd.Flags().String("mysql-user", "", "MySQL user")
	cmd.Flags().String("mysql-password", "", "MySQL password")
	cmd.Flags().String("mysql-database", "", "MySQL database name")

	// MariaDB flags
	cmd.Flags().String("mariadb-root-password", "", "MariaDB root password")
	cmd.Flags().String("mariadb-user", "", "MariaDB user")
	cmd.Flags().String("mariadb-password", "", "MariaDB password")
	cmd.Flags().String("mariadb-database", "", "MariaDB database name")

	// MongoDB flags
	cmd.Flags().String("mongo-root-username", "", "MongoDB root username")
	cmd.Flags().String("mongo-root-password", "", "MongoDB root password")
	cmd.Flags().String("mongo-database", "", "MongoDB database name")

	// Redis flags
	cmd.Flags().String("redis-password", "", "Redis password")

	// KeyDB flags
	cmd.Flags().String("keydb-password", "", "KeyDB password")

	// Clickhouse flags
	cmd.Flags().String("clickhouse-admin-user", "", "Clickhouse admin user")
	cmd.Flags().String("clickhouse-admin-password", "", "Clickhouse admin password")

	// Dragonfly flags
	cmd.Flags().String("dragonfly-password", "", "Dragonfly password")

	return cmd
}
