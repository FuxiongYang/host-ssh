package remoteProtocol

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use: "version",
	Short: "host-shh version",

	Run: func(cmd *cobra.Command, args []string){
		fmt.Println(AppVersion)
	},
}


func init() {
	RootCmd.AddCommand(versionCmd)
}

