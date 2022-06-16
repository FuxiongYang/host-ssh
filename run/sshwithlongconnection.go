package run

import (
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

const (
	// SSHPort define ssh port.
	SSHPort      = "22"
	tcpDialRetry = 3
	timeOut      = 10 * time.Second
)

var err error

type Conn interface {
	connect() error
	RunCommand() (string, error)
	Close()
}

// CreateSSHClient create ssh client
func CreateSSHClient(user, host, port, authType, authValue string) (*ssh.Client, error){
	config := &ssh.ClientConfig{
		User: user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	switch strings.ToUpper(authType) {
	case "PASSWORD":
		config.Auth = []ssh.AuthMethod{ssh.Password(authValue)}
	case "PUBLICKEY":
		var publicKey ssh.AuthMethod
		if publicKey, err = getPublicKey(authValue); err != nil{
			log.Error("load public key err: %v", err)
		}
		config.Auth = []ssh.AuthMethod{publicKey}
	}

	addr := net.JoinHostPort(host, port)
	for i := 0; i < tcpDialRetry; i++ {
		client, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			log.Error("connect to %s failed, error message: %v, retrying...", addr, err)
			continue
		}else {
			return client, nil
		}
	}
	log.Error("connect to %s failed with %d time retry", addr, tcpDialRetry)
	return nil, err
}

func getPublicKey(file string) (ssh.AuthMethod, error){
	buf , err := ioutil.ReadFile(file)
	if err != nil{
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buf)
	if err != nil{
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

// SSHConnectionWithTTY supports to execute multi command with tty
type SSHConnectionWithTTY struct {
	client *ssh.Client
	session *ssh.Session
	stdin io.WriteCloser
	stdout io.Reader
}

func NewConnectionWithTTY(client *ssh.Client)(*SSHConnectionWithTTY, error){
	var sshC = &SSHConnectionWithTTY{}

	sshC.client = client
	if err := sshC.connect(); err != nil{
		return nil, err
	}
	return sshC, nil
}

func (c *SSHConnectionWithTTY) connect() error{
	c.session, err = c.client.NewSession()
	modes := ssh.TerminalModes{
		ssh.ECHO: 0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err = c.session.RequestPty("linux", 64, 80, modes); err != nil{
		log.Error("request pty error: %s", err.Error())
	}

	// StdinPipee() returns a pipe that will be connected to the remote command's standard input when the command starts.
	// StdoutPipe() returns a pipe that will be connected to the remote command's standard output when the command starts.
	if c.stdin, err = c.session.StdinPipe(); err != nil{
		return err
	}

	if c.stdout, err = c.session.StdoutPipe(); err != nil{
		return err
	}

	// start remote shell
	return c.session.Shell()
}

func (c *SSHConnectionWithTTY) RunCommand(cmd string)(string, error){
	if _, err = c.stdin.Write([]byte(cmd + "\r\n")); err != nil{
		return "", err
	}
	// wait for command to complete
	// we'll assume the moment we've gone 1 secs w/o any output that our command is done
	time.Sleep(1 * time.Second)

	strTmp := make([]byte, 8000)
	n, err := c.stdout.Read(strTmp)
	if n > 0 {
		return string(strTmp[:n]), nil
	}
	return string(strTmp), err
}

func (c *SSHConnectionWithTTY) Close(){
	_ = c.session.Close()
	_ = c.client.Close()
}

// SSHConnection supports to execute multi command
type SSHConnection struct {
	client *ssh.Client
	session *ssh.Session
	stdin io.WriteCloser
	stdout io.Reader
	stdinLines chan string
}

func NewSSHConnection(client *ssh.Client) (*SSHConnection, error){
	var sshC = &SSHConnection{}

	sshC.client = client
	if err := sshC.connect(); err != nil{
		return nil, err
	}
	return sshC, nil
}

func (c *SSHConnection) connect() error {
	var err error
	c.session, err = c.client.NewSession()
	if err != nil {
		return err
	}

	// StdinPipee() returns a pipe that will be connected to the remote command's standard input when the command starts.
	// StdoutPipe() returns a pipe that will be connected to the remote command's standard output when the command starts.
	c.stdin, err = c.session.StdinPipe()
	if err != nil {
		return err
	}

	c.stdout, err = c.session.StdoutPipe()
	if err != nil {
		return err
	}

	// Start remote shell
	return c.session.Shell()
}

func (c *SSHConnection) RunCommand(cmd string) (string, error){
	var strTmp string
	_, err := c.stdin.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", err
	}

	// wait for command to complete
	// we'll assume the moment we've gone 1 secs w/o any output that our command is done
InputLoop:
	for {
		timer := time.NewTimer(time.Second)
		select {
		case line, ok := <-c.stdinLines:
			if !ok {
				log.Error("Finished processing, command not executed:s", cmd)
				break InputLoop
			}
			strTmp += line
			strTmp += "\n"
		case <-timer.C:
			break InputLoop
		}
	}
	return strTmp, nil
}

func (c *SSHConnection) Close() {
	_ = c.stdin.Close()
	_ = c.session.Wait()
	_ = c.client.Close()
}