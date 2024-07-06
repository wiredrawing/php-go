/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var toggleBool bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "phpgo",
	//Short: "Short",
	//Long:  `Long`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Printf("------------------------------>")
		//fmt.Printf("args %v", args)
		//fmt.Printf("php-path %v", cmd.Flag("toggle").Value)
	},
	Version: "0.0.1",
}

//var rootCmd = &cobra.Command{
//	//Use:   "--php-path",
//	Short: "A brief description of your applicatioF",
//	Long: `A longer description that spans multiple lines and likely contains
//examples and usage of using your application. For example:
//
//Cobra is a CLI library for Go that empowers applications.
//This application is a tool to generate the needed files
//to quickly create a Cobra application.`,
//	// Uncomment the following line if your bare application]
//	// has an action associated with it:
//	Run: func(cmd *cobra.Command, args []string) {
//		//fmt.Printf("args %v", args)
//		//fmt.Printf("php-path %v", cmd.Flag("toggle").Value)
//	},
//}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
type CommandLineOption struct {
	Phppath      string
	Surveillance string
	Message      string
	Prompt       string
	Version      bool
	Help         bool
	Toggle       bool
	SaveFileName string
	EditorPath   string
}

func Execute() CommandLineOption {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	//help := rootCmd.Flag("help")
	//fmt.Printf("help %v", help.Value.String())
	// (1)version
	version, _ := rootCmd.Flags().GetBool("version")
	// (2)help
	help, _ := rootCmd.Flags().GetBool("help")
	// (3)toggle
	toggle, _ := rootCmd.Flags().GetBool("toggle")
	//fmt.Printf("version value:%[1]v type: %[1]T \n", version)
	//fmt.Printf("help value:%[1]v type: %[1]T  \n", help)
	//fmt.Printf("toggle value:%[1]v type: %[1]T  \n", toggle)
	// (4)phppath
	phppath, _ := rootCmd.Flags().GetString("phppath")
	// (5)surveillance
	surveillance, _ := rootCmd.Flags().GetString("surveillance")
	// (6)message
	message, _ := rootCmd.Flags().GetString("message")
	// (7)prompt
	prompt, _ := rootCmd.Flags().GetString("prompt")
	// (8)save-file-name
	saveFileName, _ := rootCmd.Flags().GetString("save-file-name")
	// (9)editor-path
	editorPath, _ := rootCmd.Flags().GetString("editor-path")
	config := make(map[string]string, 10)
	config["toggle"] = fmt.Sprintf("%v", toggle)
	config["phppath"] = phppath
	config["surveillance"] = surveillance
	config["message"] = message
	config["prompt"] = prompt

	var commandLineOption CommandLineOption = CommandLineOption{
		Phppath:      phppath,
		Surveillance: surveillance,
		Message:      message,
		Prompt:       prompt,
		Version:      version,
		Help:         help,
		Toggle:       toggle,
		SaveFileName: saveFileName,
		EditorPath:   editorPath,
	}
	//config["saveFileName"] = rootCmd.Flag("save-file-name").Value.String()
	//for key, value := range config {
	//	fmt.Printf("%v << %v >>\r\n", key, value)
	//}
	return commandLineOption
	//return config
}
func init() {
	rootCmd.PersistentFlags().BoolVarP(&toggleBool, "toggle", "t", false, "boolean型で受け取る")
	//rootCmd.PersistentFlags().BoolP("help", "l", false, "Help message for toggle")
	rootCmd.PersistentFlags().String("save-file-name", "save.php", "Input file name to save on the working directory.")
	rootCmd.PersistentFlags().StringP("phppath", "e", "php", "PHP Path")
	rootCmd.PersistentFlags().StringP("surveillance", "s", "", "指定したPHPファイルを監視する(※空文字の場合は,監視しない)")
	rootCmd.PersistentFlags().StringP("message", "m", "Help message for toggle", "Input the message")
	rootCmd.PersistentFlags().StringP("prompt", "p", ">>>", "Input the prompt")
	rootCmd.PersistentFlags().StringP("editor-path", "d", "", "Input the editor path")
}
