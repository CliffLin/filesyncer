package main

import (
	"filesyncer/internal/syncer"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	Paths []path `mapstructure:"paths"`
}

type path struct {
	Remote string      `mapstructure:"remote"`
	Local  string      `mapstructure:"local"`
	Type   string      `mapstructure:"type"`
	Config interface{} `mapstructure:"config"`
}

var cfg config
var cfgFile string

var rootCmd = &cobra.Command{
	Use: "filesyncer",
	Run: func(cmd *cobra.Command, args []string) {
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}

		if err := viper.Unmarshal(&cfg); err != nil {
			panic(err)
		}

		for _, path := range cfg.Paths {
			go func(remote, local, syncer_type string, config interface{}) {
				syncer, err := syncer.GetSyncer(syncer_type, local, remote, config)
				if err != nil {
					log.Println(err)
					return
				}
				syncer.Run()
			}(path.Remote, path.Local, path.Type, path.Config)
		}
		<-make(chan struct{})
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.filesyncer.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".filesyncer")
	}

	viper.AutomaticEnv()

}

func main() {
	rootCmd.Execute()
}
