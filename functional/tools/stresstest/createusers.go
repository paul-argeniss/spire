package main

import (
	"flag"
	"fmt"
	"os/exec"
)

type CreateUsers struct {
}

//Help prints the cmd usage
func (*CreateUsers) Help() string {
	return "Usage"
}

//Run create users
func (*CreateUsers) Run(args []string) int {
	var users int
	flags := flag.NewFlagSet("createusers", flag.ContinueOnError)
	flags.IntVar(&users, "workloads", 0, "Number of workloads to run in parallel")

	err := flags.Parse(args)
	if err != nil {
		panic(err)
	}
	if users == 0 {
		return 1
	}

	// Create users
	for i := 0; i < users; i++ {
		uid := 1000 + i

		fmt.Printf("Creating user %d\n", uid)

		// Create user
		o, err := exec.Command("bash", "-c", fmt.Sprintf("useradd --uid %d user%d", uid, uid)).CombinedOutput()
		if err != nil {
			fmt.Println(string(o))
			panic(err)
		}
	}

	return 0
}

//Synopsis of the command
func (*CreateUsers) Synopsis() string {
	return "Runs the server"
}
