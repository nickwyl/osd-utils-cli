package cost

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/golang/mock/gomock"
	"github.com/openshift/osd-utils-cli/pkg/provider/aws/mock"
	//"gotest.tools/assert"
	"testing"
	gomega "github.com/onsi/gomega"
)

type mockSuite struct {
	mockCtrl      *gomock.Controller
	mockAwsClient *mock.MockClient
}

func setupAWSMocks(t *testing.T) *mockSuite {
	mocks := &mockSuite{
		mockCtrl: gomock.NewController(t),
	}

	mocks.mockAwsClient = mock.NewMockClient(mocks.mockCtrl)

	return mocks
}

func TestCreateCostCategory(t *testing.T) {
	//g := NewGomegaWithT(t)
	//testCases := []struct {
	//	title	     string
	//	OUid         *string
	//	OU           *organizations.OrganizationalUnit
	//	setupAwsClientMock func (r *mock.MockClientMockRecorder)
	//	errExpected  bool
	//}{
	//	{
	//		title: "Error calling CreateCostCategoryDefinition API",
	//		OUid: aws.String("ou-0wd6-oq5d7v8g"),
	//		OU: &organizations.OrganizationalUnit{Id: aws.String("ou-0wd6-oq5d7v8g")},
	//		setupAwsClientMock: func(r *mock.MockClientMockRecorder) {
	//			r.CreateCostCategoryDefinition(gomock.Any()).
	//				Return(nil, errors.New("FakeError")).Times(1)
	//			r.ListOrganizationalUnitsForParent(gomock.Any()).
	//				Return(nil, nil).Times(1)
	//		},
	//		errExpected: true,
	//	},
	//	{
	//		title: "Cost category already exists",
	//		OUid: aws.String("ou-0wd6-oq5d7v8g"),
	//		OU: &organizations.OrganizationalUnit{Id: aws.String("ou-0wd6-oq5d7v8g")},
	//		setupAwsClientMock: func(r *mock.MockClientMockRecorder) {
	//			r.CreateCostCategoryDefinition(gomock.Any()).
	//				Return(nil, awserr.New(
	//					"ValidationException",
	//					"Failed to create Cost Category: Cost category name already exists",
	//					nil,
	//					)).Times(1)
	//			r.ListOrganizationalUnitsForParent(gomock.Any()).
	//				Return(nil, nil).Times(1)
	//		},
	//		errExpected: true,
	//	},
	//}

	//for _, tc := range testCases {
	//	t.Run(tc.title, func(t *testing.T) {
	//		mocks := setupAWSMocks(t)
	//
	//		// This is necessary for the mocks to report failures like methods not being called an expected number of times.
	//		// after mocks is defined
	//		defer mocks.mockCtrl.Finish()
	//
	//		//OU := &organizations.OrganizationalUnit{Id: aws.String("ou-0wd6-oq5d7v8g")}
	//		//var OUSlice []*organizations.OrganizationalUnit
	//		//OUSlice = append(OUSlice, OU)
	//		//
	//		//OUs := &organizations.ListOrganizationalUnitsForParentOutput{
	//		//	OrganizationalUnits: OUSlice,
	//		//	NextToken: nil,
	//		//}
	//		//
	//		//
	//		//mocks.mockAwsClient.EXPECT().ListOrganizationalUnitsForParent(gomock.Any()).Return(OUs, nil).Times(1)
	//
	//		err := createCostCategory(tc.OUid, tc.OU, mocks.mockAwsClient)
	//
	//		if tc.errExpected {
	//			g.Expect(err).Should(HaveOccurred())
	//		} else {
	//			g.Expect(err).ShouldNot(HaveOccurred())
	//		}
	//	})
	//}


	controller := gomock.NewController(t)
	defer controller.Finish()

	awsReceiver := mock.NewMockClient(controller)

	input := &costexplorer.CreateCostCategoryDefinitionInput{
		Name:        aws.String("ou-0wd6-0huixd5v"),
		RuleVersion: aws.String("CostCategoryExpression.v1"),
		Rules: []*costexplorer.CostCategoryRule{
			{
				Rule: &costexplorer.Expression{
					Dimensions: &costexplorer.DimensionValues{
						Key:    aws.String("LINKED_ACCOUNT"),
						Values: []*string{aws.String("843659886818")},
					},
				},
				Value: aws.String("ou-0wd6-0huixd5v"),
			},
		},
	}

	awsReceiver.EXPECT().CreateCostCategoryDefinition(input).
		Return(nil, awserr.New(
			"ValidationException",
			"Failed to create Cost Category: Cost category name already exists",
			nil,
		)).Times(1)


	awsClient, _ := opsCost.initAWSClients()
	OU := getOU(awsClient, "ou-0wd6-0huixd5v")

	err := createCostCategory(aws.String("ou-0wd6-0huixd5v"), OU, awsClient)

	g := gomega.NewGomegaWithT(t)

	g.Expect(err)

}