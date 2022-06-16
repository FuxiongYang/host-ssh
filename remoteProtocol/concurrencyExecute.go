package remoteProtocol

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/FuxiongYang/host-ssh/config"

	"github.com/FuxiongYang/host-ssh/run"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"
	"gopkg.in/cheggaaa/pb.v1"
)

var concExecuteCmd = &cobra.Command{
	Use:   "conc-execute",
	Short: "concurrency execute command by ssh",

	Run: ConcExecuteFunc,
}

var (
	isVBHMode   bool
	isKeepALive bool
	ipVBHFile   string

	commandsExecuteServerTotal int
	executeCountPerSession     int
	executeRate                int

	commandInterval int
	commandsList    []string

	passwordMode  = "password"
	publicKeyMode = "publicKey"
)

func init() {
	RootCmd.AddCommand(concExecuteCmd)

	concExecuteCmd.Flags().BoolVar(&isVBHMode, "vbh-switch", false, "open the Volc VBH mode, that can help to skip vbh command")
	concExecuteCmd.Flags().StringVar(&ipVBHFile, "vbh-host-ip-file", "", "vbh host ip")

	concExecuteCmd.Flags().BoolVar(&isKeepALive, "keep-a-live", false, "send message by long connection, default: false")

	concExecuteCmd.Flags().IntVarP(&executeRate, "rate", "r", 0, "Maximum range requests per second (0 is no limit)")
	concExecuteCmd.Flags().IntVar(&executeCountPerSession, "execute-count-per-session", 1, "Maximum execute requests per session (default 1)")
	concExecuteCmd.Flags().IntVar(&commandsExecuteServerTotal, "total", 1, "Total number of clients execute the commands, default: 1")

	concExecuteCmd.Flags().IntVar(&commandInterval, "command-interval", 0, "the interval time to sleep between executing each of commands, default: 0(ms)")
	concExecuteCmd.Flags().StringSliceVar(&commandsList, "commands", []string{"pwd"}, "commands need to execute in host, need to split by ',' between each, e.g. [ host-ssh conc-execute --commands \"ls -lh\",cd ]")
}

