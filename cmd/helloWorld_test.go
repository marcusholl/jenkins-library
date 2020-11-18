package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type helloWorldMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newHelloWorldTestsUtils() helloWorldMockUtils {
	utils := helloWorldMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunHelloWorld(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		// init
		config := helloWorldOptions{}

		utils := newHelloWorldTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runHelloWorld(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		// init
		config := helloWorldOptions{}

		utils := newHelloWorldTestsUtils()

		// test
		err := runHelloWorld(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
