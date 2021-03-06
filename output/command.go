// Copyright 2018 github.com/FuxiongYang/host-remoteProtocol Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Author: andes
// Email: email.tata@qq.com

package output

import (
	//	"context"
	"fmt"
	"github.com/FuxiongYang/host-ssh/logs"
	"github.com/FuxiongYang/host-ssh/machine"

	//	"strings"
	"sync"
	"time"
)

const (
	TIMEOUT = 4500
)

var (
	log = logs.NewLogger()
)

//new print result
func Print(res machine.Result) {
	//fmt.Printf("ip=%s\n", res.Ip)
	//index := strings.Index(cmd, ";")
	//newcmd := cmd[index+1:]
	//fmt.Printf("ip=%s|command=%s\n", ip, cmd)
	//fmt.Printf("command=%s\n", res.Cmd)
	if res.Err != nil {
		//fmt.Printf("return=1\n")
		log.Error("return=1\n %s\n", res.Err)
		//fmt.Printf("%s\n", res.Err)
	} else {
		//fmt.Printf("return=0\n")
		log.Debug("return=0\n %s\n", res.Result)
		//fmt.Printf("%s\n", res.Result)
	}
	//fmt.Println("----------------------------------------------------------")
}

func PrintResults2(crs chan machine.Result, ls int, wt *sync.WaitGroup, ccons chan struct{}, timeout int) {
	if timeout == 0 {
		timeout = TIMEOUT
	}

	for i := 0; i < ls; i++ {
		select {
		case rs := <-crs:
			//PrintResult(rs.Ip, rs.Cmd, rs.Result)
			//log.Debug("execute result is %v\n", rs)
			Print(rs)
		case <-time.After(time.Second * time.Duration(timeout)):
			fmt.Printf("getSSHClient error,SSH-Read-TimeOut,Timeout=%ds", timeout)
		}
		wt.Done()
		<-ccons
	}

}

//print push file result
func PrintPushResult(ip, src, dst string, err error) {
	fmt.Println("ip=", ip)
	fmt.Println("command=", "scp "+src+" root@"+ip+":"+dst)
	if err != nil {
		fmt.Printf("return=1\n")
		fmt.Println(err)
	} else {
		fmt.Printf("return=0\n")
		fmt.Printf("Push %s to %s ok.\n", src, dst)
	}
	fmt.Println("----------------------------------------------------------")
}

//print pull result
func PrintPullResult(ip, src, dst string, err error) {
	fmt.Println("ip=", ip)
	fmt.Println("command=", "scp "+" root@"+ip+":"+dst+" "+src)
	if err != nil {
		fmt.Printf("return=1\n")
		fmt.Println(err)
	} else {
		fmt.Printf("return=0\n")
		fmt.Printf("Pull from %s to %s ok.\n", dst, src)
	}
	fmt.Println("----------------------------------------------------------")
}
