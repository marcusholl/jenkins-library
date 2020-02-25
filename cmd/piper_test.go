package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

type execMockRunner struct {
	dir                 []string
	env                 [][]string
	calls               []execCall
	stdout              io.Writer
	stderr              io.Writer
	stdoutReturn        map[string]string
	shouldFailOnCommand map[string]error
}

type execCall struct {
	exec   string
	params []string
}

type shellMockRunner struct {
	dir                 string
	env                 [][]string
	calls               []string
	shell               []string
	stdout              io.Writer
	stderr              io.Writer
	stdoutReturn        map[string]string
	shouldFailOnCommand map[string]error
}

func (m *execMockRunner) Dir(d string) {
	m.dir = append(m.dir, d)
}

func (m *execMockRunner) Env(e []string) {
	m.env = append(m.env, e)
}

func (m *execMockRunner) RunExecutable(e string, p ...string) error {

	exec := execCall{exec: e, params: p}
	m.calls = append(m.calls, exec)

	c := strings.Join(append([]string{e}, p...), " ")

	return handleCall(c, m.stdoutReturn, m.shouldFailOnCommand, m.stdout)
}

func (m *execMockRunner) Stdout(out io.Writer) {
	m.stdout = out
}

func (m *execMockRunner) Stderr(err io.Writer) {
	m.stderr = err
}

func (m *shellMockRunner) Dir(d string) {
	m.dir = d
}

func (m *shellMockRunner) Env(e []string) {
	m.env = append(m.env, e)
}

func (m *shellMockRunner) RunShell(s string, c string) error {

	m.shell = append(m.shell, s)
	m.calls = append(m.calls, c)

	return handleCall(c, m.stdoutReturn, m.shouldFailOnCommand, m.stdout)
}

