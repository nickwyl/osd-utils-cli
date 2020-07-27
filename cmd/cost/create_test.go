package cost

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	awsosd "github.com/openshift/osd-utils-cli/pkg/provider/aws"
	"github.com/openshift/osd-utils-cli/pkg/provider/aws/mock"
	"testing"
)

type mockSuite struct {
	mockCtrl      *gomock.Controller
	mockAWSClient *mock.MockCostExplorerClient
}

//setupCostExplorerMocks is an easy way to setup all of the default mocks
func setupCostExplorerMocks(t *testing.T) *mockSuite {
	mocks := &mockSuite{
		mockCtrl: gomock.NewController(t),
	}

	mocks.mockAWSClient = mock.NewMockCostExplorerClient(mocks.mockCtrl)
	return mocks
}

func TestCreateCostCategory(t *testing.T) {
	g := NewGomegaWithT(t)
	testCases := []struct {
		title	     string
		OUid         string
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
			OUid: "ou-0wd6-oq5d7v8g",
			setupAWSMock: func (r *mock.MockCostExplorerClientMockRecorder) {
				r.CreateCostCategoryDefinition(gomock.Any()).
					Return(nil, awserr.New(
						"ValidationException",
						"Failed to create Cost Category: Cost category name already exists",
						nil,
						)).Times(1)
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mocks := setupCostExplorerMocks(t)
			tc.setupAWSMock(mocks.mockAWSClient.EXPECT())

			// This is necessary for the mocks to report failures like methods not being called an expected number of times.
			// after mocks is defined
			defer mocks.mockCtrl.Finish()

			//exists, err := costexplorer.CostExplorer.CreateCostCategoryDefinition()
			err := createCostCategory()

			if tc.errExpected {
				g.Expect(err).Should(HaveOccurred())
			} else {
				g.Expect(exists).Should(Equal(tc.exists))
			}
		})
	}
}