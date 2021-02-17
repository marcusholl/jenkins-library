package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type transportRequestUploadCTSXXXMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTransportRequestUploadCTSXXXTestsUtils() transportRequestUploadCTSXXXMockUtils {
	utils := transportRequestUploadCTSXXXMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunTransportRequestUploadCTSXXX(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := transportRequestUploadCTSXXXOptions{}

		utils := newTransportRequestUploadCTSXXXTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runTransportRequestUploadCTSXXX(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := transportRequestUploadCTSXXXOptions{}

		utils := newTransportRequestUploadCTSXXXTestsUtils()

		// test
		err := runTransportRequestUploadCTSXXX(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
