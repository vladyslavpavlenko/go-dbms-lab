package main

import (
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/config"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/handlers"
)

func commands(app *config.AppConfig) *cobra.Command {
	repo := handlers.NewRepo(app)
	handlers.NewHandlers(repo)
	var rootCmd = &cobra.Command{}

	var cmdInsertM = &cobra.Command{
		Use:   "insert-m <id> <title> <category> <instructor>",
		Short: "Inserts a record into the master table.",
		Args:  cobra.ExactArgs(4),
		Run:   handlers.Repo.InsertMaster,
	}

	var cmdInsertS = &cobra.Command{
		Use:   "insert-s <id> <course_id> <issued_to>",
		Short: "Inserts a record into the slave table.",
		Args:  cobra.ExactArgs(3),
		Run:   handlers.Repo.InsertSlave,
	}

	var cmdCalcM = &cobra.Command{
		Use:   "calc-m [id]",
		Short: "Calculates the number of entries in the master table.",
		Args:  cobra.MaximumNArgs(1),
		Run:   handlers.Repo.CalcMaster,
	}

	var cmdCalcS = &cobra.Command{
		Use:   "calc-s",
		Short: "Calculates the number of entries in the slave table.",
		Args:  cobra.ExactArgs(0),
		Run:   handlers.Repo.CalcSlave,
	}

	var cmdUtM = &cobra.Command{
		Use:   "ut-m",
		Short: "Prints all entries of the master table.",
		Args:  cobra.ExactArgs(0),
		Run:   handlers.Repo.UtMaster,
	}

	var cmdUtS = &cobra.Command{
		Use:   "ut-s",
		Short: "Prints all entries of the slave table.",
		Args:  cobra.ExactArgs(0),
		Run:   handlers.Repo.UtSlave,
	}

	var cmdGetM = &cobra.Command{
		Use:   "get-m <id> [field_name]",
		Short: "Retrieves specific entries from the master table.",
		Args:  cobra.MinimumNArgs(1),
		Run:   handlers.Repo.GetMaster,
	}

	var cmdGetS = &cobra.Command{
		Use:   "get-s <id> <course_id> [field_name]",
		Short: "Retrieves specific entries from the master table.",
		Args:  cobra.MinimumNArgs(1),
		Run:   handlers.Repo.GetSlave,
	}

	var cmdUpdateM = &cobra.Command{
		Use:   "update-m <id> <title> <category> <instructor>",
		Short: "Updates fields of a record accessed by its ID.",
		Args:  cobra.MinimumNArgs(2),
		Run:   handlers.Repo.UpdateMaster,
	}

	var cmdUpdateS = &cobra.Command{
		Use:   "update-s <id> <issued_to>",
		Short: "Updates fields of a record accessed by its ID.",
		Args:  cobra.MinimumNArgs(2),
		Run:   handlers.Repo.UpdateSlave,
	}

	var cmdDeleteM = &cobra.Command{
		Use:   "del-m <id>",
		Short: "Deletes entry by its ID.",
		Args:  cobra.MinimumNArgs(1),
		Run:   handlers.Repo.DeleteMaster,
	}

	rootCmd.AddCommand(cmdInsertM)
	rootCmd.AddCommand(cmdCalcM)
	rootCmd.AddCommand(cmdUtM)
	rootCmd.AddCommand(cmdGetM)
	rootCmd.AddCommand(cmdUpdateM)
	rootCmd.AddCommand(cmdDeleteM)

	rootCmd.AddCommand(cmdInsertS)
	rootCmd.AddCommand(cmdCalcS)
	rootCmd.AddCommand(cmdUtS)
	rootCmd.AddCommand(cmdGetS)
	rootCmd.AddCommand(cmdUpdateS)
	//rootCmd.AddCommand(cmdDeleteS)

	return rootCmd
}
