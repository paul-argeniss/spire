package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/spiffe/spire/proto/api/registration"
	"github.com/spiffe/spire/proto/common"
)

type Register struct {
}

//Help prints the cmd usage
func (*Register) Help() string {
	return "Usage"
}

//Run create users
func (*Register) Run(args []string) int {
	var users, ttl int
	var parent string
	flags := flag.NewFlagSet("register", flag.ContinueOnError)
	flags.IntVar(&users, "workloads", 5, "Number of workloads to run in parallel")
	flags.StringVar(&parent, "parent", "", "Parent Spiffe ID")
	flags.IntVar(&ttl, "ttl", 120, "SVID TTL")

	err := flags.Parse(args)
	if err != nil {
		panic(err)
	}
	if parent == "" {
		return 1
	}

	c, err := newRegistrationClient(serverAddr)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	// Register workloads
	for i := 0; i < users; i++ {
		uid := 1000 + i

		wg.Add(1)
		go func(uid int) {
			defer wg.Done()

			// Register workload
			selectorValue := fmt.Sprintf("uid:%d", uid)
			spiffeID := spiffeIDPrefix + fmt.Sprintf("uid%d", uid)
			fmt.Printf("Parent ID: %s\nSelector Value: %s\nSpiffe ID: %s\n", parent, selectorValue, spiffeID)
			entry := &common.RegistrationEntry{
				ParentId: parent,
				Selectors: []*common.Selector{
					&common.Selector{
						Type:  "unix",
						Value: selectorValue,
					},
				},
				SpiffeId: spiffeID,
				Ttl:      int32(ttl),
			}
			entryID, err := c.CreateEntry(context.TODO(), entry)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Created entry ID %s\n", entryID.Id)
		}(uid)
	}

	wg.Wait()

	return 0
}

//Synopsis of the command
func (*Register) Synopsis() string {
	return "Registers workloads"
}

func newRegistrationClient(address string) (registration.RegistrationClient, error) {
	// TODO: Pass a bundle in here
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	return registration.NewRegistrationClient(conn), err
}
