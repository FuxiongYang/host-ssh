package remoteProtocol

import (
	"fmt"
	"github.com/FuxiongYang/host-ssh/machine"
	"github.com/FuxiongYang/host-ssh/run"
	"github.com/FuxiongYang/host-ssh/tools"
	"github.com/spf13/cobra"
	"sync"
)

var cmdCmd = &cobra.Command{
	Use: "execute",
	Short: "execute command by ssh",

	Run: CmdFunc,
}

var (
	command string
)

func init(){
	RootCmd.AddCommand(cmdCmd)
	cmdCmd.Flags().StringVar(&command, "command", "", "command need to be executed in remote")
}

func CmdFunc(cmd *cobra.Command, args []string){
	if flag := tools.CheckSafe(command, blackList); !flag && force == false {
		fmt.Printf("Dangerous command in %s", command)
		fmt.Printf("You can use the `-f` option to force to excute")
		log.Error("Dangerous command in %s", command)
		return
	}

	puser := run.NewUser(user, port, psw, force, encFlag)
	log.Info("host-ssh execute cmd=[%s]", command)

	if host != "" {
		log.Info("[servers]=%s", host)
		run.SingleRun(host, command, puser, force, ptimeout)

	} else {
		cr := make(chan machine.Result)
		ccons := make(chan struct{}, cons)
		wg := &sync.WaitGroup{}
		run.ServersRun(command, puser, wg, cr, ipFile, ccons, psafe, ptimeout)
		wg.Wait()
	}
}