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

	var cmdGetM = &cobra.Command{
		Use:   "get-m <id> [field_name]",
		Short: "Retrieves specific entries from the master table.",
		Args:  cobra.MinimumNArgs(1),
		Run:   handlers.Repo.GetMaster,
	}

	// Uncomment and correct this command if needed.
	// var cmdInsertS = &cobra.Command{
	//     Use:   "insert-s <id> <master_id> <issued_to>",
	//     Short: "Inserts a record into the slave table.",
	//     Args:  cobra.ExactArgs(3),
	//     Run:   handlers.Repo.InsertSlave,
	// }

	rootCmd.AddCommand(cmdInsertM)
	rootCmd.AddCommand(cmdCalcM)
	rootCmd.AddCommand(cmdUtM)
	rootCmd.AddCommand(cmdGetM)
	rootCmd.AddCommand(cmdCalcS)

	return rootCmd
}
