package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	logging "cloud.google.com/go/logging/apiv2"
	loggingpb "google.golang.org/genproto/googleapis/logging/v2"
)

func realMain(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("project_id is missing")
	}
	projectID := args[1]

	ctx := context.Background()
	c, err := logging.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	stream, err := c.TailLogEntries(ctx)
	if err != nil {
		return err
	}

	go func() {
		if err := stream.Send(&loggingpb.TailLogEntriesRequest{
			ResourceNames: []string{
				fmt.Sprintf("projects/%s", projectID),
			},
		}); err != nil {
			// TODO: proper error handling
			log.Println(err)
			return
		}
	}()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		for _, e := range resp.Entries {
			l, err := json.Marshal(e)
			if err != nil {
				return err
			}

			log.Println(string(l))
		}
	}

	return nil
}

func main() {
	if err := realMain(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
