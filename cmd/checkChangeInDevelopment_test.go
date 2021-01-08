package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	_ "github.com/stretchr/testify/assert"
	"testing"
)

type checkChangeInDevelopmentMockUtils struct {
	*mock.ExecMockRunner
}

func newCheckChangeInDevelopmentTestsUtils() checkChangeInDevelopmentMockUtils {
	utils := checkChangeInDevelopmentMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}

func TestRunCheckChangeInDevelopment(t *testing.T) {
	t.Parallel()
}
