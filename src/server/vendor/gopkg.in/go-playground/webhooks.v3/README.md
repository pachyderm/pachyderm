Library webhooks
================
<img align="right" src="https://raw.githubusercontent.com/go-playground/webhooks/v3/logo.png">![Project status](https://img.shields.io/badge/version-3.3.1-green.svg)
[![Build Status](https://travis-ci.org/go-playground/webhooks.svg?branch=v3)](https://travis-ci.org/go-playground/webhooks)
[![Coverage Status](https://coveralls.io/repos/go-playground/webhooks/badge.svg?branch=v3&service=github)](https://coveralls.io/github/go-playground/webhooks?branch=v3)
[![Go Report Card](https://goreportcard.com/badge/go-playground/webhooks)](https://goreportcard.com/report/go-playground/webhooks)
[![GoDoc](https://godoc.org/gopkg.in/go-playground/webhooks.v3?status.svg)](https://godoc.org/gopkg.in/go-playground/webhooks.v3)
![License](https://img.shields.io/dub/l/vibe-d.svg)

Library webhooks allows for easy receiving and parsing of GitHub, Bitbucket and GitLab Webhook Events

Features:

* Parses the entire payload, not just a few fields.
* Fields + Schema directly lines up with webhook posted json

Notes:

* Currently only accepting json payloads.

Installation
------------

Use go get.

```shell
	go get -u gopkg.in/go-playground/webhooks.v3
```

Then import the package into your own code.

	import "gopkg.in/go-playground/webhooks.v3"

Usage and Documentation
------

Please see http://godoc.org/gopkg.in/go-playground/webhooks.v3 for detailed usage docs.

##### Examples:

Multiple Handlers for each event you subscribe to
```go
package main

import (
	"fmt"
	"strconv"

	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

const (
	path = "/webhooks"
	port = 3016
)

func main() {

	hook := github.New(&github.Config{Secret: "MyGitHubSuperSecretSecrect...?"})
	hook.RegisterEvents(HandleRelease, github.ReleaseEvent)
	hook.RegisterEvents(HandlePullRequest, github.PullRequestEvent)

	err := webhooks.Run(hook, ":"+strconv.Itoa(port), path)
	if err != nil {
		fmt.Println(err)
	}
}

// HandleRelease handles GitHub release events
func HandleRelease(payload interface{}, header webhooks.Header) {

	fmt.Println("Handling Release")

	pl := payload.(github.ReleasePayload)

	// only want to compile on full releases
	if pl.Release.Draft || pl.Release.Prerelease || pl.Release.TargetCommitish != "master" {
		return
	}

	// Do whatever you want from here...
	fmt.Printf("%+v", pl)
}

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(payload interface{}, header webhooks.Header) {

	fmt.Println("Handling Pull Request")

	pl := payload.(github.PullRequestPayload)

	// Do whatever you want from here...
	fmt.Printf("%+v", pl)
}
```

Single receiver for events you subscribe to
```go
package main

import (
	"fmt"
	"strconv"

	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

const (
	path = "/webhooks"
	port = 3016
)

func main() {

	hook := github.New(&github.Config{Secret: "MyGitHubSuperSecretSecrect...?"})
	hook.RegisterEvents(HandleMultiple, github.ReleaseEvent, github.PullRequestEvent) // Add as many as you want

	err := webhooks.Run(hook, ":"+strconv.Itoa(port), path)
	if err != nil {
		fmt.Println(err)
	}
}

// HandleMultiple handles multiple GitHub events
func HandleMultiple(payload interface{}, header webhooks.Header) {

	fmt.Println("Handling Payload..")

	switch payload.(type) {

	case github.ReleasePayload:
		release := payload.(github.ReleasePayload)
		// Do whatever you want from here...
		fmt.Printf("%+v", release)

	case github.PullRequestPayload:
		pullRequest := payload.(github.PullRequestPayload)
		// Do whatever you want from here...
		fmt.Printf("%+v", pullRequest)
	}
}
```

Contributing
------

Pull requests for other services are welcome!

If the changes being proposed or requested are breaking changes, please create an issue for discussion.

License
------
Distributed under MIT License, please see license file in code for more details.
