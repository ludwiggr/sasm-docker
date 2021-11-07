package container

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/containerd/containerd/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pterodactyl/wings/system"
)

// ImagePullProgressDetails contains details of an ongoing pull
type ImagePullProgressDetails struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

// ImagePullStatus contains the status of an ongoing pull
type ImagePullStatus struct {
	ID             string                    `json:"id"`
	Status         string                    `json:"status"`
	Progress       string                    `json:"progress"`
	ProgressDetail *ImagePullProgressDetails `json:"progressDetail"`
}

// Event contains a container event
type Event struct {
	Type EventType
	Data interface{}
}

// LaunchableContainer is a container which can be launched by this launcher
type LaunchableContainer struct {
	onContainerEventFuncs []OnContainerEventFuc
	image                 string
	containerBody         *container.ContainerCreateCreatedBody
	containerName         string
	containerEnv          []string
	containerBinds        []string
	state                 State
	stream                *types.HijackedResponse
	// Tracks the environment state.
	st *system.AtomicString
}

// State is the state of a container
type State int

const (
	// OfflineState means the container is offline
	OfflineState State = iota
	// PullingState means the container image is being pulled
	PullingState
	// StartingState means the container is starting
	StartingState
	// RunningState means the container is running
	RunningState
	// StoppingState means the container is stopping
	StoppingState
)

// EventType is the type of a container event
type EventType int

const (
	// ImagePullStatusChanged means that the pull status or progress changed
	ImagePullStatusChanged EventType = iota
	// StateChanged means that the ContainerState changed
	StateChanged
	// ConsoleOutput means that there was some console output by the container
	ConsoleOutput
)

// OnContainerEventFuc is the type of a callback function which gets called on every container event
type OnContainerEventFuc func(event Event)

// A custom console writer that allows us to keep a function blocked until the
// given stream is properly closed. This does nothing special, only exists to
// make a noop io.Writer.
type noopWriter struct{}

var _ io.Writer = noopWriter{}

// Write implements the required Write function to satisfy the io.Writer interface.
func (nw noopWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

var ctx = context.Background()
var dockerClient *client.Client = nil

// New initializes the container module
func New() error {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.L.Error("Unable to connect to docker daemon")
		dockerClient = nil
		return err
	}
	return nil
}

// IsConnectedToDocker returns if we are connected to the docker daemon
func IsConnectedToDocker() bool {
	return dockerClient != nil
}

// OnContainerEvent allows to register a callback function
func (c *LaunchableContainer) OnContainerEvent(callback OnContainerEventFuc) {
	c.onContainerEventFuncs = append(c.onContainerEventFuncs, callback)
}

func newContainer(image string, containerID string, containerEnv []string, containerBinds []string) (*LaunchableContainer, error) {

	c := LaunchableContainer{
		image:          image,
		containerName:  containerID,
		containerEnv:   containerEnv,
		containerBinds: containerBinds,
	}

	return &c, nil
}

// Launch launches the container
func (c *LaunchableContainer) Launch() error {

	if err := c.Stop(); err != nil {
		return err
	}

	if err := c.Remove(); err != nil {
		return err
	}

	if err := c.Pull(); err != nil {
		return err
	}

	if err := c.Start(); err != nil {
		return err
	}

	if err := c.Attach(); err != nil {
		return err
	}

	return nil
}

// Stop stops the container
func (c *LaunchableContainer) Stop() error {
	if err := dockerClient.ContainerStop(ctx, c.containerName, nil); err != nil && !client.IsErrNotFound(err) {
		log.L.Error("Could not stop old container: " + err.Error())
		return err
	}
	c.setState(OfflineState)
	return nil
}

// Remove removes the conatiner
func (c *LaunchableContainer) Remove() error {
	if err := dockerClient.ContainerRemove(ctx, c.containerName, types.ContainerRemoveOptions{Force: true}); err != nil && !client.IsErrNotFound(err) {
		log.L.Error("Could not remove old container: " + err.Error())
		return err
	}
	c.setState(OfflineState)
	return nil
}

