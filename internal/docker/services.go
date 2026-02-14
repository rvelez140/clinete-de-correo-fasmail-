package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

type PostgresOptions struct {
	User        string
	Password    string
	DBName      string
	NetworkName string
}

type RedisOptions struct {
	Password    string
	NetworkName string
}

const (
	PostgresContainerName = "fasmail-postgres"
	RedisContainerName    = "fasmail-redis"
	PostgresImage         = "postgres:16-alpine"
	RedisImage            = "redis:7-alpine"
	PostgresVolume        = "fasmail-pgdata"
	RedisVolume           = "fasmail-redis-data"
)

func (c *Client) CreatePostgres(ctx context.Context, opts PostgresOptions) (string, error) {
	if c.ContainerExists(ctx, PostgresContainerName) {
		c.RemoveContainer(ctx, PostgresContainerName)
	}

	if err := c.PullImage(ctx, PostgresImage); err != nil {
		return "", fmt.Errorf("pull postgres image: %w", err)
	}

	env := []string{
		fmt.Sprintf("POSTGRES_USER=%s", opts.User),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", opts.Password),
		fmt.Sprintf("POSTGRES_DB=%s", opts.DBName),
	}

	healthcheck := &container.HealthConfig{
		Test:        []string{"CMD-SHELL", fmt.Sprintf("pg_isready -U %s -d %s", opts.User, opts.DBName)},
		Interval:    5 * time.Second,
		Timeout:     5 * time.Second,
		Retries:     10,
		StartPeriod: 10 * time.Second,
	}

	containerCfg := &container.Config{
		Image:       PostgresImage,
		Env:         env,
		Healthcheck: healthcheck,
	}

	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: PostgresVolume,
				Target: "/var/lib/postgresql/data",
			},
		},
	}

	networkCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			opts.NetworkName: {},
		},
	}

	resp, err := c.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, PostgresContainerName)
	if err != nil {
		return "", fmt.Errorf("create postgres container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("start postgres container: %w", err)
	}

	if err := c.WaitForHealthy(ctx, resp.ID, 60*time.Second); err != nil {
		return "", fmt.Errorf("postgres health check: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) CreateRedis(ctx context.Context, opts RedisOptions) (string, error) {
	if c.ContainerExists(ctx, RedisContainerName) {
		c.RemoveContainer(ctx, RedisContainerName)
	}

	if err := c.PullImage(ctx, RedisImage); err != nil {
		return "", fmt.Errorf("pull redis image: %w", err)
	}

	cmd := []string{"redis-server", "--appendonly", "yes", "--maxmemory", "256mb", "--maxmemory-policy", "allkeys-lru"}
	if opts.Password != "" {
		cmd = append(cmd, "--requirepass", opts.Password)
	}

	healthcheck := &container.HealthConfig{
		Test:     []string{"CMD", "redis-cli", "ping"},
		Interval: 5 * time.Second,
		Timeout:  5 * time.Second,
		Retries:  10,
	}

	containerCfg := &container.Config{
		Image:       RedisImage,
		Cmd:         cmd,
		Healthcheck: healthcheck,
	}

	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: RedisVolume,
				Target: "/data",
			},
		},
	}

	networkCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			opts.NetworkName: {},
		},
	}

	resp, err := c.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, RedisContainerName)
	if err != nil {
		return "", fmt.Errorf("create redis container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("start redis container: %w", err)
	}

	if err := c.WaitForHealthy(ctx, resp.ID, 30*time.Second); err != nil {
		return "", fmt.Errorf("redis health check: %w", err)
	}

	return resp.ID, nil
}
