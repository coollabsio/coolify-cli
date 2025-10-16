package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var databasesCmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{"databases", "db", "dbs"},
	Short:   "Manage Coolify databases",
	Long:    `Manage Coolify databases (PostgreSQL, MySQL, MongoDB, Redis, MariaDB, KeyDB, Clickhouse, Dragonfly).`,
}

var listDatabasesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all databases",
	Long:  `List all databases in your Coolify instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		databases, err := dbService.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list databases: %w", err)
		}

		formatter, err := output.NewFormatter(Format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(databases)
	},
}

var getDatabaseCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get database details",
	Long:  `Get detailed information about a specific database by UUID.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		database, err := dbService.Get(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}

		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
		formatter, err := output.NewFormatter(Format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(database)
	},
}

var createDatabaseCmd = &cobra.Command{
	Use:   "create <type>",
	Short: "Create a new database",
	Long: `Create a new database of the specified type.

Supported types: postgresql, mysql, mariadb, mongodb, redis, keydb, clickhouse, dragonfly

Examples:
  coolify databases create postgresql --server-uuid=<uuid> --project-uuid=<uuid> --environment-name=production
  coolify databases create mysql --server-uuid=<uuid> --project-uuid=<uuid> --environment-name=production --name="My MySQL"`,
	Args: exactArgs(1, "<type>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
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
				req.PostgresDb = &db
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

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		database, err := dbService.Create(ctx, dbType, req)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

		formatter, err := output.NewFormatter(Format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(database)
	},
}

var updateDatabaseCmd = &cobra.Command{
	Use:   "update <uuid>",
	Short: "Update a database",
	Long:  `Update a database's configuration by UUID.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		req := &models.DatabaseUpdateRequest{}
		hasChanges := false

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			req.Name = &name
			hasChanges = true
		}
		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			req.Description = &desc
			hasChanges = true
		}
		if cmd.Flags().Changed("image") {
			image, _ := cmd.Flags().GetString("image")
			req.Image = &image
			hasChanges = true
		}
		if cmd.Flags().Changed("is-public") {
			isPublic, _ := cmd.Flags().GetBool("is-public")
			req.IsPublic = &isPublic
			hasChanges = true
		}
		if cmd.Flags().Changed("public-port") {
			port, _ := cmd.Flags().GetInt("public-port")
			req.PublicPort = &port
			hasChanges = true
		}

		// Resource limits
		if cmd.Flags().Changed("limits-memory") {
			mem, _ := cmd.Flags().GetString("limits-memory")
			req.LimitsMemory = &mem
			hasChanges = true
		}
		if cmd.Flags().Changed("limits-cpus") {
			cpus, _ := cmd.Flags().GetString("limits-cpus")
			req.LimitsCpus = &cpus
			hasChanges = true
		}

		if !hasChanges {
			return fmt.Errorf("no fields to update")
		}

		// Validate is-public requires public-port
		if req.IsPublic != nil && *req.IsPublic {
			// If setting to public, check if port is provided or fetch current database to check existing port
			if req.PublicPort == nil || *req.PublicPort == 0 {
				client, err := getAPIClient(cmd)
				if err != nil {
					return fmt.Errorf("failed to get API client: %w", err)
				}

				dbService := service.NewDatabaseService(client)
				currentDB, err := dbService.Get(ctx, uuid)
				if err != nil {
					return fmt.Errorf("failed to get current database: %w", err)
				}

				// Check if database already has a public port
				if currentDB.PublicPort == nil || *currentDB.PublicPort == 0 {
					return fmt.Errorf("cannot set database as public without a public port. Please provide --public-port")
				}
			}
		}

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		err = dbService.Update(ctx, uuid, req)
		if err != nil {
			return fmt.Errorf("failed to update database: %w", err)
		}

		fmt.Println("Database updated successfully")
		return nil
	},
}

var deleteDatabaseCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete a database",
	Long:  `Delete a database and optionally clean up its configurations, volumes, and networks.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		force, _ := cmd.Flags().GetBool("force")
		deleteConfigurations, _ := cmd.Flags().GetBool("delete-configurations")
		deleteVolumes, _ := cmd.Flags().GetBool("delete-volumes")
		dockerCleanup, _ := cmd.Flags().GetBool("docker-cleanup")
		deleteConnectedNetworks, _ := cmd.Flags().GetBool("delete-connected-networks")

		if !force {
			fmt.Printf("Are you sure you want to delete database %s? (y/N): ", uuid)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Delete cancelled")
				return nil
			}
		}

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		err = dbService.Delete(ctx, uuid, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks)
		if err != nil {
			return fmt.Errorf("failed to delete database: %w", err)
		}

		fmt.Println("Database deleted successfully")
		return nil
	},
}

