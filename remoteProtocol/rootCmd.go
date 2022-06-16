package remoteProtocol

import (
	"fmt"

	"github.com/FuxiongYang/host-ssh/enc"
	"gopkg.in/cheggaaa/pb.v1"

	//"github.com/FuxiongYang/host-ssh/enc"
	//"github.com/FuxiongYang/host-ssh/help"
	"github.com/FuxiongYang/host-ssh/logs"
	"github.com/FuxiongYang/host-ssh/tools"

	"github.com/spf13/cobra"

	"path/filepath"
	"strings"
)

var RootCmd = &cobra.Command{
	Use:   "host-shh",
	Short: "A Smart Host Remote Connection Tool",
	Long:  " host-ssh is a smart host remote connection tool.It is developed by Go,compiled into a separate binary without any dependencies.",

	PersistentPreRun:  Run,
	PersistentPostRun: PostRun,
}

var (
	//
	bar *pb.ProgressBar

	//common options
	port     string
	host     string
	user     string
	psw      string
	prunType string

	//batch running options
	ipFile string
	cons   int

	//safe options
	encFlag   bool
	force     bool
	psafe     bool
	pkey      string
	blackList = []string{"rm", "mkfs", "mkfs.ext3", "make.ext2", "make.ext4", "make2fs", "shutdown", "reboot", "init", "dd"}

	//log options
	plogLevel string
	plogPath  string
	log       = logs.NewLogger()
	logFile   = "host-ssh.log"

	pversion bool

	//Timeout
	ptimeout int
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&port, "port", "P", "22", "remoteProtocol port")
	RootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "remoteProtocol ip")
	RootCmd.PersistentFlags().StringVarP(&user, "user", "U", "root", "remoteProtocol user")
	RootCmd.PersistentFlags().StringVarP(&psw, "password", "p", "", "remoteProtocol password")
	//RootCmd.PersistentFlags().StringVar(&prunType, "t", "cmd", "running mode: cmd|push|pull")

	RootCmd.PersistentFlags().StringVar(&ipFile, "ip-file", "ip.txt", "ip file when batch running mode")
	RootCmd.PersistentFlags().IntVarP(&cons, "conns", "c", 30, "the number of concurrency when b")

	RootCmd.PersistentFlags().BoolVar(&encFlag, "password-encrypt", false, "password is Encrypted")
	RootCmd.PersistentFlags().BoolVar(&force, "force-on-nosafe", false, "force to run even if it is not safe")
	RootCmd.PersistentFlags().BoolVarP(&psafe, "exit-on-error-occurs", "s", false, "if -s is setting, host-ssh will exit when error occurs")
	RootCmd.PersistentFlags().StringVar(&pkey, "key", "", "aes key for password decrypt and encryption")

	RootCmd.PersistentFlags().StringVarP(&plogLevel, "log-level", "l", "info", "log level (debug|info|warn|error")
	RootCmd.PersistentFlags().StringVar(&plogPath, "log-path", "./log/", "logfile path")

	//RootCmd.PersistentFlags().BoolVar(&pversion, "version", false, "host-ssh version")

	RootCmd.PersistentFlags().IntVar(&ptimeout, "timeout", 10, "remoteProtocol timeout setting")
}

