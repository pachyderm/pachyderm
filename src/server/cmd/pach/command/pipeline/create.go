package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"

	"google.golang.org/grpc"

	"github.com/Jeffail/gabs"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"

	"github.com/urfave/cli"
)

func newCreateCommand() cli.Command {
	return cli.Command{
		Name:        "create",
		Aliases:     []string{"c"},
		Usage:       "Create a new pipeline.",
		ArgsUsage:   "create-pipeline -f pipeline.json",
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
	pipelineSpec := string(pachyderm.MustAsset("doc/deployment/pipeline_spec.md"))
	return fmt.Sprintf("Create a new pipeline from a spec\n\n%s", pipelineSpec)
}

func actCreate(c *cli.Context) error {
	cfgReader, err := newPipelineManifestReader(c.String("file"))
	if err != nil {
		return err
	}
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	for {
		request, err := cfgReader.nextCreatePipelineRequest()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if c.Bool("push-images") {
			pushedImage, err := pushImage(c.String("registry"), c.String("username"), c.String("password"), request.Transform.Image)
			if err != nil {
				return err
			}
			request.Transform.Image = pushedImage
		}
		if _, err := clnt.PpsAPIClient.CreatePipeline(context.Background(), request); err != nil {
			return sanitizeErr(err)
		}
	}
	return nil
}

// pipelineManifestReader helps with unmarshalling pipeline configs from JSON. It's used by
// create-pipeline and update-pipeline
type pipelineManifestReader struct {
	buf     bytes.Buffer
	decoder *json.Decoder
}

func newPipelineManifestReader(path string) (result *pipelineManifestReader, retErr error) {
	result = new(pipelineManifestReader)
	var pipelineReader io.Reader
	if path == "-" {
		pipelineReader = io.TeeReader(os.Stdin, &result.buf)
		fmt.Print("Reading from stdin.\n")
	} else if url, err := url.Parse(path); err == nil && url.Scheme != "" {
		resp, err := http.Get(url.String())
		if err != nil {
			return nil, sanitizeErr(err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil && retErr == nil {
				retErr = sanitizeErr(err)
			}
		}()
		rawBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		pipelineReader = io.TeeReader(strings.NewReader(string(rawBytes)), &result.buf)
	} else {
		rawBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		pipelineReader = io.TeeReader(strings.NewReader(string(rawBytes)), &result.buf)
	}
	result.decoder = json.NewDecoder(pipelineReader)
	return result, nil
}

func (r *pipelineManifestReader) nextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	var result ppsclient.CreatePipelineRequest
	s, err := replaceMethodAliases(r.decoder)
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, describeSyntaxError(err, r.buf)
	}
	if err := jsonpb.UnmarshalString(s, &result); err != nil {
		return nil, err
	}
	return &result, nil
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
