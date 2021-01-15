package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
	transportrequest "github.com/SAP/jenkins-library/pkg/transportrequest/solman"
)

type transportRequestUploadSOLMANMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newTransportRequestUploadSOLMANTestsUtils() transportRequestUploadSOLMANMockUtils {
	utils := transportRequestUploadSOLMANMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunTransportRequestUploadSOLMAN(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := transportRequestUploadSOLMANOptions{}

		utils := newTransportRequestUploadSOLMANTestsUtils()

		// TODO needs to be replaced by mock ...
		action := transportrequest.SOLMANUploadAction{}
		// test
		err := runTransportRequestUploadSOLMAN(&config, &action, nil, utils)

		// assert
		assert.NoError(t, err)
	})
}
