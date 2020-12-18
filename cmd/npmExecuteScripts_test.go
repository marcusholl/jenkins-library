package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/stretchr/testify/assert"
	"testing"
)

// NpmMockUtilsBundle for mocking
type NpmMockUtilsBundle struct {
	*mock.FilesMock
	execRunner *mock.ExecMockRunner
}

// GetExecRunner return the execRunner mock
func (u *NpmMockUtilsBundle) GetExecRunner() npm.ExecRunner {
	return u.execRunner
}

// newNpmMockUtilsBundle creates an instance of NpmMockUtilsBundle
func newNpmMockUtilsBundle() NpmMockUtilsBundle {
	utils := NpmMockUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &mock.ExecMockRunner{}}
	return utils
}

func TestNpmExecuteScripts(t *testing.T) {
	t.Run("Call with packagesList", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorList: []string{"src/package.json"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"src/package.json"}, npmExecutor.Config.PackagesList)
			assert.Equal(t, []string{"ci-build", "ci-test"}, npmExecutor.Config.RunScripts)
			assert.Empty(t, npmExecutor.Config.ExcludeList)
			assert.False(t, npmExecutor.Config.VirtualFrameBuffer)
			assert.Empty(t, npmExecutor.Config.RunOptions)
			assert.Empty(t, npmExecutor.Config.ScriptOptions)
			assert.True(t, npmExecutor.Config.Install)

			// TODO check with collegues: I believe the check below is wrong. Should we have only the one file which is
			// contained in the build descriptor list above?
			assert.Equal(t, []string{"package.json", "src/package.json"}, npmExecutor.Config.FoundPackageFiles)
		}
	})

	t.Run("Call with excludeList", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorExcludeList: []string{"**/path/**"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			assert.Empty(t, npmExecutor.Config.PackagesList)
			assert.Equal(t, []string{"ci-build", "ci-test"}, npmExecutor.Config.RunScripts)
			assert.Equal(t, []string {"**/path/**"}, npmExecutor.Config.ExcludeList)
			assert.False(t, npmExecutor.Config.VirtualFrameBuffer)
			assert.Empty(t, npmExecutor.Config.RunOptions)
			assert.Empty(t, npmExecutor.Config.ScriptOptions)
			assert.True(t, npmExecutor.Config.Install)
			assert.Equal(t, []string{"package.json", "src/package.json"}, npmExecutor.Config.FoundPackageFiles)
		}
	})

	t.Run("Call with scriptOptions", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, ScriptOptions: []string{"--run"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			assert.Empty(t, npmExecutor.Config.PackagesList)
			assert.Equal(t, []string{"ci-build", "ci-test"}, npmExecutor.Config.RunScripts)
			assert.Empty(t, npmExecutor.Config.ExcludeList)
			assert.False(t, npmExecutor.Config.VirtualFrameBuffer)
			assert.Empty(t, npmExecutor.Config.RunOptions)
			assert.Equal(t, []string{"--run"}, npmExecutor.Config.ScriptOptions)
			assert.True(t, npmExecutor.Config.Install)
			assert.Equal(t, []string{"package.json", "src/package.json"}, npmExecutor.Config.FoundPackageFiles)
		}
	})

	t.Run("Call with install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			assert.Empty(t, npmExecutor.Config.PackagesList)
			assert.Equal(t, []string{"ci-build", "ci-test"}, npmExecutor.Config.RunScripts)
			assert.Empty(t, npmExecutor.Config.ExcludeList)
			assert.False(t, npmExecutor.Config.VirtualFrameBuffer)
			assert.Empty(t, npmExecutor.Config.RunOptions)
			assert.Empty(t, npmExecutor.Config.ScriptOptions)
			assert.True(t, npmExecutor.Config.Install)
			assert.Equal(t, []string{"package.json", "src/package.json"}, npmExecutor.Config.FoundPackageFiles)
		}
	})

	t.Run("Call without install", func(t *testing.T) {
		// TODO check with collegues: test name suggests this should run
		// with Install = false, but was true ... I set it to false ...
		config := npmExecuteScriptsOptions{/*Install: false,*/ RunScripts: []string{"ci-build", "ci-test"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			assert.Empty(t, npmExecutor.Config.PackagesList)
			assert.Equal(t, []string{"ci-build", "ci-test"}, npmExecutor.Config.RunScripts)
			assert.Empty(t, npmExecutor.Config.ExcludeList)
			assert.False(t, npmExecutor.Config.VirtualFrameBuffer)
			assert.Empty(t, npmExecutor.Config.RunOptions)
			assert.Empty(t, npmExecutor.Config.ScriptOptions)
			assert.False(t, npmExecutor.Config.Install)

			// TODO check with collegues: In this case no files are resolved. In the test before
			// we checked for the well known two files
			assert.Empty(t, npmExecutor.Config.FoundPackageFiles)
		}
	})

	t.Run("Call with virtualFrameBuffer", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, VirtualFrameBuffer: true}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			assert.Empty(t, npmExecutor.Config.PackagesList)
			assert.Equal(t, []string{"ci-build", "ci-test"}, npmExecutor.Config.RunScripts)
			assert.Empty(t, npmExecutor.Config.ExcludeList)
			assert.True(t, npmExecutor.Config.VirtualFrameBuffer)
			assert.Empty(t, npmExecutor.Config.RunOptions)
			assert.Empty(t, npmExecutor.Config.ScriptOptions)
			assert.True(t, npmExecutor.Config.Install)
			assert.Equal(t, []string{"package.json", "src/package.json"}, npmExecutor.Config.FoundPackageFiles)
		}
	})

	t.Run("Test integration with npm pkg", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build"}}

		options := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-build\": \"\" } }"))
		utils.AddFile("package-lock.json", []byte(""))

		npmExecutor := npm.Execute{Utils: &utils, Options: options}

		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry"}}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-build"}}, utils.execRunner.Calls[3])
			}
		}
	})
}
