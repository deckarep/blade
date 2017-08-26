package ssh

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/deckarep/blade/lib/recipe"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

var (
	concurrencySem        chan int
	hostQueue             = make(chan string)
	hostWg                sync.WaitGroup
	successfullyCompleted int32
	failedCompleted       int32
)

// StartSSHSession kicks off a session of work.
// TODO: This should take a recipe, and a struct of overrides.
func StartSSHSession(recipe *recipe.BladeRecipe, port int) {
	// Assumme root.
	if recipe.Overrides.User == "" {
		recipe.Overrides.User = "root"
	}

	sshConfig := &ssh.ClientConfig{
		User: recipe.Overrides.User,
		Auth: []ssh.AuthMethod{
			SSHAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshCmd := recipe.Required.Command

	// Concurrency must be at least 1 to make progress.
	if recipe.Overrides.Concurrency == 0 {
		recipe.Overrides.Concurrency = 1
	}

	concurrencySem = make(chan int, recipe.Overrides.Concurrency)
	go consumeAndLimitConcurrency(sshConfig, sshCmd)

	// If Hosts is defined, just use that as a discrete list.
	allHosts := recipe.Required.Hosts

	if len(allHosts) == 0 {
		// Otherwise do dynamic host lookup here.
		commandSlice := strings.Split(recipe.Required.HostLookupCommand, " ")
		out, err := exec.Command(commandSlice[0], commandSlice[1:]...).Output()
		if err != nil {
			fmt.Println("Couldn't execute command:", err.Error())
			return
		}

		allHosts = strings.Split(string(out), ",")
	}

	log.Print(color.GreenString(fmt.Sprintf("Starting recipe: %s", recipe.Meta.Name)))

	totalHosts := len(allHosts)
	for _, h := range allHosts {
		startHost(h, port)
	}

	hostWg.Wait()
	log.Print(color.GreenString(fmt.Sprintf("Completed recipe: %s - %d sucess | %d failed | %d total",
		recipe.Meta.Name,
		atomic.LoadInt32(&successfullyCompleted),
		atomic.LoadInt32(&failedCompleted),
		totalHosts)))
}

func startHost(host string, port int) {
	trimmedHost := strings.TrimSpace(host)

	// If it doesn't contain port :22 add it
	if !strings.Contains(trimmedHost, ":") {
		trimmedHost = fmt.Sprintf("%s:%d", trimmedHost, port)
	}

	// Ignore what you can't parse as host:port
	_, _, err := net.SplitHostPort(trimmedHost)
	if err != nil {
		log.Printf("Couldn't parse: %s", trimmedHost)
		return
	}

	//hostSet.Add(trimmedHost)

	// Finally queue it up for processing
	hostQueue <- trimmedHost
	hostWg.Add(1)
}

func consumeAndLimitConcurrency(sshConfig *ssh.ClientConfig, command string) {
	for host := range hostQueue {
		concurrencySem <- 1
		go func(h string) {
			defer func() {
				<-concurrencySem
				hostWg.Done()
			}()
			doSSH(sshConfig, h, command)
		}(host)
	}
}

func doSSH(sshConfig *ssh.ClientConfig, hostname string, command string) {
	var finalError error
	defer func() {
		if finalError != nil {
			log.Println(color.YellowString(hostname) + fmt.Sprintf(" error %s", finalError.Error()))
			//hostErrorSet.Add(hostname)
		}
	}()

	client, err := ssh.Dial("tcp", hostname, sshConfig)
	if err != nil {
		finalError = fmt.Errorf("Failed to dial remote host: %s", err.Error())
		return
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		finalError = fmt.Errorf("Failed to create session: %s", err.Error())
		return
	}
	defer session.Close()

	out, err := session.StdoutPipe()
	if err != nil {
		log.Fatal("Couldn't create pipe to session stdout... :(")
	}
	scanner := bufio.NewScanner(out)

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	if err := session.Run(command); err != nil {
		log.Printf("Failed to run command: %s", err.Error())
	}

	currentHost := strings.Split(hostname, ":")[0]
	logHost := color.GreenString(currentHost + ":")

	for scanner.Scan() {
		fmt.Println(logHost + " " + scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Print(color.RedString(currentHost) + ": Error reading output from this host.")
	}

	atomic.AddInt32(&successfullyCompleted, 1)
}