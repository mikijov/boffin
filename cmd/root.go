/*
Copyright (C) 2019 Milutin Jovanvović

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package cmd ...
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var dbDir string
var dryRun bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "boffin",
	Short: "A brief description of your application",
	Long: `Copyright (C) 2020 by Milutin Jovanvović. This program comes with ABSOLUTELY NO
WARRANTY. This is free software, and you are welcome to redistribute it under
certain conditions. See LICENSE for details.

Boffin is a utility that helps collect files and file changes from multiple
sources while keeping only the most recent copy of the file, as well as keeping
destination directory structure.

Example use case would be to collect photos from multiple devices. Once copied
to the repository, Boffin keeps track of the file changes, so if the file is
changed, renamed or moved in the repository, it will not be imported again in
the future.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

func stderr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		stderr("ERROR: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	log.SetFlags(0)

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.boffin)")
	rootCmd.PersistentFlags().StringVar(&dbDir, "db-dir", "", "db directory if out of BASE (default is BASE_DIR/.boffin)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "do not make any changed to files")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			stderr("ERROR: %v\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".boffin" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".boffin")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		stderr("Using config file: %s\n", viper.ConfigFileUsed())
	}
}
