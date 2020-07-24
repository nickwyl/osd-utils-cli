package cost

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/openshift/osd-utils-cli/pkg/provider/aws/mock"
	"testing"
)

func TestCreateCostCategory(t *testing.T) {
	g := NewGomegaWithT(t)
	testCases := []struct {
		title	     string
		setupAWSMock func(r *mock.MockCostExplorerClientMockRecorder)
		errExpected  bool
	}{
		{
			title: "Error calling CreateCostCategoryDefinition API",
			setupAWSMock: func (r *mock.MockCostExplorerClientMockRecorder) {
				r.CreateCostCategoryDefinition(gomock.Any()).
					Return(nil, errors.New("FakeError")).Times(1)
			},
			errExpected: true,
		},
		{
			title: "Cost category already exists",
			setupAWSMock: func (r *mock.MockCostExplorerClientMockRecorder) {
				r.CreateCostCategoryDefinition(gomock.Any()).
					Return(nil, awserr.New(
						"ValidationException",
						"Failed to create Cost Category: Cost category name already exists",
						nil,
						)).Times(1)
			},
		},
	}
	_ = g
}