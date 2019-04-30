package main

import (
	"io"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/radu-matei/coras/pkg/coras"

	"github.com/spf13/cobra"
)

type pushCmd struct {
	inputBundle  string
	home         home.Home
	bundleIsFile bool
	targetRef    string
}

func newPushCmd(out io.Writer) *cobra.Command {
	const usage = `Pushes a CNAB bundle to an OCI repository.`
	const pushDesc = `
Pushes a CNAB bundle to an OCI registry by pushing all container
images referenced in the bundle to the target repository (all images are
pushed to the same repository, and are referenceable through their digest).

The first argument is the bundle to push (or the path to the bundle file, if the
--bundle-is-file flag is passed), and the second argument is the target repository
where the bundle and all referenced images will be pushed.

Insecure registries can be passed through the --insecure-registries flags,
and --allow-fallbacks enables automatic compatibility fallbacks for registries
without support for custom media type, or OCI manifests.

Examples:
$ duffle push bundle-reference registry/usernamne/bundle:tag
$ duffle push path-to-bundle.json --bundle-is-file registtry/username/bundle:tag
`
	var push pushCmd

	cmd := &cobra.Command{
		Use:   "push [BUNDLE] [TARGET REPOSITORY] [options]",
		Short: usage,
		Long:  pushDesc,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			push.home = home.Home(homePath())
			push.inputBundle = args[0]
			push.targetRef = args[1]
			return push.run()
		},
	}

	cmd.Flags().BoolVarP(&push.bundleIsFile, "bundle-is-file", "f", false, "Indicates that the bundle source is a file path")

	return cmd
}

func (p *pushCmd) run() error {
	bundleFile, err := resolveBundleFilePath(p.inputBundle, p.home.String(), p.bundleIsFile)
	if err != nil {
		return err
	}

	return coras.Push(bundleFile, p.targetRef, false)
}
