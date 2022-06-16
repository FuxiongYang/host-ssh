package remoteProtocol

import (
	"github.com/FuxiongYang/host-ssh/machine"
	"github.com/FuxiongYang/host-ssh/run"
	"github.com/spf13/cobra"
	"sync"
)

var pushFileCmd = &cobra.Command{
	Use: "push",
	Short: "push local file to remote file",

	Run: PushFunc,
}

var (
	localFilePath  string
	remoteFilePath string
)

func init(){
	RootCmd.AddCommand(pushFileCmd)
	pushFileCmd.Flags().StringVar(&localFilePath, "local-file", "", "path of local host")
	pushFileCmd.Flags().StringVar(&remoteFilePath, "remote-file", "", "path of remote host")
}

func PushFunc(cmd *cobra.Command, args []string){

	log.Info("host-ssh -t=push local-file=%s, remote-path=%s", localFilePath, remoteFilePath)

	puser := run.NewUser(user, port, psw, force, encFlag)
	if host != "" {
		log.Info("[servers]=%s", host)
		run.SinglePush(host, localFilePath, remoteFilePath, puser, force, ptimeout)
	} else {
		cr := make(chan machine.Result, 20)
		ccons := make(chan struct{}, cons)
		wg := &sync.WaitGroup{}
		run.ServersPush(localFilePath, remoteFilePath, puser, ipFile, wg, ccons, cr, ptimeout)
		wg.Wait()
	}
}