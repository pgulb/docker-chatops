package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func ListContainers(ctx context.Context) (string, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	defer apiClient.Close()

	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	containers, err := apiClient.ContainerList(ctxTimeout, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	var resp []string
	resp = append(resp, "*Containers:*\n\n")
	for _, ctr := range containers {
		mountsRaw := ctr.Mounts
		mounts := []string{}
		for _, mount := range mountsRaw {
			mounts = append(mounts, fmt.Sprintf("%v:%v", mount.Source, mount.Destination))
		}
		mountStr := strings.Join(mounts, "\n")
		portsRaw := ctr.Ports
		ports := []string{}
		for _, port := range portsRaw {
			ports = append(ports, fmt.Sprintf("%v->%v", port.PrivatePort, port.PublicPort))
		}
		portsStr := strings.Join(ports, "\n")
		resp = append(
			resp,
			fmt.Sprintf("Name: %v\nImage: %v\ncommand: %v\nmounts: %v\nports: %v\nstatus: %v\n\n",
				strings.Join(ctr.Names, ", "),
				ctr.Image,
				ctr.Command,
				mountStr,
				portsStr,
				ctr.Status,
			),
		)
	}
	return strings.Join(resp, ""), nil
}

func ListContainersNamesOnly(ctx context.Context) ([]string, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer apiClient.Close()

	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	containers, err := apiClient.ContainerList(ctxTimeout, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var resp []string
	for _, ctr := range containers {
		resp = append(
			resp,
			ctr.Names...,
		)
	}
	return resp, nil
}

func TailLogs(ctx context.Context, containerName string) (string, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	defer apiClient.Close()

	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	logsRaw, err := apiClient.ContainerLogs(
		ctxTimeout,
		containerName,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "30",
		},
	)
	if err != nil {
		return "", err
	}
	defer logsRaw.Close()
	logsBytes, err := io.ReadAll(logsRaw)
	logs := string(logsBytes)
	if err != nil {
		return "", err
	}

	return logs, nil
}

func RestartContainer(ctx context.Context, containerName string) (string, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	defer apiClient.Close()

	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	err = apiClient.ContainerRestart(
		ctxTimeout,
		containerName,
		container.StopOptions{Timeout: nil},
	)
	if err != nil {
		return "", err
	}
	return "Container restarted.", nil
}
