package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/docker/libcompose/cli/command"
	"github.com/kballard/go-shellquote"
	"github.com/mozillazg/go-unidecode"
	rancherApp "github.com/ouzklcn/rancher-compose/app"
	"github.com/ouzklcn/rancher-compose/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"reflect"
)

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func afterApp(c *cli.Context) error {
	snsArn := c.GlobalString("callback-sns-arn")
	snsCallbackTemplate := c.GlobalString("callback-message")
	snsRegion := os.Getenv("AWS_REGION")
	if snsArn != "" && snsCallbackTemplate != "" && snsRegion != "" {
		sess, err := session.NewSession(&aws.Config{Region: aws.String(snsRegion)})
		if err != nil {
			return err
		}
		service := sns.New(sess)

		params := &sns.PublishInput{
			Message:  aws.String(snsCallbackTemplate),
			TopicArn: aws.String(snsArn),
		}

		if _, err := service.Publish(params); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(snsEvent events.SNSEvent) {
	logrus.Infof("%#v", snsEvent)
	factory := &rancherApp.ProjectFactory{}
	deleter := rancherApp.ProjectDeleter{}

	app := cli.NewApp()
	app.Name = "rancher-compose"
	app.Usage = "Docker-compose to Rancher"
	app.Version = version.VERSION
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Before = beforeApp
	app.After = afterApp
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "verbose,debug",
		},
		cli.StringSliceFlag{
			Name:   "file,f",
			Usage:  "Specify one or more alternate compose files (default: docker-compose.yml)",
			Value:  &cli.StringSlice{"/tmp/docker-compose.yml"},
			EnvVar: "COMPOSE_FILE",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "project-name,p",
			Usage:  "Specify an alternate project name (default: directory name)",
			EnvVar: "COMPOSE_PROJECT_NAME",
		},
		cli.StringFlag{
			Name: "url",
			Usage: "Specify the Rancher API endpoint URL",
			EnvVar: "RANCHER_URL",
		},
		cli.StringFlag{
			Name: "access-key",
			Usage: "Specify Rancher API access key",
			EnvVar: "RANCHER_ACCESS_KEY",
		},
		cli.StringFlag{
			Name: "secret-key",
			Usage: "Specify Rancher API secret key",
			EnvVar: "RANCHER_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "rancher-file,r",
			Usage:  "Specify an alternate Rancher compose file (default: rancher-compose.yml)",
			Value: "/tmp/rancher-compose.yml",
			Hidden: true,
		},
		cli.StringFlag{
			Name:  "env-file,e",
			Usage: "Specify a file from which to read environment variables",
		},
		cli.StringFlag{
			Name:  "bindings-file,b",
			Usage: "Specify a file from which to read bindings",
		},
		cli.StringFlag{
			Name: "github-access-token",
			Usage: "Specify Github Access token",
			EnvVar: "GITHUB_ACCESS_TOKEN",
		},
		cli.StringFlag{
			Name:  "github-repository-owner",
			Usage: "Specify Github repo owner",
			EnvVar: "GITHUB_REPOSITORY_OWNER",
		},
		cli.StringFlag{
			Name:  "github-repository-name",
			Usage: "Specify Github repo name",
			EnvVar: "GITHUB_REPOSITORY_NAME",
		},
		cli.StringFlag{
			Name:  "github-ref",
			Usage: "Specify Github ref",
			EnvVar: "GITHUB_REF",
		},
		cli.StringFlag{
			Name:  "github-rancher-file",
			Usage: "Specify Rancher compose file location in Github repo (default: rancher-compose.yml)",
			EnvVar: "GITHUB_RANCHER_FILE",
		},
		cli.StringFlag{
			Name:  "github-docker-file",
			Usage: "Specify Docker compose file location in Github repo (default: docker-compose.yml)",
			EnvVar: "GITHUB_DOCKER_FILE",
		},
		cli.StringFlag{
			Name:  "callback-sns-arn",
			Usage: "AWS SNS Arn for callback message)",
			EnvVar: "AWS_SNS_ARN",
		},
		cli.StringFlag{
			Name:  "callback-message",
			Usage: "Message to make a callback",
			EnvVar: "CALLBACK_MESSAGE",
		},
	}
	app.Commands = []cli.Command{
		rancherApp.CreateCommand(factory),
		rancherApp.UpCommand(factory),
		command.StartCommand(factory),
		command.LogsCommand(factory),
		rancherApp.RestartCommand(factory),
		rancherApp.StopCommand(factory),
		command.ScaleCommand(factory),
		command.RmCommand(factory),
		rancherApp.PullCommand(factory),
		rancherApp.UpgradeCommand(factory),
		rancherApp.DownCommand(factory, deleter),
	}

	message := unidecode.Unidecode(snsEvent.Records[0].SNS.Message)
	parsed, err := shellquote.Split(message)
	if err != nil {
		return
	}

	logrus.Infof("parsed: %s", parsed)
	for k, v := range snsEvent.Records[0].SNS.MessageAttributes  {
		if IsInstanceOf(v, map[string]interface{}(nil))  {
			if v.(map[string]interface{})["Type"] == "String" {
				os.Setenv(k , v.(map[string]interface{})["Value"].(string))
			}
		} else if IsInstanceOf(v, (string)("")) {
			os.Setenv(k, v.(string))
		}
	}

	if err := app.Run(parsed); err != nil {
		logrus.Error(err.Error())
	}
}

func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}
