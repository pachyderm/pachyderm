package job

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"

	"google.golang.org/grpc"

	"github.com/Jeffail/gabs"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/jsonpb"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	"github.com/pachyderm/pachyderm/src/server/pps/example"
	"github.com/urfave/cli"
)

func newCreateCommand() cli.Command {
	return cli.Command{
		Name:        "create",
		Aliases:     []string{"c"},
		Usage:       "Create a new job. Returns the id of the created job.",
		ArgsUsage:   "-f job.json",
		Description: descCreate(),
		Action:      actCreate,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file, f",
				Usage: "The file containing the job, it can be a url or local file. - reads from stdin.",
				Value: "-",
			},
			cli.BoolFlag{
				Name:  "push-images, p",
				Usage: "If true, push local docker images into the cluster registry.",
			},
			cli.StringFlag{
				Name:  "registry, r",
				Usage: "The registry to push images to.",
				Value: "docker.io",
			},
			cli.StringFlag{
				Name:  "username, u",
				Usage: "The username to push images as, defaults to your OS username.",
			},
			cli.StringFlag{
				Name:  "password, pw",
				Usage: "Your password for the registry being pushed to.",
			},
		},
	}
}

func descCreate() string {
	str, err := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(example.CreateJobRequest)
	if err != nil {
		pkgcmd.ErrorAndExit("error from CreateJob: %s", err.Error())
	}
	return fmt.Sprintf("Create a new job from a spec, the spec looks like this\n%s", str)
}

func actCreate(c *cli.Context) (retErr error) {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	var jobReader io.Reader
	jobPath := c.String("file")
	if jobPath == "-" {
		jobReader = io.TeeReader(os.Stdin, &buf)
		fmt.Print("Reading from stdin.\n")
	} else if url, err := url.Parse(jobPath); err == nil && url.Scheme != "" {
		resp, err := http.Get(url.String())
		if err != nil {
			return sanitizeErr(err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil && retErr == nil {
				retErr = sanitizeErr(err)
			}
		}()
		jobReader = resp.Body
	} else {
		jobFile, err := os.Open(jobPath)
		if err != nil {
			return sanitizeErr(err)
		}
		defer func() {
			if err := jobFile.Close(); err != nil && retErr == nil {
				retErr = sanitizeErr(err)
			}
		}()
		jobReader = io.TeeReader(jobFile, &buf)
	}
	var request ppsclient.CreateJobRequest
	decoder := json.NewDecoder(jobReader)
	s, err := replaceMethodAliases(decoder)
	if err != nil {
		return describeSyntaxError(err, buf)
	}
	if err := jsonpb.UnmarshalString(s, &request); err != nil {
		return sanitizeErr(err)
	}
	if c.Bool("push-images") {
		pushedImage, err := pushImage(c.String("registry"), c.String("username"), c.String("password"), request.Transform.Image)
		if err != nil {
			return err
		}
		request.Transform.Image = pushedImage
	}
	job, err := clnt.PpsAPIClient.CreateJob(context.Background(), &request)
	if err != nil {
		return sanitizeErr(err)
	}
	fmt.Println(job.ID)
	return nil
}

func replaceMethodAliases(decoder *json.Decoder) (string, error) {
	// We want to allow for a syntactic suger where the user
	// can specify a method with a string such as "map" or "reduce".
	// To that end, we check for the "method" field and replace
	// the string with an actual method object before we unmarshal
	// the json spec into a protobuf message
	pipeline, err := gabs.ParseJSONDecoder(decoder)
	if err != nil {
		return "", err
	}

	// No need to do anything if the pipeline does not specify inputs
	if !pipeline.ExistsP("inputs") {
		return pipeline.String(), nil
	}

	inputs := pipeline.S("inputs")
	children, err := inputs.Children()
	if err != nil {
		return "", err
	}
	for _, input := range children {
		if !input.ExistsP("method") {
			continue
		}
		methodAlias, ok := input.S("method").Data().(string)
		if ok {
			strat, ok := client.MethodAliasMap[methodAlias]
			if ok {
				input.Set(strat, "method")
			} else {
				return "", fmt.Errorf("unrecognized input alias: %s", methodAlias)
			}
		}
	}

	return pipeline.String(), nil
}

// FIXME: refactor
func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}

func describeSyntaxError(originalErr error, parsedBuffer bytes.Buffer) error {
	sErr, ok := originalErr.(*json.SyntaxError)
	if !ok {
		return originalErr
	}

	buffer := make([]byte, sErr.Offset)
	parsedBuffer.Read(buffer)

	lineOffset := strings.LastIndex(string(buffer[:len(buffer)-1]), "\n")
	if lineOffset == -1 {
		lineOffset = 0
	}

	lines := strings.Split(string(buffer[:len(buffer)-1]), "\n")
	lineNumber := len(lines)

	descriptiveErrorString := fmt.Sprintf("Syntax Error on line %v:\n%v\n%v^\n%v\n",
		lineNumber,
		string(buffer[lineOffset:]),
		strings.Repeat(" ", int(sErr.Offset)-2-lineOffset),
		originalErr,
	)

	return errors.New(descriptiveErrorString)
}

// pushImage pushes an image as registry/user/image. Registry and user can be left empty.
func pushImage(registry string, username string, password string, image string) (string, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return "", err
	}
	repo, _ := docker.ParseRepositoryTag(image)
	components := strings.Split(repo, "/")
	name := components[len(components)-1]
	if username == "" {
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		username = user.Username
	}
	pushRepo := fmt.Sprintf("%s/%s/%s", registry, username, name)
	pushTag := uuid.NewWithoutDashes()
	if err := client.TagImage(image, docker.TagImageOptions{
		Repo:    pushRepo,
		Tag:     pushTag,
		Context: context.Background(),
	}); err != nil {
		return "", err
	}
	var authConfig docker.AuthConfiguration
	if password != "" {
		authConfig = docker.AuthConfiguration{ServerAddress: registry}
		authConfig.Username = username
		authConfig.Password = password
	} else {
		authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
		if err != nil {
			return "", fmt.Errorf("error parsing auth: %s, try running `docker login`", err.Error())
		}
		for _, _authConfig := range authConfigs.Configs {
			serverAddress := _authConfig.ServerAddress
			if strings.Contains(serverAddress, registry) {
				authConfig = _authConfig
				break
			}
		}
	}
	fmt.Printf("Pushing %s:%s, this may take a while.\n", pushRepo, pushTag)
	if err := client.PushImage(
		docker.PushImageOptions{
			Name: pushRepo,
			Tag:  pushTag,
		},
		authConfig,
	); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", pushRepo, pushTag), nil
}
