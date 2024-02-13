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
		Short: "Inserts a master record.",
		Args:  cobra.ExactArgs(4),
		Run:   handlers.Repo.InsertMaster,
	}

	var cmdCalcM = &cobra.Command{
		Use:   "calc-m [id]",
		Short: "Prints the number of entries in a master table",
		Args:  cobra.MaximumNArgs(1),
		Run:   handlers.Repo.CalcMaster,
	}

	var cmdCalcS = &cobra.Command{
		Use:   "calc-s",
		Short: "Prints the number of entries in a slave table",
		Args:  cobra.ExactArgs(0),
		Run:   handlers.Repo.CalcMaster,
	}

	var cmdUtM = &cobra.Command{
		Use:   "ut-m",
		Short: "Prints all the entries of the master table",
		Args:  cobra.ExactArgs(0),
		Run:   handlers.Repo.UtMaster,
	}

	var cmdGetM = &cobra.Command{
		Use:   "get-m",
		Short: "Prints all the entries of the master table",
		Run:   handlers.Repo.GetMasterAll,
	}

	//var cmdGetM = &cobra.Command{
	//	Use:   "get-m <id> [field_name]",
	//	Short: "Prints all the entries of the master table.",
	//	Args:  cobra.MinimumNArgs(1),
	//	Run:   handlers.Repo.GetMaster,
	//}

	//var cmdInsertS = &cobra.Command{
	//	Use:   "insert-s <id> <master_id> <issued_to>",
	//	Short: "Inserts a master record.",
	//	Args:  cobra.ExactArgs(3),
	//	Run:   handlers.Repo.InsertSlave,
	//}

	rootCmd.AddCommand(cmdInsertM)
	rootCmd.AddCommand(cmdCalcM)
	rootCmd.AddCommand(cmdUtM)
	rootCmd.AddCommand(cmdCalcS)

	rootCmd.AddCommand(cmdGetM)
	cmdGetM.Flags().Bool("all", false, "If set, prints all entries of the master table")

	return rootCmd
}
