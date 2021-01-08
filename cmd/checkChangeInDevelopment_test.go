package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type checkChangeInDevelopmentMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newCheckChangeInDevelopmentTestsUtils() checkChangeInDevelopmentMockUtils {
	utils := checkChangeInDevelopmentMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunCheckChangeInDevelopment(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := checkChangeInDevelopmentOptions{}

		utils := newCheckChangeInDevelopmentTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runCheckChangeInDevelopment(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := checkChangeInDevelopmentOptions{}

		utils := newCheckChangeInDevelopmentTestsUtils()

		// test
		err := runCheckChangeInDevelopment(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