var startDatabaseCmd = &cobra.Command{
	Use:   "start <uuid>",
	Short: "Start a database",
	Long:  `Start a database by UUID.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		response, err := dbService.Start(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to start database: %w", err)
		}

		fmt.Println(response.Message)
		return nil
	},
}

var stopDatabaseCmd = &cobra.Command{
	Use:   "stop <uuid>",
	Short: "Stop a database",
	Long:  `Stop a database by UUID.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		response, err := dbService.Stop(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to stop database: %w", err)
		}

		fmt.Println(response.Message)
		return nil
	},
}

var restartDatabaseCmd = &cobra.Command{
	Use:   "restart <uuid>",
	Short: "Restart a database",
	Long:  `Restart a database by UUID.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		response, err := dbService.Restart(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to restart database: %w", err)
		}

		fmt.Println(response.Message)
		return nil
	},
}

// Backup commands

var backupsCmd = &cobra.Command{
	Use:     "backup",
	Aliases: []string{"backups"},
	Short:   "Manage database backups",
	Long:    `Manage database backup configurations and executions. All commands require the database UUID first to establish context.`,
}

var listBackupsCmd = &cobra.Command{
	Use:   "list <database_uuid>",
	Short: "List all backup configurations for a database",
	Long:  `List all backup configurations for a specific database.`,
	Args:  exactArgs(1, "<database_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbUUID := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		backups, err := dbService.ListBackups(ctx, dbUUID)
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		formatter, err := output.NewFormatter(Format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(backups)
	},
}

var updateBackupCmd = &cobra.Command{
	Use:   "update <database_uuid> <backup_uuid>",
	Short: "Update backup configuration",
	Long:  `Update a backup configuration settings (frequency, retention, S3, etc.). First UUID is the database, second is the specific backup configuration.`,
	Args:  exactArgs(2, "<database_uuid> <backup_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbUUID := args[0]
		backupUUID := args[1]

		req := &models.DatabaseBackupUpdateRequest{}
		hasChanges := false

		if cmd.Flags().Changed("enabled") {
			enabled, _ := cmd.Flags().GetBool("enabled")
			req.Enabled = &enabled
			hasChanges = true
		}
		if cmd.Flags().Changed("frequency") {
			freq, _ := cmd.Flags().GetString("frequency")
			req.Frequency = &freq
			hasChanges = true
		}
		if cmd.Flags().Changed("save-s3") {
			saveS3, _ := cmd.Flags().GetBool("save-s3")
			req.SaveS3 = &saveS3
			hasChanges = true
		}
		if cmd.Flags().Changed("s3-storage-uuid") {
			s3UUID, _ := cmd.Flags().GetString("s3-storage-uuid")
			req.S3StorageUUID = &s3UUID
			hasChanges = true
		}
		if cmd.Flags().Changed("databases-to-backup") {
			dbs, _ := cmd.Flags().GetString("databases-to-backup")
			req.DatabasesToBackup = &dbs
			hasChanges = true
		}
		if cmd.Flags().Changed("dump-all") {
			dumpAll, _ := cmd.Flags().GetBool("dump-all")
			req.DumpAll = &dumpAll
			hasChanges = true
		}

		// Retention settings
		if cmd.Flags().Changed("retention-amount-locally") {
			amount, _ := cmd.Flags().GetInt("retention-amount-locally")
			req.DatabaseBackupRetentionAmountLocally = &amount
			hasChanges = true
		}
		if cmd.Flags().Changed("retention-days-locally") {
			days, _ := cmd.Flags().GetInt("retention-days-locally")
			req.DatabaseBackupRetentionDaysLocally = &days
			hasChanges = true
		}
		if cmd.Flags().Changed("retention-max-storage-locally") {
			storage, _ := cmd.Flags().GetInt("retention-max-storage-locally")
			req.DatabaseBackupRetentionMaxStorageLocally = &storage
			hasChanges = true
		}
		if cmd.Flags().Changed("retention-amount-s3") {
			amount, _ := cmd.Flags().GetInt("retention-amount-s3")
			req.DatabaseBackupRetentionAmountS3 = &amount
			hasChanges = true
		}
		if cmd.Flags().Changed("retention-days-s3") {
			days, _ := cmd.Flags().GetInt("retention-days-s3")
			req.DatabaseBackupRetentionDaysS3 = &days
			hasChanges = true
		}
		if cmd.Flags().Changed("retention-max-storage-s3") {
			storage, _ := cmd.Flags().GetInt("retention-max-storage-s3")
			req.DatabaseBackupRetentionMaxStorageS3 = &storage
			hasChanges = true
		}

		if !hasChanges {
			return fmt.Errorf("no fields to update")
		}

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		err = dbService.UpdateBackup(ctx, dbUUID, backupUUID, req)
		if err != nil {
			return fmt.Errorf("failed to update backup: %w", err)
		}

		fmt.Println("Backup configuration updated successfully")
		return nil
	},
}

var createBackupCmd = &cobra.Command{
	Use:   "create <database_uuid>",
	Short: "Create a new scheduled backup configuration",
	Long: `Create a new scheduled backup configuration for a database. Configure frequency, retention, S3 storage, and other backup options.

Example: coolify database backup create abc123 --frequency "0 0 * * *" --enabled`,
	Args: exactArgs(1, "<database_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbUUID := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		req := &models.DatabaseBackupCreateRequest{}

		// Apply flags if provided
		if cmd.Flags().Changed("frequency") {
			frequency, _ := cmd.Flags().GetString("frequency")
			req.Frequency = &frequency
		}
		if cmd.Flags().Changed("enabled") {
			enabled, _ := cmd.Flags().GetBool("enabled")
			req.Enabled = &enabled
		}
		if cmd.Flags().Changed("save-s3") {
			saveS3, _ := cmd.Flags().GetBool("save-s3")
			req.SaveS3 = &saveS3
		}
		if cmd.Flags().Changed("s3-storage-uuid") {
			s3UUID, _ := cmd.Flags().GetString("s3-storage-uuid")
			req.S3StorageUUID = &s3UUID
		}
		if cmd.Flags().Changed("databases") {
			databases, _ := cmd.Flags().GetString("databases")
			req.DatabasesToBackup = &databases
		}
		if cmd.Flags().Changed("dump-all") {
			dumpAll, _ := cmd.Flags().GetBool("dump-all")
			req.DumpAll = &dumpAll
		}
		if cmd.Flags().Changed("retention-amount-local") {
			amount, _ := cmd.Flags().GetInt("retention-amount-local")
			req.DatabaseBackupRetentionAmountLocally = &amount
		}
		if cmd.Flags().Changed("retention-days-local") {
			days, _ := cmd.Flags().GetInt("retention-days-local")
			req.DatabaseBackupRetentionDaysLocally = &days
		}
		if cmd.Flags().Changed("retention-storage-local") {
			storage, _ := cmd.Flags().GetString("retention-storage-local")
			req.DatabaseBackupRetentionMaxStorageLocally = &storage
		}
		if cmd.Flags().Changed("retention-amount-s3") {
			amount, _ := cmd.Flags().GetInt("retention-amount-s3")
			req.DatabaseBackupRetentionAmountS3 = &amount
		}
		if cmd.Flags().Changed("retention-days-s3") {
			days, _ := cmd.Flags().GetInt("retention-days-s3")
			req.DatabaseBackupRetentionDaysS3 = &days
		}
		if cmd.Flags().Changed("retention-storage-s3") {
			storage, _ := cmd.Flags().GetString("retention-storage-s3")
			req.DatabaseBackupRetentionMaxStorageS3 = &storage
		}
		if cmd.Flags().Changed("timeout") {
			timeout, _ := cmd.Flags().GetInt("timeout")
			req.Timeout = &timeout
		}
		if cmd.Flags().Changed("disable-local") {
			disableLocal, _ := cmd.Flags().GetBool("disable-local")
			req.DisableLocalBackup = &disableLocal
		}

		dbService := service.NewDatabaseService(client)
		backup, err := dbService.CreateBackup(ctx, dbUUID, req)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		formatter, err := output.NewFormatter(Format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(backup)
	},
}

var deleteBackupCmd = &cobra.Command{
	Use:   "delete <database_uuid> <backup_uuid>",
	Short: "Delete backup configuration",
	Long:  `Delete a backup configuration and optionally all its executions from S3. First UUID is the database, second is the specific backup configuration.`,
	Args:  exactArgs(2, "<database_uuid> <backup_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbUUID := args[0]
		backupUUID := args[1]

		force, _ := cmd.Flags().GetBool("force")
		deleteS3, _ := cmd.Flags().GetBool("delete-s3")

		if !force {
			fmt.Printf("Are you sure you want to delete backup configuration %s? (y/N): ", backupUUID)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Delete cancelled")
				return nil
			}
		}

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		err = dbService.DeleteBackup(ctx, dbUUID, backupUUID, deleteS3)
		if err != nil {
			return fmt.Errorf("failed to delete backup: %w", err)
		}

		fmt.Println("Backup configuration deleted successfully")
		return nil
	},
}

var listBackupExecutionsCmd = &cobra.Command{
	Use:   "executions <database_uuid> <backup_uuid>",
	Short: "List backup executions",
	Long:  `List all executions for a backup configuration. First UUID is the database, second is the specific backup configuration.`,
	Args:  exactArgs(2, "<database_uuid> <backup_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbUUID := args[0]
		backupUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		executions, err := dbService.ListBackupExecutions(ctx, dbUUID, backupUUID)
		if err != nil {
			return fmt.Errorf("failed to list backup executions: %w", err)
		}

		formatter, err := output.NewFormatter(Format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(executions)
	},
}

var deleteBackupExecutionCmd = &cobra.Command{
	Use:   "delete-execution <database_uuid> <backup_uuid> <execution_uuid>",
	Short: "Delete backup execution",
	Long:  `Delete a specific backup execution and optionally from S3. First UUID is the database, second is the backup configuration, third is the specific execution.`,
	Args:  exactArgs(3, "<database_uuid> <backup_uuid> <execution_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbUUID := args[0]
		backupUUID := args[1]
		executionUUID := args[2]

		force, _ := cmd.Flags().GetBool("force")
		deleteS3, _ := cmd.Flags().GetBool("delete-s3")

		if !force {
			fmt.Printf("Are you sure you want to delete backup execution %s? (y/N): ", executionUUID)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Delete cancelled")
				return nil
			}
		}

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)
		err = dbService.DeleteBackupExecution(ctx, dbUUID, backupUUID, executionUUID, deleteS3)
		if err != nil {
			return fmt.Errorf("failed to delete backup execution: %w", err)
		}

		fmt.Println("Backup execution deleted successfully")
		return nil
	},
}

var backupNowCmd = &cobra.Command{
	Use:   "trigger <database_uuid> <backup_uuid>",
	Short: "Trigger immediate backup",
	Long:  `Trigger an immediate backup for a specific backup configuration. First UUID is the database, second is the specific backup configuration to trigger.`,
	Args:  exactArgs(2, "<database_uuid> <backup_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbUUID := args[0]
		backupUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		dbService := service.NewDatabaseService(client)

		// Trigger immediate backup by updating with backup_now flag
		req := &models.DatabaseBackupUpdateRequest{
			BackupNow: boolPtr(true),
		}

		err = dbService.UpdateBackup(ctx, dbUUID, backupUUID, req)
		if err != nil {
			return fmt.Errorf("failed to trigger backup: %w", err)
		}

		fmt.Println("Immediate backup triggered successfully")
		return nil
	},
}

func boolPtr(b bool) *bool {
	return &b
}

func init() {
	rootCmd.AddCommand(databasesCmd)
	databasesCmd.AddCommand(listDatabasesCmd)
	databasesCmd.AddCommand(getDatabaseCmd)
	databasesCmd.AddCommand(createDatabaseCmd)
	databasesCmd.AddCommand(updateDatabaseCmd)
	databasesCmd.AddCommand(deleteDatabaseCmd)
	databasesCmd.AddCommand(startDatabaseCmd)
	databasesCmd.AddCommand(stopDatabaseCmd)
	databasesCmd.AddCommand(restartDatabaseCmd)
	databasesCmd.AddCommand(backupsCmd)

	// Backup subcommands
	backupsCmd.AddCommand(listBackupsCmd)
	backupsCmd.AddCommand(createBackupCmd)
	backupsCmd.AddCommand(updateBackupCmd)
	backupsCmd.AddCommand(deleteBackupCmd)
	backupsCmd.AddCommand(backupNowCmd)
	backupsCmd.AddCommand(listBackupExecutionsCmd)
	backupsCmd.AddCommand(deleteBackupExecutionCmd)

	// Create command flags
	createDatabaseCmd.Flags().String("server-uuid", "", "Server UUID (required)")
	createDatabaseCmd.Flags().String("project-uuid", "", "Project UUID (required)")
	createDatabaseCmd.Flags().String("environment-name", "", "Environment name")
	createDatabaseCmd.Flags().String("environment-uuid", "", "Environment UUID")
	createDatabaseCmd.Flags().String("destination-uuid", "", "Destination UUID if server has multiple destinations")
	createDatabaseCmd.Flags().String("name", "", "Database name")
	createDatabaseCmd.Flags().String("description", "", "Database description")
	createDatabaseCmd.Flags().String("image", "", "Docker image")
	createDatabaseCmd.Flags().Bool("instant-deploy", false, "Deploy immediately after creation")
	createDatabaseCmd.Flags().Bool("is-public", false, "Make database publicly accessible")
	createDatabaseCmd.Flags().Int("public-port", 0, "Public port")
	createDatabaseCmd.Flags().String("limits-memory", "", "Memory limit (e.g., '512m', '2g')")
	createDatabaseCmd.Flags().String("limits-cpus", "", "CPU limit (e.g., '0.5', '2')")

	// PostgreSQL flags
	createDatabaseCmd.Flags().String("postgres-user", "", "PostgreSQL user")
	createDatabaseCmd.Flags().String("postgres-password", "", "PostgreSQL password")
	createDatabaseCmd.Flags().String("postgres-db", "", "PostgreSQL database name")

	// MySQL flags
	createDatabaseCmd.Flags().String("mysql-root-password", "", "MySQL root password")
	createDatabaseCmd.Flags().String("mysql-user", "", "MySQL user")
	createDatabaseCmd.Flags().String("mysql-password", "", "MySQL password")
	createDatabaseCmd.Flags().String("mysql-database", "", "MySQL database name")

	// MariaDB flags
	createDatabaseCmd.Flags().String("mariadb-root-password", "", "MariaDB root password")
	createDatabaseCmd.Flags().String("mariadb-user", "", "MariaDB user")
	createDatabaseCmd.Flags().String("mariadb-password", "", "MariaDB password")
	createDatabaseCmd.Flags().String("mariadb-database", "", "MariaDB database name")

	// MongoDB flags
	createDatabaseCmd.Flags().String("mongo-root-username", "", "MongoDB root username")
	createDatabaseCmd.Flags().String("mongo-root-password", "", "MongoDB root password")
	createDatabaseCmd.Flags().String("mongo-database", "", "MongoDB database name")

	// Redis flags
	createDatabaseCmd.Flags().String("redis-password", "", "Redis password")

	// KeyDB flags
	createDatabaseCmd.Flags().String("keydb-password", "", "KeyDB password")

	// Clickhouse flags
	createDatabaseCmd.Flags().String("clickhouse-admin-user", "", "Clickhouse admin user")
	createDatabaseCmd.Flags().String("clickhouse-admin-password", "", "Clickhouse admin password")

	// Dragonfly flags
	createDatabaseCmd.Flags().String("dragonfly-password", "", "Dragonfly password")

	// Update command flags - only common configuration options
	updateDatabaseCmd.Flags().String("name", "", "Database name")
	updateDatabaseCmd.Flags().String("description", "", "Database description")
	updateDatabaseCmd.Flags().String("image", "", "Docker image")
	updateDatabaseCmd.Flags().Bool("is-public", false, "Make database publicly accessible")
	updateDatabaseCmd.Flags().Int("public-port", 0, "Public port")
	updateDatabaseCmd.Flags().String("limits-memory", "", "Memory limit")
	updateDatabaseCmd.Flags().String("limits-cpus", "", "CPU limit")

	// Delete command flags
	deleteDatabaseCmd.Flags().Bool("delete-configurations", true, "Delete configurations")
	deleteDatabaseCmd.Flags().Bool("delete-volumes", true, "Delete volumes")
	deleteDatabaseCmd.Flags().Bool("docker-cleanup", true, "Run docker cleanup")
	deleteDatabaseCmd.Flags().Bool("delete-connected-networks", true, "Delete connected networks")

	// Backup create command flags
	createBackupCmd.Flags().String("frequency", "", "Backup frequency (cron expression, e.g., '0 0 * * *' for daily)")
	createBackupCmd.Flags().Bool("enabled", false, "Enable backup schedule")
	createBackupCmd.Flags().Bool("save-s3", false, "Save backups to S3")
	createBackupCmd.Flags().String("s3-storage-uuid", "", "S3 storage UUID")
	createBackupCmd.Flags().String("databases-to-backup", "", "Comma-separated list of databases to backup")
	createBackupCmd.Flags().Bool("dump-all", false, "Dump all databases")
	createBackupCmd.Flags().Int("retention-amount-locally", 0, "Number of backups to retain locally")
	createBackupCmd.Flags().Int("retention-days-locally", 0, "Days to retain backups locally")
	createBackupCmd.Flags().String("retention-max-storage-locally", "", "Max storage for local backups (e.g., '1GB', '500MB')")
	createBackupCmd.Flags().Int("retention-amount-s3", 0, "Number of backups to retain in S3")
	createBackupCmd.Flags().Int("retention-days-s3", 0, "Days to retain backups in S3")
	createBackupCmd.Flags().String("retention-max-storage-s3", "", "Max storage for S3 backups (e.g., '1GB', '500MB')")
	createBackupCmd.Flags().Int("timeout", 0, "Backup timeout in seconds")
	createBackupCmd.Flags().Bool("disable-local-backup", false, "Disable local backup storage")

	// Backup update command flags
	updateBackupCmd.Flags().Bool("enabled", false, "Enable or disable backup")
	updateBackupCmd.Flags().String("frequency", "", "Backup frequency (cron expression)")
	updateBackupCmd.Flags().Bool("save-s3", false, "Save backups to S3")
	updateBackupCmd.Flags().String("s3-storage-uuid", "", "S3 storage UUID")
	updateBackupCmd.Flags().String("databases-to-backup", "", "Comma-separated list of databases to backup")
	updateBackupCmd.Flags().Bool("dump-all", false, "Dump all databases")
	updateBackupCmd.Flags().Int("retention-amount-locally", 0, "Number of backups to retain locally")
	updateBackupCmd.Flags().Int("retention-days-locally", 0, "Days to retain backups locally")
	updateBackupCmd.Flags().Int("retention-max-storage-locally", 0, "Max storage for local backups (MB)")
	updateBackupCmd.Flags().Int("retention-amount-s3", 0, "Number of backups to retain in S3")
	updateBackupCmd.Flags().Int("retention-days-s3", 0, "Days to retain backups in S3")
	updateBackupCmd.Flags().Int("retention-max-storage-s3", 0, "Max storage for S3 backups (MB)")

	// Backup delete command flags
	deleteBackupCmd.Flags().Bool("delete-s3", false, "Delete backup files from S3")

	// Backup execution delete command flags
	deleteBackupExecutionCmd.Flags().Bool("delete-s3", false, "Delete backup file from S3")
}
