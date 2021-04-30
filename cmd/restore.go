/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"

	sqlite "github.com/gentoomaniac/backup-tool/pkg/db/sqlite"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore a backup",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("restore called")
		db, _ := cmd.Flags().GetString("db")
		backupID, _ := cmd.Flags().GetInt32("backup")
		database, _ := sqlite.InitDB(db)

		fmt.Println(int(backupID))
		backup := sqlite.GetBackupBackupById(database, int(backupID))

		log.Printf("Backup: %s - %s - %d\n", backup.Name, backup.Description, backup.Timestamp)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringP("db", "d", "backup.db", "Database file with backup meta information")
	viper.BindPFlag("db", restoreCmd.Flags().Lookup("db"))

	restoreCmd.Flags().Int32P("backup", "b", 0, "Backup to restore")
	restoreCmd.MarkFlagRequired("backup")
}
