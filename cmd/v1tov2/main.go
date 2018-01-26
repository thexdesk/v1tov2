package main

import (
	"context"
	"log"
	"net/http"

	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/schema1"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/palantir/stacktrace"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	contentStore, err := local.NewStore("/tmp/store")
	if err != nil {
		return stacktrace.Propagate(err, "failed to create local content store")
	}

	resolver := docker.NewResolver(docker.ResolverOptions{
		Credentials: nil,
		PlainHTTP:   true,
		Client:      &http.Client{},
	})

	ctx := context.Background()
	ref := "172.17.0.1:5000/repo/alpine"
	fetcher, err := resolver.Fetcher(ctx, ref)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create fetcher for '%s'", ref)
	}

	converter := schema1.NewConverter(contentStore, fetcher)

	dgst, err := digest.Parse("sha256:a2938f1d34192c41489111c22320fb7972e65862f1e841621f32dccb56802b57")
	if err != nil {
		return stacktrace.Propagate(err, "failed to parse digest")
	}

	desc := ocispec.Descriptor{
		Digest:    dgst,
		MediaType: images.MediaTypeDockerSchema1Manifest,
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

	return err
}

func credentials(host string) (username string, secret string, err error) {
	return "", "token", nil
}
