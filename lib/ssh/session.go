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
func StartSSHSession(recipe *recipe.BladeRecipe) {
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

	sshCmds := recipe.Required.Commands

	// Concurrency must be at least 1 to make progress.
	if recipe.Overrides.Concurrency == 0 {
		recipe.Overrides.Concurrency = 1
	}

	concurrencySem = make(chan int, recipe.Overrides.Concurrency)
	go consumeAndLimitConcurrency(sshConfig, sshCmds)

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
		startHost(h, recipe.Overrides.Port)
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

func consumeAndLimitConcurrency(sshConfig *ssh.ClientConfig, command []string) {
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

func doSSH(sshConfig *ssh.ClientConfig, hostname string, commands []string) {
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

	// Since we can run multiple commands, we need to keep track of intermediate failures
	// and log accordingly or do some type of aggregate report.
	for _, cmd := range commands {
		doSingleSSHCommand(client, hostname, cmd)
	}

	// Technically this is only successful when errors didn't occur above.
	atomic.AddInt32(&successfullyCompleted, 1)
}

// doSingleSSHCommand - the rule is one command can only ever occur per session.
func doSingleSSHCommand(client *ssh.Client, hostname, command string) {
	var finalError error
	defer func() {
		if finalError != nil {
			log.Println(color.YellowString(hostname) + fmt.Sprintf(" error %s", finalError.Error()))
			//hostErrorSet.Add(hostname)
		}
	}()

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

	// This is just a demo of how to not use buffering get immediate output.
	// This way is useful when you don't want to have to wait for buffering such as
	// echo 'sleeping for 5 seconds' && sleep 5
	// Note: just comment this out to go back to regular line scanning.
	//go io.Copy(os.Stdout, out)

	currentHost := strings.Split(hostname, ":")[0]
	logHost := color.GreenString(currentHost + ":")

	// Kick off the scanner async because we don't want to have to wait
	// for the the Run command to finish before we start dumping data to
	// the screen.
	// TODO: move to a thread-safe logger because this aint.
	go func() {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			fmt.Println(logHost + " " + scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			fmt.Print(color.RedString(currentHost) + ": Error reading output from this host.")
		}
	}()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	if err := session.Run(command); err != nil {
		log.Printf("Failed to run the command: %s", err.Error())
	}

}

// Notes: the above uses a buffer
// An alternative way to wire up the output is to just copy data from one end to another
// stdin, err := session.StdinPipe()
// if err != nil {
// 	return fmt.Errorf("Unable to setup stdin for session: %v", err)
// }
// go io.Copy(stdin, os.Stdin)

// stdout, err := session.StdoutPipe()
// if err != nil {
// 	return fmt.Errorf("Unable to setup stdout for session: %v", err)
// }
// go io.Copy(os.Stdout, stdout)

// stderr, err := session.StderrPipe()
// if err != nil {
// 	return fmt.Errorf("Unable to setup stderr for session: %v", err)
// }
// go io.Copy(os.Stderr, stderr)