func ConcExecuteFunc(cmd *cobra.Command, args []string) {
	var hostsVBH []config.Host
	var err error
	var clientsPool chan *ssh.Client
	var wg sync.WaitGroup
	var executeMode string = "common"

	//preset generate ssh clientsPool
	t00 := time.Now()
	if host != "" {
		fmt.Printf("wait for generating clients pool...\n")
		log.Info("start to generate clientsPool from IP :[%s]", host)
		clientsPool = mustCreateClientsByIp(commandsExecuteServerTotal, host)
	} else {
		fmt.Printf("wait for generating clients pool...\n")
		log.Info("start to generate clientsPool from IP file:[%s]", ipFile)
		clientsPool = mustCreateClientsByIpFile(commandsExecuteServerTotal, ipFile)
	}
	close(clientsPool)
	t01 := time.Now()

	t1 := time.Now()
	bar = pb.New(len(clientsPool))
	bar.Format("Bom !")
	bar.Start()

	// deal with the problem of login with vbh instance
	lenthIpInUser := 0

	if isVBHMode {
		executeMode = "VBH"
		vbhHostUser := run.NewUser(user, port, psw, force, encFlag)
		if ipVBHFile != "" {
			hostsVBH, err = run.ParseIpfile(ipVBHFile, vbhHostUser)
			log.Debug("ips : %v", hostsVBH)
			if err != nil {
				log.Error("Parse %s error, error=%s", ipVBHFile, err)
			}
			lenthIpInUser = len(hostsVBH)
		}
	}

	// adjust rate of use client
	if executeRate == 0 {
		executeRate = math.MaxInt32
	}
	limit := rate.NewLimiter(rate.Limit(executeRate), 1)
	rand.Seed(time.Now().UnixNano())

	log.Warn("end of preset [start time is [%v], end time is [%v], time cost: [%v]], execute command by concurrency with %s mode\n", t00, t01, t01.Sub(t00), executeMode)
	fmt.Printf("end of preset[%v], start to execute command with %s mode... \n", t01.Sub(t00), executeMode)

	for client := range clientsPool {
		wg.Add(1)
		var (
			cuser = client.User()
			ip    = client.RemoteAddr().String()
		)
		go func(c *ssh.Client) {
			defer wg.Done()
			limit.Wait(context.Background())

			tty, err := run.NewConnectionWithTTY(c)
			if err != nil {
				log.Error("init tty failed, [%s]\n", err.Error())
				return
			}
			defer tty.Close()

			welcome, err := tty.RunCommand("")
			if err != nil {
				log.Error("error to welcome, error: [%s]", err.Error())
				return
			}
			log.Info("init tty successfully...")
			log.Debug("init tty successfully, %s\n", welcome)
			if isVBHMode {
				log.Info("")
				if lenthIpInUser != 0 {
					index := rand.Intn(lenthIpInUser)
					ipInUser := hostsVBH[index].Ip
					out, err := tty.RunCommand("s " + ipInUser)
					if err != nil {
						log.Error("error to search host in vbh instance with user, error: [%s]", err.Error())
						return
					}
					log.Debug("search host [%s], messages %s\n", ipInUser, out)
				}

				out, err := tty.RunCommand("1")
				if err != nil {
					log.Error("error to login host in vbh instance with user, error: [%s]", err.Error())
					return
				}
				log.Debug("login host through vbh, messages:\n %s\n", out)
			}
			executeCommands(tty, executeCountPerSession, commandsList, cuser, ip, commandInterval)

			bar.Increment()
		}(client)

	}
	bar.Finish()
	wg.Wait()

	t2 := time.Now()
	fmt.Printf("end of execution [%v] \n", t2.Sub(t1))
	log.Info("start time is[%v], end time is [%v], time cost: [%v]", t1, t2, t2.Sub(t1))
}

func executeCommands(tty *run.SSHConnectionWithTTY, count int, cmds []string, user, ip string, interval int) {
	for i := 0; i < count; i++ {
		executeCommand := cmds[i%len(cmds)]
		log.Info("[%d]execute command:[%s] by user:[%s] in remote:[%s]", i, executeCommand, user, ip)
		out, err := tty.RunCommand(executeCommand)
		if err != nil {
			log.Error("error of executing Command [] by user:[%s] in remote:[%s]\", message: [%s]\n", executeCommand, user, ip)
			continue
		}

		log.Debug("[%d]execute command:[%d] by user:[%s] in remote:[%s], result:\n out\n", executeCommand, user, ip, out)
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func mustCreateClientsByIp(total int, ip string) chan *ssh.Client {
	var clients = make(chan *ssh.Client, total)

	for i := 0; i < total; i++ {
		client, err := run.CreateSSHClient(user, ip, port, passwordMode, psw)
		if err != nil {
			log.Error("error of creating ssh client, [%s]", err.Error())
		}
		fmt.Printf("generate client[%d], msg: [%v]\n", i+1, client)
		clients <- client
	}
	return clients
}

func mustCreateClientsByIpFile(total int, ipFile string) chan *ssh.Client {
	var clients = make(chan *ssh.Client, total)
	puser := run.NewUser(user, port, psw, force, encFlag)
	hosts, err := run.ParseIpfile(ipFile, puser)
	if err != nil {
		log.Error("Parse %s error, error=%s", ipFile, err)
		return nil
	}
	lenth := len(hosts)

	for i := 0; i < total; i++ {
		h := hosts[i%lenth]
		client, err := run.CreateSSHClient(h.User, h.Ip, h.Port, passwordMode, h.Psw)
		if err != nil {
			log.Error("error of creating ssh client, [%s]", err.Error())
			fmt.Printf("generate client[%d] error, Err msg: [%v]\n", i+1, err)
			continue
		}
		fmt.Printf("generate client[%d], msg: [%v]\n", i+1, client)
		clients <- client
	}
	return clients
}
