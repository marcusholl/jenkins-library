package cmd

import (
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type versionOptions struct {
	Weekday string `json:"Weekday,omitempty"`
}

var myVersionOptions versionOptions
var versionStepConfigJSON string

// VersionCommand Returns the version of the piper binary
func VersionCommand() *cobra.Command {
	metadata := versionMetadata()
	var createVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Returns the version of the piper binary",
		Long:  `Writes the commit hash and the tag (if any) to stdout and exits with 0.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("version")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "version", &myVersionOptions, OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return version(myVersionOptions)
		},
	}

	addVersionFlags(createVersionCmd)
	return createVersionCmd
}

func addVersionFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myVersionOptions.Weekday, "Weekday", "SUNDAY", "The days of the week")

}

// retrieve step metadata
func versionMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "Weekday",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "enum",
						Mandatory: false,
					},
				},
			},
		},
	}
	return theMetaData
}
