package main

import (
	"context"
	"log"
	"net/http"

	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/schema1"
	"github.com/palantir/stacktrace"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	resolver := docker.NewResolver(docker.ResolverOptions{
		Credentials: nil,
		PlainHTTP:   true,
		Client:      &http.Client{},
	})

	ctx := context.Background()
	ref := "172.17.0.1:5000/repo/alpine:v1"
	fetcher, err := resolver.Fetcher(ctx, ref)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create fetcher for '%s'", ref)
	}

	contentStore, err := local.NewStore("/tmp/store")
	if err != nil {
		return stacktrace.Propagate(err, "failed to create local content store")
	}

	converter := schema1.NewConverter(contentStore, fetcher)

	_, desc, err := resolver.Resolve(ctx, ref)
	if err != nil {
		return stacktrace.Propagate(err, "failed to resolve reference '%s'", ref)
	}

	err = images.Dispatch(ctx, converter, desc)
	if err != nil {
		return stacktrace.Propagate(err, "failed to pull manifest for digest '%s'", desc.Digest)
	}

	v2desc, err := converter.Convert(ctx, schema1.ConvertToDockerSchema2Manifest())
	if err != nil {
		return stacktrace.Propagate(err, "failed to convert v1 manifest for '%s' to v2 manifest", desc.Digest)
	}

	pusher, err := resolver.Pusher(ctx, ref)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create pusher for '%s'", ref)
	}

	err = remotes.PushContent(ctx, contentStore, pusher, nil, v2desc)
	if err != nil {
		return stacktrace.Propagate(err, "failed to push blobs")
	}

	return err
}
