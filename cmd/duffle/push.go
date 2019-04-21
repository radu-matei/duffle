package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/reference"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cnab-to-oci/remotes"
	"github.com/spf13/cobra"
)

type pushCmd struct {
	inputBundle        string
	home               home.Home
	bundleIsFile       bool
	targetRef          string
	insecureRegistries []string
	allowFallbacks     bool
}

func newPushCmd(out io.Writer) *cobra.Command {
	const usage = `Pushes a CNAB bundle to a repository.`

	var push pushCmd

	cmd := &cobra.Command{
		Use:   "push <bundle file> [options]",
		Short: usage,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if push.targetRef == "" {
				return errors.New("--target flag must be set with a namespace ")
			}
			push.home = home.Home(homePath())
			push.inputBundle = args[0]
			return push.run()
		},
	}

	cmd.Flags().StringVarP(&push.targetRef, "target", "t", "", "reference where the bundle will be pushed")
	cmd.Flags().BoolVarP(&push.bundleIsFile, "bundle-is-file", "f", false, "Indicates that the bundle source is a file path")
	cmd.Flags().StringSliceVar(&push.insecureRegistries, "insecure-registries", nil, "Use plain HTTP for those registries")
	cmd.Flags().BoolVar(&push.allowFallbacks, "allow-fallbacks", true, "Enable automatic compatibility fallbacks for registries without support for custom media type, or OCI manifests")

	return cmd
}

func (p *pushCmd) run() error {

	bundleFile, err := resolveBundleFilePath(p.inputBundle, p.home.String(), p.bundleIsFile)
	if err != nil {
		return err
	}

	b, err := loadBundle(bundleFile)
	if err != nil {
		return err
	}

	resolverConfig := createResolver(p.insecureRegistries)
	ref, err := reference.ParseNormalizedNamed(p.targetRef)
	if err != nil {
		return err
	}

	err = remotes.FixupBundle(context.Background(), b, ref, resolverConfig, remotes.WithEventCallback(displayEvent))
	if err != nil {
		return err
	}
	d, err := remotes.Push(context.Background(), b, ref, resolverConfig.Resolver, p.allowFallbacks)
	if err != nil {
		return err
	}
	fmt.Printf("Pushed successfully, with digest %q\n", d.Digest)
	return nil
}

func createResolver(insecureRegistries []string) remotes.ResolverConfig {
	return remotes.NewResolverConfigFromDockerConfigFile(config.LoadDefaultConfigFile(os.Stderr), insecureRegistries...)
}

func displayEvent(ev remotes.FixupEvent) {
	switch ev.EventType {
	case remotes.FixupEventTypeCopyImageStart:
		fmt.Fprintf(os.Stderr, "Starting to copy image %s...\n", ev.SourceImage)
	case remotes.FixupEventTypeCopyImageEnd:
		if ev.Error != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy image %s: %s\n", ev.SourceImage, ev.Error)
		} else {
			fmt.Fprintf(os.Stderr, "Completed image %s copy\n", ev.SourceImage)
		}
	}
}
