package cost

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/openshift/osd-utils-cli/pkg/provider/aws/mock"
	"testing"
)

type mockSuite struct {
	mockCtrl      *gomock.Controller
	mockOrgClient *mock.MockOrganizationsClient
	mockCEClient  *mock.MockCostExplorerClient
}

func setupAWSMocks(t *testing.T) *mockSuite {
	mocks := &mockSuite{
		mockCtrl: gomock.NewController(t),
	}

	mocks.mockOrgClient = mock.NewMockOrganizationsClient(mocks.mockCtrl)
	mocks.mockCEClient = mock.NewMockCostExplorerClient(mocks.mockCtrl)

	return mocks
}

func TestCreateCostCategory(t *testing.T) {
	g := NewGomegaWithT(t)
	testCases := []struct {
		title	     string
		OUid         *string
		OU           *organizations.OrganizationalUnit
		setupOrgMock func (r *mock.MockOrganizationsClientMockRecorder)
		setupCEMock  func(r *mock.MockCostExplorerClientMockRecorder)
		errExpected  bool
	}{
		{
			title: "Error calling CreateCostCategoryDefinition API",
			OUid: aws.String("ou-0wd6-oq5d7v8g"),
			OU: &organizations.OrganizationalUnit{Id: aws.String("ou-0wd6-oq5d7v8g")},
			setupOrgMock: func(d *mock.MockOrganizationsClientMockRecorder) {
				d.ListOrganizationalUnitsForParent(gomock.Any()).
					Return(nil).Times(1)
			},
			setupCEMock: func(r *mock.MockCostExplorerClientMockRecorder) {
				r.CreateCostCategoryDefinition(gomock.Any()).
					Return(nil, errors.New("FakeError")).Times(1)
			},
			errExpected: true,
		},
		{
			title: "Cost category already exists",
			OUid: aws.String("ou-0wd6-oq5d7v8g"),
			OU: &organizations.OrganizationalUnit{Id: aws.String("ou-0wd6-oq5d7v8g")},
			setupOrgMock: func(r *mock.MockOrganizationsClientMockRecorder) {
				r.ListOrganizationalUnitsForParent(gomock.Any()).
					Return(nil).Times(1)
			},
			setupCEMock: func(r *mock.MockCostExplorerClientMockRecorder) {
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
			mocks := setupAWSMocks(t)

			// This is necessary for the mocks to report failures like methods not being called an expected number of times.
			// after mocks is defined
			defer mocks.mockCtrl.Finish()

			//OU := &organizations.OrganizationalUnit{Id: aws.String("ou-0wd6-oq5d7v8g")}
			err := createCostCategory(tc.OUid, tc.OU, mocks.mockOrgClient, mocks.mockCEClient)

			if tc.errExpected {
				g.Expect(err).Should(HaveOccurred())
			} else {
				g.Expect(err).ShouldNot(HaveOccurred())
			}
		})
	}
}