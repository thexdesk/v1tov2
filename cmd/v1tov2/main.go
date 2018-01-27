package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
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

	v1descs, err := converter.Handle(ctx, desc)
	if err != nil {
		return stacktrace.Propagate(err, "failed to pull manifest for digest '%s'", desc.Digest)
	}
	log.Printf("v1descs: %s\n", v1descs)

	v2desc, err := converter.Convert(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to convert v1 manifest for '%s' to v2 manifest", desc.Digest)
	}
	log.Printf("v2desc: %s\n", v2desc)

	pusher, err := resolver.Pusher(ctx, ref)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create pusher for '%s'", ref)
	}

	writer, err := pusher.Push(ctx, v2desc)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create writer for pushing blob '%s'", v2desc.Digest.String())
	}
	defer writer.Close()

	readerAt, err := contentStore.ReaderAt(ctx, v2desc.Digest)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create reader at")
	}
	defer readerAt.Close()

	sectionReader := io.NewSectionReader(readerAt, 0, v2desc.Size)
	err = content.Copy(ctx, writer, sectionReader, v2desc.Size, v2desc.Digest)
	if err != nil {
		return stacktrace.Propagate(err, "failed to copy blob from local content store to pusher")
	}

	return err
}

func credentials(host string) (username string, secret string, err error) {
	return "", "token", nil
}
