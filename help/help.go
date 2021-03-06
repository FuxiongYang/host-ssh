package help

const Help = `			 host-ssh

NAME
	host-ssh is a smart remoteProtocol tool.It is developed by Go,compiled into a separate binary without any dependencies.

DESCRIPTION
		host-ssh can do the follow things:
		1.runs cmd on the remote host.
		2.push a local file or path to the remote host.
		3.pull remote host file to local.

USAGE
	1.Single Mode
		remote-comand:
		host-ssh -t cmd  -h host -P port(default 22) -u user(default root) -p passswrod [-f] command 

		Files-transfer:   
		<push file>   
		host-ssh -t push  -h host -P port(default 22) -u user(default root) -p passswrod [-f] localfile  remotepath 

		<pull file> 
		host-ssh -t pull -h host -P port(default 22) -u user(default root) -p passswrod [-f] remotefile localpath 

	2.Batch Mode
		Ssh-comand:
		host-ssh -t cmd -i ip_filename -P port(default 22) -u user(default root) -p passswrod [-f] command 

		Files-transfer:   
		host-ssh -t push -i ip_filename -P port(default 22) -u user(default root) -p passswrod [-f] localfile  remotepath 
		gosh -t pull -i ip_filename -P port(default 22) -u user(default root) -p passswrod [-f] remotefile localpath

EMAIL
    	email.tata@qq.com 
`
