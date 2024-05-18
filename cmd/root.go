/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	//Use:   "--php-path",
	Short: "A brief description of your applicatioF",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application]
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Printf("args %v", args)
		//fmt.Printf("php-path %v", cmd.Flag("toggle").Value)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() map[string]string {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	toggle := rootCmd.Flag("toggle")
	phppath := rootCmd.Flag("phppath")
	surveillance := rootCmd.Flag("surveillance")
	message := rootCmd.Flag("message")
	config := make(map[string]string)
	config["toggle"] = toggle.Value.String()
	config["phppath"] = phppath.Value.String()
	config["surveillance"] = surveillance.Value.String()
	config["message"] = message.Value.String()
	config["prompt"] = rootCmd.Flag("prompt").Value.String()
	for key, value := range config {
		fmt.Printf("%v << %v >>", key, value)
	}
	return config
	//fmt.Printf("toggle << %v >>", toggle.Value)
	//fmt.Printf("phppath << %v >>", phppath.Value)
	//fmt.Printf("rootCmd.Execute() %v", rootCmd.Flags())
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.php-go.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//rootCmd.Flags().String("phppath", "php", "PHP Path")
	//rootCmd.Flags().StringP("surveillance", "s", "", "指定したPHPファイルを監視する")
	rootCmd.PersistentFlags().String("phppath", "php", "PHP Path")
	rootCmd.PersistentFlags().StringP("surveillance", "s", "", "指定したPHPファイルを監視する(※空文字の場合は,監視しない)")
	rootCmd.PersistentFlags().StringP("message", "m", "Help message for toggle", "Input the message")
	rootCmd.PersistentFlags().StringP("prompt", "p", " $ ", "Input the prompt")
}
