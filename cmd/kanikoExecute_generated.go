// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type kanikoExecuteOptions struct {
	BuildOptions                []string `json:"buildOptions,omitempty"`
	ContainerBuildOptions       string   `json:"containerBuildOptions,omitempty"`
	ContainerImage              string   `json:"containerImage,omitempty"`
	ContainerPreparationCommand string   `json:"containerPreparationCommand,omitempty"`
	CustomTLSCertificateLinks   []string `json:"customTlsCertificateLinks,omitempty"`
	DockerConfigJSON            string   `json:"dockerConfigJSON,omitempty"`
	DockerfilePath              string   `json:"dockerfilePath,omitempty"`
}

// KanikoExecuteCommand Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.
func KanikoExecuteCommand() *cobra.Command {
	const STEP_NAME = "kanikoExecute"

	metadata := kanikoExecuteMetadata()
	var stepConfig kanikoExecuteOptions
	var startTime time.Time

	var createKanikoExecuteCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.",
		Long:  `Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.DockerConfigJSON)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			kanikoExecute(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addKanikoExecuteFlags(createKanikoExecuteCmd, &stepConfig)
	return createKanikoExecuteCmd
}

func addKanikoExecuteFlags(cmd *cobra.Command, stepConfig *kanikoExecuteOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.BuildOptions, "buildOptions", []string{`--skip-tls-verify-pull`}, "Defines a list of build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.")
	cmd.Flags().StringVar(&stepConfig.ContainerBuildOptions, "containerBuildOptions", os.Getenv("PIPER_containerBuildOptions"), "Deprected, please use buildOptions. Defines the build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.")
	cmd.Flags().StringVar(&stepConfig.ContainerImage, "containerImage", os.Getenv("PIPER_containerImage"), "Defines the full name of the Docker image to be created including registry, image name and tag like `my.docker.registry/path/myImageName:myTag`. If left empty, image will not be pushed.")
	cmd.Flags().StringVar(&stepConfig.ContainerPreparationCommand, "containerPreparationCommand", `rm -f /kaniko/.docker/config.json`, "Defines the command to prepare the Kaniko container. By default the contained credentials are removed in order to allow anonymous access to container registries.")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List containing download links of custom TLS certificates. This is required to ensure trusted connections to registries with custom certificates.")
	cmd.Flags().StringVar(&stepConfig.DockerConfigJSON, "dockerConfigJSON", os.Getenv("PIPER_dockerConfigJSON"), "Path to the file `.docker/config.json` - this is typically provided by your CI/CD system. You can find more details about the Docker credentials in the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/).")
	cmd.Flags().StringVar(&stepConfig.DockerfilePath, "dockerfilePath", `Dockerfile`, "Defines the location of the Dockerfile relative to the Jenkins workspace.")

}

// retrieve step metadata
func kanikoExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "kanikoExecute",
			Aliases: []config.Alias{},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "buildOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "containerBuildOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name: "containerImage",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/imageNameTag",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "containerImageNameAndTag"}},
					},
					{
						Name:        "containerPreparationCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name: "dockerConfigJSON",
						ResourceRef: []config.ResourceReference{
							{
								Name: "dockerConfigJsonCredentialsId",
								Type: "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:        "dockerfilePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerfile"}},
					},
				},
			},
		},
	}
	return theMetaData
}