func handleCall(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {

	if stdoutReturn != nil {
		for k, v := range stdoutReturn {

			found := k == call

			if !found {

				r, e := regexp.Compile(k)
				if e != nil {
					return e
					// we don't distinguish here between an error returned
					// since it was configured or returning this error here
					// indicating an invalid regex. Anyway: when running the
					// test we will see it ...
				}
				if r.MatchString(call) {
					found = true

				}
			}

			if found {
				stdout.Write([]byte(v))
			}
		}
	}

	if shouldFailOnCommand != nil {
		for k, v := range shouldFailOnCommand {

			found := k == call

			if !found {
				r, e := regexp.Compile(k)
				if e != nil {
					return e
					// we don't distinguish here between an error returned
					// since it was configured or returning this error here
					// indicating an invalid regex. Anyway: when running the
					// test we will see it ...
				}
				if r.MatchString(call) {
					found = true

				}
			}

			if found {
				return v
			}
		}
	}

	return nil
}

func (m *shellMockRunner) Stdout(out io.Writer) {
	m.stdout = out
}

func (m *shellMockRunner) Stderr(err io.Writer) {
	m.stderr = err
}

type stepOptions struct {
	TestParam string `json:"testParam,omitempty"`
}

func openFileMock(name string) (io.ReadCloser, error) {
	var r string
	switch name {
	case "testDefaults.yml":
		r = "general:\n  testParam: testValue"
	case "testDefaultsInvalid.yml":
		r = "invalid yaml"
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
}

func TestAddRootFlags(t *testing.T) {
	var testRootCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
	addRootFlags(testRootCmd)

	assert.NotNil(t, testRootCmd.Flag("customConfig"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("defaultConfig"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("parametersJSON"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("stageName"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("stepConfigJSON"), "expected flag not available")
	assert.NotNil(t, testRootCmd.Flag("verbose"), "expected flag not available")

}

func TestPrepareConfig(t *testing.T) {
	defaultsBak := GeneralConfig.DefaultConfig
	GeneralConfig.DefaultConfig = []string{"testDefaults.yml"}
	defer func() { GeneralConfig.DefaultConfig = defaultsBak }()

	t.Run("using stepConfigJSON", func(t *testing.T) {
		stepConfigJSONBak := GeneralConfig.StepConfigJSON
		GeneralConfig.StepConfigJSON = `{"testParam": "testValueJSON"}`
		defer func() { GeneralConfig.StepConfigJSON = stepConfigJSONBak }()
		testOptions := stepOptions{}
		var testCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
		testCmd.Flags().StringVar(&testOptions.TestParam, "testParam", "", "test usage")
		metadata := config.StepData{
			Spec: config.StepSpec{
				Inputs: config.StepInputs{
					Parameters: []config.StepParameters{
						{Name: "testParam", Scope: []string{"GENERAL"}},
					},
				},
			},
		}

		PrepareConfig(testCmd, &metadata, "testStep", &testOptions, openFileMock)
		assert.Equal(t, "testValueJSON", testOptions.TestParam, "wrong value retrieved from config")
	})

	t.Run("using config files", func(t *testing.T) {
		t.Run("success case", func(t *testing.T) {
			testOptions := stepOptions{}
			var testCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
			testCmd.Flags().StringVar(&testOptions.TestParam, "testParam", "", "test usage")
			metadata := config.StepData{
				Spec: config.StepSpec{
					Inputs: config.StepInputs{
						Parameters: []config.StepParameters{
							{Name: "testParam", Scope: []string{"GENERAL"}},
						},
					},
				},
			}

			err := PrepareConfig(testCmd, &metadata, "testStep", &testOptions, openFileMock)
			assert.NoError(t, err, "no error expected but error occured")

			//assert config
			assert.Equal(t, "testValue", testOptions.TestParam, "wrong value retrieved from config")

			//assert that flag has been marked as changed
			testCmd.Flags().VisitAll(func(pflag *flag.Flag) {
				if pflag.Name == "testParam" {
					assert.True(t, pflag.Changed, "flag should be marked as changed")
				}
			})
		})

		t.Run("error case", func(t *testing.T) {
			GeneralConfig.DefaultConfig = []string{"testDefaultsInvalid.yml"}
			testOptions := stepOptions{}
			var testCmd = &cobra.Command{Use: "test", Short: "This is just a test"}
			metadata := config.StepData{}

			err := PrepareConfig(testCmd, &metadata, "testStep", &testOptions, openFileMock)
			assert.Error(t, err, "error expected but none occured")
		})
	})
}

func TestGetProjectConfigFile(t *testing.T) {

	tt := []struct {
		filename       string
		filesAvailable []string
		expected       string
	}{
		{filename: ".pipeline/config.yml", filesAvailable: []string{}, expected: ".pipeline/config.yml"},
		{filename: ".pipeline/config.yml", filesAvailable: []string{".pipeline/config.yml"}, expected: ".pipeline/config.yml"},
		{filename: ".pipeline/config.yml", filesAvailable: []string{".pipeline/config.yaml"}, expected: ".pipeline/config.yaml"},
		{filename: ".pipeline/config.yaml", filesAvailable: []string{".pipeline/config.yml", ".pipeline/config.yaml"}, expected: ".pipeline/config.yaml"},
		{filename: ".pipeline/config.yml", filesAvailable: []string{".pipeline/config.yml", ".pipeline/config.yaml"}, expected: ".pipeline/config.yml"},
	}

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			dir, err := ioutil.TempDir("", "")
			defer os.RemoveAll(dir) // clean up
			assert.NoError(t, err)

			if len(test.filesAvailable) > 0 {
				configFolder := filepath.Join(dir, filepath.Dir(test.filesAvailable[0]))
				err = os.MkdirAll(configFolder, 0700)
				assert.NoError(t, err)
			}

			for _, file := range test.filesAvailable {
				ioutil.WriteFile(filepath.Join(dir, file), []byte("general:"), 0700)
			}

			assert.Equal(t, filepath.Join(dir, test.expected), getProjectConfigFile(filepath.Join(dir, test.filename)))
		})
	}
}
