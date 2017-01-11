package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	dock "github.com/docker/docker/client"
	"github.com/libgit2/git2go"
)

const (
	dockerImage   = "blueprint"
	dockerRepoDir = "/repo"
)

var (
	gitRootDir   string
	dockerClient *dock.Client
	httpClient   = &http.Client{
		Timeout: time.Minute,
	}
)

func init() {
	// TEMP
	os.Setenv("DOCKER_API_VERSION", "1.24")

	var err error

	gitRootDir, err = findGitRoot()
	if err != nil {
		er(err)
	}

	dockerClient, err = dock.NewEnvClient()
	if err != nil {
		er(fmt.Errorf("unable to connect to docker: %s", err))
	}

}

func findGitRoot() (string, error) {
	repoDir, err := git.Discover(".", true, nil)
	if err != nil {
		return "", fmt.Errorf("unable to find git repo: %s", err)
	}

	_, err = git.OpenRepository(repoDir)
	if err != nil {
		return "", fmt.Errorf("unable to open git repo: %s", err)
	}

	return filepath.Dir(strings.TrimSuffix(strings.TrimSuffix(repoDir, `/`), `\`)), nil
}
func er(msg interface{}) {
	fmt.Println("bp:", msg)
	os.Exit(-1)
}

func runInDocker(command string, cmp interface{}) error {
	jsn, err := json.Marshal(cmp)
	if err != nil {
		er(fmt.Errorf("unable to marshal api to json: %s", err))
	}

	ctx := context.Background()
	c, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:     dockerImage,
		OpenStdin: true,
		StdinOnce: true,
		Cmd:       strslice.StrSlice{"bp-in-docker", command},
	}, &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:"+dockerRepoDir, gitRootDir)},
	}, nil, "")
	if err != nil {
		return err
	}

	defer dockerClient.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})

	hij, err := dockerClient.ContainerAttach(ctx, c.ID, types.ContainerAttachOptions{
		Stdout: false,
		Stdin:  true,
		Stderr: false,
		Stream: true,
	})
	if err != nil {
		return fmt.Errorf("unable to attach to container: %s", err)
	}

	if err := dockerClient.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("unable to start docker container: %s", err)
	}

	if _, err := hij.Conn.Write(jsn); err != nil {
		return fmt.Errorf("unable to write to container: %s", err)
	}
	if err := hij.CloseWrite(); err != nil {
		return fmt.Errorf("unable to close attachment to container: %s", err)
	}

	_, err = dockerClient.ContainerWait(ctx, c.ID)
	if err != nil {
		return fmt.Errorf("unable to wait for docker container: %s", err)
	}

	logs, err := dockerClient.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return fmt.Errorf("unable to get container logs: %s", err)
	}

	if _, err := io.Copy(os.Stderr, logs); err != nil {
		return fmt.Errorf("unable to show docker logs: %s", err)
	}

	return nil
}