// Pull pulls the container
func (c *LaunchableContainer) Pull() error {
	c.setState(PullingState)
	defer c.setState(OfflineState)

	reader, err := dockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		s := ImagePullStatus{}

		//fmt.Println(scanner.Text())
		if err := json.Unmarshal(scanner.Bytes(), &s); err == nil {
			c.handleContainerEvent(Event{
				Type: ImagePullStatusChanged,
				Data: s,
			})
		}
	}
	return nil
}

// Start starts the container
func (c *LaunchableContainer) Start() error {
	c.setState(StartingState)
	cont, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image: image,
			Env:   c.containerEnv,
		},
		&container.HostConfig{
			Binds: c.containerBinds,
		}, nil, nil, c.containerName)

	if err != nil {
		c.setState(OfflineState)
		return err
	}

	c.containerBody = &cont

	err = dockerClient.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		c.setState(OfflineState)
		return err
	}

	c.setState(RunningState)
	return nil
}

// Attach attaches the log listener to the container
func (c *LaunchableContainer) Attach() error {

	if err := c.followOutput(); err != nil {
		return err
	}

	opts := types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}

	// Set the stream again with the container.
	st, err := dockerClient.ContainerAttach(ctx, c.containerBody.ID, opts)
	if err != nil {
		return err
	}
	c.stream = &st

	go func() {
		defer c.stream.Close()
		defer func() {
			c.setState(OfflineState)
			c.Remove()
			c.stream = nil
		}()

		// Block the completion of this routine until the container is no longer running. This allows
		// the pollResources function to run until it needs to be stopped. Because the container
		// can be polled for resource usage, even when stopped, we need to have this logic present
		// in order to cancel the context and therefore stop the routine that is spawned.
		//
		// For now, DO NOT use client#ContainerWait from the Docker package. There is a nasty
		// bug causing containers to hang on deletion and cause servers to lock up on the system.
		//
		// This weird code isn't intuitive, but it keeps the function from ending until the container
		// is stopped and therefore the stream reader ends up closed.
		// @see https://github.com/moby/moby/issues/41827
		w := new(noopWriter)
		if _, err := io.Copy(w, c.stream.Reader); err != nil {
			c.log("error", "could not copy from environment stream to noop writer: "+err.Error())
		}
	}()

	return nil
}

func (c *LaunchableContainer) followOutput() error {

	opts := types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Follow:     true,
		Since:      time.Now().Format(time.RFC3339),
	}

	reader, err := dockerClient.ContainerLogs(context.Background(), c.containerBody.ID, opts)
	if err != nil {
		return err
	}

	go c.scanOutput(reader)

	return nil
}

func (c *LaunchableContainer) scanOutput(reader io.ReadCloser) {
	defer reader.Close()

	if err := system.ScanReader(reader, func(line string) {
		c.handleContainerEvent(Event{
			Type: ConsoleOutput,
			Data: line,
		})
	}); err != nil && err != io.EOF {
		c.log("error", "error processing scanner line in console output: "+err.Error())
		return
	}

	if c.state == StoppingState || c.state == OfflineState {
		return
	}

	_ = reader.Close()

	go c.followOutput()
}

func (c *LaunchableContainer) handleContainerEvent(e Event) {
	for _, callbackFunc := range c.onContainerEventFuncs {
		callbackFunc(e)
	}
}

func (c *LaunchableContainer) setState(s State) {
	if s < OfflineState || s > StoppingState {
		panic(fmt.Errorf("invalid container state received: %d", s))
	}

	// Emit the event to any listeners that are currently registered.
	if c.state != s {
		// If the state changed make sure we update the internal tracking to note that.
		c.state = s
		c.handleContainerEvent(Event{
			Type: StateChanged,
			Data: s,
		})
	}
}

func (c *LaunchableContainer) log(l string, m string) {
	log.L.Info(m)
}
