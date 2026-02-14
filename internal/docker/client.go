package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
)

type Client struct {
	cli *dockerclient.Client
}

func NewClient() (*Client, error) {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

func (c *Client) IsAvailable(ctx context.Context) bool {
	if c.cli == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.cli.Ping(ctx)
	return err == nil
}

func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

func (c *Client) EnsureNetwork(ctx context.Context, networkName string) error {
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return fmt.Errorf("list networks: %w", err)
	}

	for _, n := range networks {
		if n.Name == networkName {
			return nil
		}
	}

	_, err = c.cli.NetworkCreate(ctx, networkName, network.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return fmt.Errorf("create network: %w", err)
	}

	return nil
}

func (c *Client) PullImage(ctx context.Context, imageName string) error {
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read all to wait for pull to complete
	buf := make([]byte, 4096)
	for {
		_, err := reader.Read(buf)
		if err != nil {
			break
		}
	}

	return nil
}

func (c *Client) WaitForHealthy(ctx context.Context, containerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		inspect, err := c.cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return fmt.Errorf("inspect container: %w", err)
		}

		if inspect.State.Health != nil {
			if inspect.State.Health.Status == "healthy" {
				return nil
			}
		} else if inspect.State.Running {
			// No healthcheck defined, just check if running
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("container %s did not become healthy within %v", containerID, timeout)
}

func (c *Client) GetContainerIP(ctx context.Context, containerID, networkName string) (string, error) {
	inspect, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("inspect container: %w", err)
	}

	if inspect.NetworkSettings != nil && inspect.NetworkSettings.Networks != nil {
		if net, ok := inspect.NetworkSettings.Networks[networkName]; ok {
			return net.IPAddress, nil
		}
	}

	if inspect.NetworkSettings != nil && inspect.NetworkSettings.IPAddress != "" {
		return inspect.NetworkSettings.IPAddress, nil
	}

	return "", fmt.Errorf("no IP found for container %s on network %s", containerID, networkName)
}

func (c *Client) RemoveContainer(ctx context.Context, name string) error {
	return c.cli.ContainerRemove(ctx, name, container.RemoveOptions{
		Force: true,
	})
}

func (c *Client) ContainerExists(ctx context.Context, name string) bool {
	_, err := c.cli.ContainerInspect(ctx, name)
	return err == nil
}
