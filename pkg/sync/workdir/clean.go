package workdir

import (
	"context"

	"github.com/leonelquinteros/gotext"

	"github.com/FinzoYFunzo/sas/pkg/runtime"
	"github.com/FinzoYFunzo/sas/pkg/settings"
	"github.com/FinzoYFunzo/sas/pkg/settings/exe"
	"github.com/FinzoYFunzo/sas/pkg/settings/parser"
	"github.com/FinzoYFunzo/sas/pkg/text"
)

func removeMake(ctx context.Context, config *settings.Configuration,
	cmdBuilder exe.ICmdBuilder, makeDeps []string, cmdArgs *parser.Arguments,
) error {
	removeArguments := cmdArgs.CopyGlobal()

	err := removeArguments.AddArg("R", "s", "u")
	if err != nil {
		return err
	}

	for _, pkg := range makeDeps {
		removeArguments.AddTarget(pkg)
	}

	// removeMake always runs non-interactively; pass noConfirm=true directly
	// rather than mutating the global settings.NoConfirm.
	err = cmdBuilder.Show(cmdBuilder.BuildPacmanCmd(ctx,
		removeArguments, config.Mode, true))

	return err
}

func cleanAfter(ctx context.Context, run *runtime.Runtime,
	cmdBuilder exe.ICmdBuilder, pkgbuildDirs map[string]string,
) {
	run.Logger.Println(gotext.Get("removing untracked AUR files from cache..."))

	i := 0
	for _, dir := range pkgbuildDirs {
		run.Logger.OperationInfoln(gotext.Get("Cleaning (%d/%d): %s", i+1, len(pkgbuildDirs), text.Cyan(dir)))

		_, stderr, err := cmdBuilder.Capture(
			cmdBuilder.BuildGitCmd(
				ctx, dir, "reset", "--hard", "HEAD"))
		if err != nil {
			run.Logger.Errorln(gotext.Get("error resetting %s: %s", dir, stderr))
		}

		if err := run.CmdBuilder.Show(
			run.CmdBuilder.BuildGitCmd(
				ctx, dir, "clean", "-fx", "--exclude", "*.pkg.*")); err != nil {
			run.Logger.Errorln(err)
		}

		i++
	}
}
