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

type ActionMock struct {
	Connection         transportrequest.SOLMANConnection
	ChangeDocumentId   string
	TransportRequestId string
	ApplicationID      string
	File               string
	CMOpts             []string
	performCalled      bool
}

func (a *ActionMock) WithConnection(c transportrequest.SOLMANConnection) {
	a.Connection = c
}
func (a *ActionMock) WithChangeDocumentId(id string) {
	a.ChangeDocumentId = id
}
func (a *ActionMock) WithTransportRequestId(id string) {
	a.TransportRequestId = id
}
func (a *ActionMock) WithApplicationID(id string) {
	a.ApplicationID = id
}
func (a *ActionMock) WithFile(f string) {
	a.File = f
}
func (a *ActionMock) WithCMOpts(opts []string) {
	a.CMOpts = opts
}
func (a *ActionMock) Perform(fs transportrequest.FileSystem, command transportrequest.Exec) error {
	a.performCalled = true
	return nil
}

func TestRunTransportRequestUploadSOLMAN(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := transportRequestUploadSOLMANOptions{
			Endpoint: "https://example.org/solman",
			Username: "me",
			Password: "********",
			ApplicationID: "XYZ",
			ChangeDocumentID: "12345678",
			TransportRequestID: "87654321",
			FilePath: "myApp.xxx",
			Cmclientops: []string{"-Dtest=abc123"},
		}
		utils := newTransportRequestUploadSOLMANTestsUtils()
		action := ActionMock{}

		err := runTransportRequestUploadSOLMAN(&config, &action, nil, utils)

		if assert.NoError(t, err) {
			assert.True(t, action.performCalled)
		}
	})
}
