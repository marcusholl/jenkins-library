package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
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
		config.Name = "SAP"

		utils := newHelloWorldTestsUtils()

		// test
		outOut := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		err := runHelloWorld(&config, nil, utils)

		w.Close()
		captured, _ := ioutil.ReadAll(r)
		os.Stdout = outOut

		assert.Equal(t, "Hello SAP!", strings.TrimSpace(string(captured)))
		// assert
		assert.NoError(t, err)
	})
}
