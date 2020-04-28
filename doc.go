/*
Package ezactions is the easy way to write a GitHub Action in go.

To get started with a hello world action:

- Create a git repo

- Copy the code below into `main.go`

- Create create a go module with `go mod init github.com/<you>/<reponame> && go mod tidy`

- Generate your Dockerfile and action.yml with `go generate .`

- Commit the new files and push to GitHub

- That's it. Now you have a hello world action.

	package main

	import (
		"fmt"
		"time"

		"github.com/willabides/ezactions"
	)

	//go:generate go run . -generate

	var action = &ezactions.Action{
		Name:        "Hello World",
		Description: "Greet someone and record the time",
		Inputs: []ezactions.ActionInput{
			{
				ID:          "who-to-greet",
				Description: "Who to greet",
				Required:    true,
				Default:     "World",
			},
		},
		Outputs: []ezactions.ActionOutput{
			{
				ID:          "time",
				Description: "The time we greeted you",
			},
		},
		Run: greet,
	}

	func greet(inputs map[string]string, resources *ezactions.RunResources) (outputs map[string]string, err error) {
		fmt.Println("Hello " + inputs["who-to-greet"])
		return map[string]string{
			"time": time.Now().Format(time.UnixDate),
		}, nil
	}

	func main() {
		action.Main()
	}

*/
package ezactions
