package ssh

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path"
	"regexp"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Queries the actual ssh-agent
func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

// LoadSSHUserName - thanks O.G. Sam.
func LoadSSHUserName() (string, error) {
	sshFilePath := path.Join(os.Getenv("HOME"), ".ssh", "config")
	sshConfig, err := ioutil.ReadFile(sshFilePath)
	if err != nil {
		return "", errors.New("could not load your SSH config.")
	}

	re, _ := regexp.Compile(`User (\w+)`)
	matches := re.FindStringSubmatch(string(sshConfig))
	if len(matches) > 1 {
		return matches[1], nil
	} else {
		return "", errors.New("could not find your SSH username. is it defined in ~/.ssh/config?")
	}
}
