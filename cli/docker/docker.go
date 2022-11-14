package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os/user"
	"strings"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/moby/term"
)

// RunOptions options for docker container run.
type RunOptions struct {
	// BuildContext docker image build context directory.
	BuildCtxDir string
	// ImageTag - docker image tag.
	ImageTag string
	// Command is a command to run in container.
	Command []string
	// Binds - directory bindings in "host_dir:container_dir" format.
	Binds []string
	// Verbose, if set, verbose output is enabled.
	Verbose bool
}

// buildDockerImage builds docker image.
func buildDockerImage(dockerClient *client.Client, imageTag string, buildContextDir string,
	verbose bool, writer io.Writer) error {
	buildCtx, err := archive.TarWithOptions(buildContextDir, &archive.TarOptions{})
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageTag},
		Remove:     true,
	}

	ctx := context.Background()
	if buildResponse, err := dockerClient.ImageBuild(ctx, buildCtx, opts); err == nil {
		if buildResponse.Body != nil {
			defer buildResponse.Body.Close()
			if !verbose {
				writer = ioutil.Discard
			}
			termFd, isTerm := term.GetFdInfo(writer)
			if err = jsonmessage.DisplayJSONMessagesStream(buildResponse.Body,
				writer, termFd, isTerm, nil); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Docker image build failed: %s", err)
	}
	return nil
}

// RunContainer builds docker image and runs docker container.
func RunContainer(runOptions RunOptions, writer io.Writer) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	log.Infof("Start building docker image %s.", runOptions.ImageTag)
	if err = buildDockerImage(dockerClient, runOptions.ImageTag, runOptions.BuildCtxDir,
		runOptions.Verbose, writer); err != nil {
		return err
	}
	log.Info("Docker image is built.")

	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	log.Debug("Creating docker container.")
	ctx := context.Background()
	createResponse, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: runOptions.ImageTag,
		Cmd:   runOptions.Command,
		Tty:   false,
		User:  fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid),
	}, &container.HostConfig{Binds: runOptions.Binds}, nil, nil, "")
	if err != nil {
		return err
	}
	defer func() {
		log.Debugf("Removing container %s", createResponse.ID[:12])
		if err := dockerClient.ContainerRemove(ctx, createResponse.ID,
			types.ContainerRemoveOptions{}); err != nil {
			log.Warnf("Failed to remove container %s", createResponse.ID)
		}
	}()
	log.Debugf("Docker container %s is created.", createResponse.ID[:12])

	log.Debugf("The following command is going to be invoked in the container: %s.",
		strings.Join(runOptions.Command, " "))
	if err := dockerClient.ContainerStart(ctx, createResponse.ID,
		types.ContainerStartOptions{}); err != nil {
		return err
	}

	out, err := dockerClient.ContainerLogs(ctx, createResponse.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true})
	if err != nil {
		return err
	}
	stdcopy.StdCopy(writer, writer, out)
	out.Close()

	statusCh, errCh := dockerClient.ContainerWait(ctx, createResponse.ID,
		container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case st := <-statusCh:
		if st.StatusCode != 0 {
			return fmt.Errorf("Command returned %d.", st.StatusCode)
		}
	}
	return nil
}
