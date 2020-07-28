package cost

import (
	"fmt"
	awsprovider "github.com/openshift/osd-utils-cli/pkg/provider/aws"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
func newCmdList(streams genericclioptions.IOStreams) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the cost of each OU under given OU",
		Run: func(cmd *cobra.Command, args []string) {

			awsClient, err := opsCost.initAWSClients()
			cmdutil.CheckErr(err)

			//Get flags
			OUid, err := cmd.Flags().GetString("ou")
			if err != nil {
				log.Fatalln("OU flag:", err)
			}
			time, err := cmd.Flags().GetString("time")
			if err != nil {
				log.Fatalln("Time flag:", err)
			}

			//Get Organizational Unit
			OU := organizations.OrganizationalUnit{Id: aws.String(OUid)}

			if err := listCostsUnderOU(&OU, awsClient, &time); err != nil {
				log.Fatalln("Error listing costs under OU:", err)
			}
		},
	}
	listCmd.Flags().String("ou", "ou-0wd6-aff5ji37", "get name of OU (default is name of v4's OU)")
	listCmd.Flags().StringP("time", "t", "", "set time")

	return listCmd
}

//List the cost of each OU under given OU
func listCostsUnderOU(OU *organizations.OrganizationalUnit, awsClient awsprovider.Client, timePtr *string) error {
	OUs, err := getOUsRecursive(OU, awsClient)
	if err != nil {
		return err
	}

	var cost float64

	//Print total cost for given OU
	if err := getOUCostRecursive(&cost, OU, awsClient, timePtr); err != nil {
		return nil
	}
	if len(OUs) != 0 {
		fmt.Printf("Cost of %s: %f\n\nCost of child OUs:\n", *OU.Id, cost)
	} else {
		fmt.Printf("Cost of %s: %f\nNo child OUs.\n", *OU.Id, cost)
	}
	//Print costs of child OUs under given OU
	for _, childOU := range OUs {
		cost = 0
		if err := getOUCostRecursive(&cost, childOU, awsClient, timePtr); err != nil {
			return nil
		}
		fmt.Printf("Cost of %s: %f\n", *childOU.Id, cost)
	}

	return nil
}
