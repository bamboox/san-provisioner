package fc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
	log "k8s.io/klog"
)

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

var CmdRunnerCounter uint64 = 0

func cmdCounterGetter() uint64 {
	return atomic.AddUint64(&CmdRunnerCounter, 1)
}

func RunSSHCMD(cmd, host, user string, tags ...string) (string, string, error) {
	return innerRunSSHLocalCMD(cmd, host, user, tags...)
}

func RunSSHLocalCMD(cmd, host string, tags ...string) (string, string, error) {
	return innerRunSSHLocalCMD(cmd, host, "root", tags...)
}

// run shell command, error may castable to *ssh.ExitError
func innerRunSSHLocalCMD(cmd, host, user string, tags ...string) (string, string, error) {
	var (
		stdOut       bytes.Buffer
		stdErr       bytes.Buffer
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
		err          error
		errInfo      string
	)

	cnt := cmdCounterGetter()
	//tag := strings.Join(tags, "|")
	tag := fmt.Sprintf("%d|%s", cnt, strings.Join(tags, "|"))
	defer func() {
		if err != nil || errInfo != "" {
			log.Errorf("RunSSHLocalCMD-tag[%s] command[%s] failed; err: [%v], errInfo: [%s]", tag, cmd, err, errInfo)
		}
	}()

	hostKeyCallBack := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			publicKeyFile("/root/.ssh/id_rsa"),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: hostKeyCallBack,
	}

	addr = fmt.Sprintf("%s:%d", host, 22)
	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		log.Errorf("ssh.Dial-tag[%s] addr [%s][%s] failed; err: [%v]", tag, addr, cmd, err)
		return "", "", err
	}
	defer func() {
		cErr := client.Close()
		if cErr != nil {
			//do nothing
		}
	}()

	// create session
	if session, err = client.NewSession(); err != nil {
		log.Errorf("client.NewSession-tag[%s] addr [%s][%s] failed; err: [%v]", tag, addr, cmd, err)
		return "", "", err
	}
	defer func() {
		cErr := session.Close()
		if cErr != nil {
			//do nothing
		}
	}()

	log.Infof("RunSSHLocalCMD-tag[%s] begin command [%s] \n", tag, cmd)
	log.Infof("--%d--RunSSHLocalCMD -------------------------------------------wait....-------------------%d- |", cnt, cnt)
	defer log.Infof("^^%d^^RunSSHLocalCMD -------------------------------------------end....--------------------%d^^^", cnt, cnt)

	session.Stdout = &stdOut
	session.Stderr = &stdErr

	err = session.Run(cmd)
	if err != nil {
		//
	}

	errInfo = stdErr.String()
	outStr := stdOut.String()
	log.Infof("| RunSSHLocalCMD-tag[%s] end command [%s] , res :[%s]  \n\n", tag, cmd, outStr)
	if errInfo != "" {
		log.Errorf("RunSSHLocalCMD-tag[%s] end command [%s] , err:[%s] \n\n", tag, cmd, errInfo)
	}

	return outStr, errInfo, err
}
