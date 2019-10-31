package cmd

import (
	//"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/spf13/cobra"
)

type xsDeployOptions struct {
	Mode          string `json:"mode,omitempty"`
	ApiURL        string `json:"apiUrl,omitempty"`
	User          string `json:"user,omitempty"`
	Password      string `json:"password,omitempty"`
	Org           string `json:"org,omitempty"`
	Space         string `json:"space,omitempty"`
	LoginOpts     string `json:"loginOpts,omitempty"`
	XsSessionFile string `json:"xsSessionFile,omitempty"`
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
	cmd.Flags().StringVar(&myXsDeployOptions.ApiURL, "apiUrl", os.Getenv("PIPER_apiUrl"), "The api url (e.g. https://example.org:12345")
	cmd.Flags().StringVar(&myXsDeployOptions.User, "user", os.Getenv("PIPER_user"), "User")
	cmd.Flags().StringVar(&myXsDeployOptions.Password, "password", os.Getenv("PIPER_password"), "Password")
	cmd.Flags().StringVar(&myXsDeployOptions.Org, "org", os.Getenv("PIPER_org"), "The org")
	cmd.Flags().StringVar(&myXsDeployOptions.Space, "space", os.Getenv("PIPER_space"), "The space")
	cmd.Flags().StringVar(&myXsDeployOptions.LoginOpts, "loginOpts", os.Getenv("PIPER_loginOpts"), "Additional options for performing xs login.")
	cmd.Flags().StringVar(&myXsDeployOptions.XsSessionFile, "xsSessionFile", os.Getenv("PIPER_xsSessionFile"), "The file keeping the xs session.")

	cmd.MarkFlagRequired("mode")
	cmd.MarkFlagRequired("apiUrl")
	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("org")
	cmd.MarkFlagRequired("space")
	cmd.MarkFlagRequired("loginOpts")
	cmd.MarkFlagRequired("xsSessionFile")
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
					{
						Name:      "apiUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "user",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "password",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "org",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "space",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "loginOpts",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
					},
					{
						Name:      "xsSessionFile",
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