//main
func Run(cmd *cobra.Command, args []string) {
	//usage := func() {
	//	fmt.Println(help.Help)
	//}
	//
	//flag.Parse()
	//
	//version
	//if pversion {
	//	fmt.Println(AppVersion)
	//	return
	//}

	if pkey != "" {
		enc.SetKey([]byte(pkey))
	}
	//
	//if flag.NArg() < 1 {
	//	usage()
	//	return
	//}
	//
	//if prunType == "" || flag.Arg(0) == "" {
	//	usage()
	//	return
	//}

	if err := initLog(); err != nil {
		fmt.Printf("init log error:%s\n", err)
		return
	}

	//异步日志，需要最后刷新和关闭
	//defer func() {
	//	log.Flush()
	//	log.Close()
	//}()

	log.Debug("parse flag ok , init log setting ok.")

	//switch prunType {
	////run command on remote server
	////case "cmd":
	////	if flag.NArg() != 1 {
	////		//usage()
	////		return
	////	}
	////
	////	cmd := flag.Arg(0)
	////
	////	if flag := tools.CheckSafe(cmd, blackList); !flag && force == false {
	////		fmt.Printf("Dangerous command in %s", cmd)
	////		fmt.Printf("You can use the `-f` option to force to excute")
	////		log.Error("Dangerous command in %s", cmd)
	////		return
	////	}
	////
	////	puser := run.NewUser(user, port, psw, force, encFlag)
	////	log.Info("host-ssh -t=cmd  cmd=[%s]", cmd)
	////
	////	if host != "" {
	////		log.Info("[servers]=%s", host)
	////		run.SingleRun(host, cmd, puser, force, ptimeout)
	////
	////	} else {
	////		cr := make(chan machine.Result)
	////		ccons := make(chan struct{}, cons)
	////		wg := &sync.WaitGroup{}
	////		run.ServersRun(cmd, puser, wg, cr, ipFile, ccons, psafe, ptimeout)
	////		wg.Wait()
	////	}
	//
	////push file or dir  to remote server
	////case "scp", "push":
	////
	////	if flag.NArg() != 2 {
	////		//usage()
	////		return
	////	}
	////
	////	src := flag.Arg(0)
	////	dst := flag.Arg(1)
	////	log.Info("host-ssh -t=push local-file=%s, remote-path=%s", src, dst)
	////
	////	puser := run.NewUser(user, port, psw, force, encFlag)
	////	if host != "" {
	////		log.Info("[servers]=%s", host)
	////		run.SinglePush(host, src, dst, puser, force, ptimeout)
	////	} else {
	////		cr := make(chan machine.Result, 20)
	////		ccons := make(chan struct{}, cons)
	////		wg := &sync.WaitGroup{}
	////		run.ServersPush(src, dst, puser, ipFile, wg, ccons, cr, ptimeout)
	////		wg.Wait()
	////	}
	//
	////pull file from remote server
	////case "pull":
	////	if flag.NArg() != 2 {
	////		//usage()
	////		return
	////	}
	////
	////	//本地目录
	////	src := flag.Arg(1)
	////	//远程文件
	////	dst := flag.Arg(0)
	////	log.Info("host-ssh -t=pull remote-file=%s  local-path=%s", dst, src)
	////
	////	puser := run.NewUser(user, port, psw, force, encFlag)
	////	if host != "" {
	////		log.Info("[servers]=%s", host)
	////		run.SinglePull(host, puser, src, dst, force)
	////	} else {
	////		run.ServersPull(src, dst, puser, ipFile, force)
	////	}
	//
	//default:
	//	//usage()
	//}
}

func PostRun(cmd *cobra.Command, args []string) {
	fmt.Printf("Flush and close log\n")
	log.Flush()
	log.Close()
}

//setting log
func initLog() error {
	switch plogLevel {
	case "debug":
		log.SetLevel(logs.LevelDebug)
	case "error":
		log.SetLevel(logs.LevelError)
	case "info":
		log.SetLevel(logs.LevelInformational)
	case "warn":
		log.SetLevel(logs.LevelWarning)
	default:
		log.SetLevel(logs.LevelInformational)
	}

	logpath := plogPath
	err := tools.MakePath(logpath)
	if err != nil {
		return err
	}

	logname := filepath.Join(logpath, logFile)
	logstring := `{"filename":"` + logname + `"}`

	//此处主要是处理windows下文件路径问题,不做转义，日志模块会报如下错误
	//logs.BeeLogger.SetLogger: invalid character 'g' in string escape code
	logstring = strings.Replace(logstring, `\`, `\\`, -1)

	err = log.SetLogger("file", logstring)
	if err != nil {
		return err
	}
	//开启日志异步提升性能
	log.Async()
	return nil
}
