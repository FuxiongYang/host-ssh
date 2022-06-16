package remoteProtocol

import (
	"github.com/FuxiongYang/host-ssh/run"
	"github.com/spf13/cobra"
)

var pullFileCmd = &cobra.Command{
	Use: "pull",
	Short: "pull file from remote host to local",

	Run: PullFunc,
}

func init(){
	RootCmd.AddCommand(pullFileCmd)
	pullFileCmd.Flags().StringVar(&localFilePath, "local-file", "", "path of local host")
	pullFileCmd.Flags().StringVar(&remoteFilePath, "remote-file", "", "path of remote host")
}

func PullFunc(cmd *cobra.Command, args []string) {
	log.Info("host-ssh -t=pull remote-file=%s  local-path=%s", remoteFilePath, localFilePath)

	puser := run.NewUser(user, port, psw, force, encFlag)
	if host != "" {
		log.Info("[servers]=%s", host)
		run.SinglePull(host, puser, localFilePath, remoteFilePath, force)
	} else {
		run.ServersPull(localFilePath, remoteFilePath, puser, ipFile, force)
	}
}
