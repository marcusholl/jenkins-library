package cmd

import (
	//"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/spf13/cobra"
)

type xsDeployOptions struct {
	Mode string `json:"mode,omitempty"`
}

var myXsDeployOptions xsDeployOptions
var xsDeployStepConfigJSON string

// XsDeployCommand Performs xs deployment
func XsDeployCommand() *cobra.Command {
	metadata := xsDeployMetadata()
	var createXsDeployCmd = &cobra.Command{
		Use:   "xsDeploy",
		Short: "Performs xs deployment",
		Long:  `Performs xs deployment`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return PrepareConfig(cmd, &metadata, "xsDeploy", &myXsDeployOptions, openPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return xsDeploy(myXsDeployOptions)
		},
	}

	addXsDeployFlags(createXsDeployCmd)
	return createXsDeployCmd
}

func addXsDeployFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myXsDeployOptions.Mode, "mode", "xxx", "The mode")

	cmd.MarkFlagRequired("mode")
}

// retrieve step metadata
func xsDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "mode",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
				},
			},
		},
	}
	return theMetaData
}
