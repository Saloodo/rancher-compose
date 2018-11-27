package app

import (
	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"
	"github.com/docker/libcompose/cli/app"
	"github.com/docker/libcompose/cli/command"
	"github.com/docker/libcompose/cli/logger"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	rLookup "github.com/ouzklcn/rancher-compose/lookup"
	"github.com/ouzklcn/rancher-compose/rancher"
	"github.com/ouzklcn/rancher-compose/upgrade"
	"github.com/urfave/cli"
)

type ProjectFactory struct {
}

type ProjectDeleter struct {
}

func (p *ProjectFactory) Create(c *cli.Context) (project.APIProject, error) {
	githubClient, err := rancher.Create(c)
	if err != nil {
		return nil, err
	}

	err = githubClient.DownloadDockerComposeFile(c.GlobalStringSlice("file"), c.GlobalString("github-docker-file"))
	if err != nil {
		return nil, err
	}

	err = githubClient.DownloadRancherComposeFile(c.GlobalString("rancher-file"), c.GlobalString("github-rancher-file"))
	if err != nil {
		return nil, err
	}

	rancherComposeFile, err := rancher.ResolveRancherCompose(c.GlobalString("file"),
		c.GlobalString("rancher-file"))
	if err != nil {
		return nil, err
	}

	qLookup, err := rLookup.NewQuestionLookup(rancherComposeFile, &lookup.OsEnvLookup{})
	if err != nil {
		return nil, err
	}

	envLookup, err := rLookup.NewFileEnvLookup(c.GlobalString("env-file"), qLookup)
	if err != nil {
		return nil, err
	}

	ctx := &rancher.Context{
		Context: project.Context{
			ResourceLookup:    &rLookup.FileResourceLookup{},
			EnvironmentLookup: envLookup,
			LoggerFactory:     logger.NewColorLoggerFactory(),
		},
		RancherComposeFile: c.GlobalString("rancher-file"),
		Url:                c.GlobalString("url"),
		AccessKey:          c.GlobalString("access-key"),
		SecretKey:          c.GlobalString("secret-key"),
		PullCached:         c.Bool("cached"),
		Uploader:           &rancher.S3Uploader{},
		Args:               c.Args(),
		BindingsFile:       c.GlobalString("bindings-file"),
	}
	qLookup.Context = ctx

	command.Populate(&ctx.Context, c)

	ctx.Upgrade = c.Bool("upgrade") || c.Bool("force-upgrade")
	ctx.ForceUpgrade = c.Bool("force-upgrade")
	ctx.Rollback = c.Bool("rollback")
	ctx.BatchSize = int64(c.Int("batch-size"))
	ctx.Interval = int64(c.Int("interval"))
	ctx.ConfirmUpgrade = c.Bool("confirm-upgrade")
	ctx.Pull = c.Bool("pull")

	return rancher.NewProject(ctx)
}

func (p *ProjectDeleter) Delete(c *cli.Context) (error) {
	githubClient, err := rancher.Create(c)
	if err != nil {
		return err
	}

	err = githubClient.DownloadDockerComposeFile(c.GlobalStringSlice("file"), c.GlobalString("github-docker-file"))
	if err != nil {
		return err
	}

	err = githubClient.DownloadRancherComposeFile(c.GlobalString("rancher-file"), c.GlobalString("github-rancher-file"))
	if err != nil {
		return err
	}

	rancherComposeFile, err := rancher.ResolveRancherCompose(c.GlobalString("file"),
		c.GlobalString("rancher-file"))
	if err != nil {
		return err
	}

	qLookup, err := rLookup.NewQuestionLookup(rancherComposeFile, &lookup.OsEnvLookup{})
	if err != nil {
		return err
	}

	envLookup, err := rLookup.NewFileEnvLookup(c.GlobalString("env-file"), qLookup)
	if err != nil {
		return err
	}

	ctx := &rancher.Context{
		Context: project.Context{
			ResourceLookup:    &rLookup.FileResourceLookup{},
			EnvironmentLookup: envLookup,
			LoggerFactory:     logger.NewColorLoggerFactory(),
		},
		RancherComposeFile: c.GlobalString("rancher-file"),
		Url:                c.GlobalString("url"),
		AccessKey:          c.GlobalString("access-key"),
		SecretKey:          c.GlobalString("secret-key"),
		PullCached:         c.Bool("cached"),
		Uploader:           &rancher.S3Uploader{},
		Args:               c.Args(),
		BindingsFile:       c.GlobalString("bindings-file"),
	}
	qLookup.Context = ctx

	command.Populate(&ctx.Context, c)

	ctx.Upgrade = c.Bool("upgrade") || c.Bool("force-upgrade")
	ctx.ForceUpgrade = c.Bool("force-upgrade")
	ctx.Rollback = c.Bool("rollback")
	ctx.BatchSize = int64(c.Int("batch-size"))
	ctx.Interval = int64(c.Int("interval"))
	ctx.ConfirmUpgrade = c.Bool("confirm-upgrade")
	ctx.Pull = c.Bool("pull")

	return rancher.DeleteProject(ctx)
}


func RemoveStack(deleter ProjectDeleter) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		err := deleter.Delete(context)
		if err != nil {
			logrus.Fatalf("Failed to read project: %v", err)
		}
		return err
	}
}

func UpgradeCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "upgrade",
		Usage:  "Perform rolling upgrade between services",
		Action: app.WithProject(factory, Upgrade),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "batch-size",
				Usage: "Number of containers to upgrade at once",
				Value: 2,
			},
			cli.IntFlag{
				Name:  "scale",
				Usage: "Final number of running containers",
				Value: -1,
			},
			cli.IntFlag{
				Name:  "interval",
				Usage: "Update interval in milliseconds",
				Value: 2000,
			},
			cli.BoolTFlag{
				Name:  "update-links",
				Usage: "Update inbound links on target service",
			},
			cli.BoolFlag{
				Name:  "wait,w",
				Usage: "Wait for upgrade to complete",
			},
			cli.BoolFlag{
				Name:  "pull, p",
				Usage: "Before doing the upgrade do an image pull on all hosts that have the image already",
			},
			cli.BoolFlag{
				Name:  "cleanup, c",
				Usage: "Remove the original service definition once upgraded, implies --wait",
			},
		},
	}
}

func RestartCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "restart",
		Usage:  "Restart services",
		Action: app.WithProject(factory, app.ProjectRestart),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "batch-size",
				Usage: "Number of containers to restart at once",
				Value: 1,
			},
			cli.IntFlag{
				Name:  "interval",
				Usage: "Restart interval in milliseconds",
				Value: 0,
			},
		},
	}
}

func UpCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Bring all services up",
		Action: app.WithProject(factory, ProjectUp),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "pull, p",
				Usage: "Before doing the upgrade do an image pull on all hosts that have the image already",
			},
			cli.BoolFlag{
				Name:  "d",
				Usage: "Do not block and log",
			},
			cli.BoolFlag{
				Name:  "upgrade, u, recreate",
				Usage: "Upgrade if service has changed",
			},
			cli.BoolFlag{
				Name:  "force-upgrade, force-recreate",
				Usage: "Upgrade regardless if service has changed",
			},
			cli.BoolFlag{
				Name:  "confirm-upgrade, c",
				Usage: "Confirm that the upgrade was success and delete old containers",
			},
			cli.BoolFlag{
				Name:  "rollback, r",
				Usage: "Rollback to the previous deployed version",
			},
			cli.IntFlag{
				Name:  "batch-size",
				Usage: "Number of containers to upgrade at once",
				Value: 2,
			},
			cli.IntFlag{
				Name:  "interval",
				Usage: "Update interval in milliseconds",
				Value: 1000,
			},
		},
	}
}

func PullCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "pull",
		Usage:  "Pulls images for services",
		Action: app.WithProject(factory, app.ProjectPull),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "cached, c",
				Usage: "Only update hosts that have the image cached, don't pull new",
			},
		},
	}
}

func CreateCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "Create all services but do not start",
		Action: app.WithProject(factory, ProjectCreate),
	}
}

func ProjectCreate(p project.APIProject, c *cli.Context) error {
	if err := p.Create(context.Background(), options.Create{}, c.Args()...); err != nil {
		return err
	}

	// This is to fix circular links... What!? It works.
	if err := p.Create(context.Background(), options.Create{}, c.Args()...); err != nil {
		return err
	}

	return nil
}

func ProjectUp(p project.APIProject, c *cli.Context) error {
	if err := p.Create(context.Background(), options.Create{}, c.Args()...); err != nil {
		return err
	}

	if err := p.Up(context.Background(), options.Up{}, c.Args()...); err != nil {
		return err
	}

	if !c.Bool("d") {
		p.Log(context.Background(), true)
		// wait forever
		<-make(chan interface{})
	}

	return nil
}

func ProjectDown(p project.APIProject, c *cli.Context) error {
	err := p.Stop(context.Background(), c.Int("timeout"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	err = p.Delete(context.Background(), options.Delete{
		RemoveVolume: c.Bool("v"),
	}, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func Upgrade(p project.APIProject, c *cli.Context) error {
	args := c.Args()
	if len(args) != 2 {
		logrus.Fatalf("Please pass arguments in the form: [from service] [to service]")
	}

	err := upgrade.Upgrade(p, args[0], args[1], upgrade.UpgradeOpts{
		BatchSize:      c.Int("batch-size"),
		IntervalMillis: c.Int("interval"),
		FinalScale:     c.Int("scale"),
		UpdateLinks:    c.Bool("update-links"),
		Wait:           c.Bool("wait"),
		CleanUp:        c.Bool("cleanup"),
		Pull:           c.Bool("pull"),
	})

	if err != nil {
		logrus.Fatal(err)
	}
	return nil
}

// StopCommand defines the libcompose stop subcommand.
func StopCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:      "stop",
		Usage:     "Stop services",
		Action:    app.WithProject(factory, app.ProjectStop),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
				Value: 10,
			},
		},
	}
}

func DownCommand(factory app.ProjectFactory, deleter ProjectDeleter) cli.Command {
	return cli.Command{
		Name:   "down",
		Usage:  "Stop services",
		Action: app.WithProject(factory, ProjectDown),
		After:  RemoveStack(deleter),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force,f",
				Usage: "Allow deletion of all services",
			},
			cli.BoolFlag{
				Name:  "v",
				Usage: "Remove volumes associated with containers",
			},
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
				Value: 10,
			},
		},
	}
}