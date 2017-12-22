package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Run struct {
}

type workloadStats struct {
	uid     int
	runtime time.Duration
	output  string
	success bool
}

//Help prints the cmd usage
func (*Run) Help() string {
	return "Usage"
}

//Run tests
func (*Run) Run(args []string) int {
	var users, timeout, base int
	var wl string
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.IntVar(&users, "workloads", 5, "Number of workloads to run in parallel")
	//flags.StringVar(&token, "token", "", "Join token used in server and agent")
	flags.StringVar(&wl, "wl", "", "Path to workload executable")
	flags.IntVar(&base, "baseuid", 0, "Base UID")
	//flags.IntVar(&ttl, "ttl", 120, "SVID TTL")
	flags.IntVar(&timeout, "timeout", 15, "Total time to run test")

	err := flags.Parse(args)
	if err != nil || wl == "" {
		return 1
	}

	var wg sync.WaitGroup

	statch := make(chan *workloadStats, users)

	// Launch workloads
	for i := 0; i < users; i++ {
		uid := base + i

		fmt.Printf("Launching workload %d\n", uid)

		c := exec.Command(wl, "-timeout", strconv.Itoa(timeout))
		c.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{Uid: uint32(uid)},
		}
		wg.Add(1)
		go func(uid int) {
			started := time.Now()
			defer wg.Done()

			o, err := c.CombinedOutput()
			if err != nil {
				fmt.Printf("%d failed...\n", uid)
			}
			statch <- &workloadStats{
				uid:     uid,
				success: err == nil,
				output:  string(o),
				runtime: time.Now().Sub(started),
			}
		}(uid)
	}
	fmt.Printf("Waiting for workloads to finish... Test time is %d seconds\n", timeout)

	wg.Wait()

	fmt.Printf("Finished. Summary:\n")

	// Print stats
	statusMap := map[bool]string{true: "success", false: "failed"}
	for i := 0; i < users; i++ {
		s := <-statch
		logfile := fmt.Sprintf("%d.log", s.uid)
		fmt.Printf("Workload %d: status: %s, runtime: %s, logfile: %s\n",
			s.uid,
			statusMap[s.success],
			s.runtime.String(),
			logfile)

		f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Failed to open/create %s: %s\n", logfile, err)
		} else {
			defer f.Close()
			f.WriteString(s.output)
		}
	}

	return 0
}

//Synopsis of the command
func (*Run) Synopsis() string {
	return "Runs the server"
}
