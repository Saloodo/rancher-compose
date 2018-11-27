package rancher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
	"os"
)

type Github struct {
	Client          *github.Client
	ctx             context.Context
	RepositoryOwner string
	RepositoryName  string
	Ref             string
}

func Create(c *cli.Context) (Github, error)  {
	if  err := validateContext(c); err != nil {
		return Github{}, err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GlobalString("github-access-token")},
	)
	tc := oauth2.NewClient(ctx, ts)

	return Github{
		Client: github.NewClient(tc), ctx:ctx,
		RepositoryOwner: c.GlobalString("github-repository-owner"),
		RepositoryName: c.GlobalString("github-repository-name"),
		Ref: c.GlobalString("github-ref"),
	}, nil
}

func (g *Github) DownloadDockerComposeFile (localFilePath []string , githubFilePath string) (error) {
	var filePathSingle string
	if len(localFilePath) == 0 {
		filePathSingle = "/tmp/docker-compose.yml"
	} else {
		filePathSingle = localFilePath[0]
	}
	if githubFilePath == "" {
		githubFilePath = "docker-compose.yml"
	}
	return g.downloadFile(filePathSingle, githubFilePath)
}

func (g *Github) DownloadRancherComposeFile (localFilePath, githubFilePath string) (error) {
	if localFilePath == "" {
		localFilePath = "/tmp/rancher-compose.yml"
	}
	if githubFilePath == "" {
		githubFilePath = "rancher-compose.yml"
	}
	return g.downloadFile(localFilePath, githubFilePath)
}

func (g *Github) downloadFile (localFilePath, githubFilePath string) (error) {
	ref := g.Ref
	if ref == "" {
		ref = "master"
	}

	fc, _, _,err := g.Client.Repositories.GetContents(g.ctx, g.RepositoryOwner, g.RepositoryName, githubFilePath, &github.RepositoryContentGetOptions{Ref: ref})
	if err != nil {
		logrus.Errorf("Problem while downloading file\n %v\n", err)
		return errors.New(fmt.Sprintf("Problem while downloading file: %v", err))
	}

	fileContent, err := fc.GetContent()
	if err != nil {
		logrus.Errorf("Problem while reading file\n %v\n", err)
		return errors.New(fmt.Sprintf("Problem while creating file: %v", err))
	}

	f, err := os.Create(localFilePath)
	if err != nil {
		logrus.Errorf("Problem while creating file\n %v\n", err)
		return errors.New(fmt.Sprintf("Problem while creating file: %v", err))
	}

	w := bufio.NewWriter(f)
	w.WriteString(fileContent)
	w.Flush()

	return nil
}

func validateContext(c *cli.Context) error  {
	accessToken := c.GlobalString("github-access-token")
	if accessToken == "" {
		return fmt.Errorf("github access token is not set")
	}

	repositoryOwner := c.GlobalString("github-repository-owner")
	if repositoryOwner == "" {
		return fmt.Errorf("github repository owner is not set")
	}

	repositoryName := c.GlobalString("github-repository-name")
	if repositoryName == "" {
		return fmt.Errorf("github repository name is not set")
	}

	return nil
}