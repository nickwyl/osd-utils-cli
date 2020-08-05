package cost

import (
	"fmt"
	awsprovider "github.com/openshift/osd-utils-cli/pkg/provider/aws"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"log"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
func newCmdList(streams genericclioptions.IOStreams) *cobra.Command {
	ops := newListOptions(streams)
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the cost of each OU under given OU",
		Run: func(cmd *cobra.Command, args []string) {

			awsClient, err := opsCost.initAWSClients()
			cmdutil.CheckErr(err)

			OU := getOU(awsClient, ops.ou)

			if err := listCostsUnderOU(OU, awsClient, &ops.time); err != nil {
				log.Fatalln("Error listing costs under OU:", err)
			}
		},
	}
	listCmd.Flags().StringVar(&ops.ou, "ou", "ou-0wd6-aff5ji37", "get name of OU (default is name of v4's OU)")
	listCmd.Flags().StringVarP(&ops.time, "time", "t", "", "set time")
	listCmd.Flags().BoolVar(&ops.csv, "csv", false, "output result as csv")

	return listCmd
}

//Store flag options for get command
type listOptions struct {
	ou        string
	time      string
	csv       bool

	genericclioptions.IOStreams
}

func newListOptions(streams genericclioptions.IOStreams) *listOptions {
	return &listOptions{
		IOStreams: streams,
	}
}

//List the cost of each OU under given OU
func listCostsUnderOU(OU *organizations.OrganizationalUnit, awsClient awsprovider.Client, timePtr *string) error {
	OUs, err := getOUsRecursive(OU, awsClient)
	if err != nil {
		return err
	}

	var cost float64
	var unit string

	//Print total cost for given OU
	if err := getOUCostRecursive(&cost, &unit, OU, awsClient, timePtr); err != nil {
		return nil
	}
	if len(OUs) != 0 {
		fmt.Printf("Cost of %s: %f\n\nCost of child OUs:\n", *OU.Name, cost)
	} else {
		fmt.Printf("Cost of %s: %f\nNo child OUs.\n", *OU.Name, cost)
	}
	//Print costs of child OUs under given OU
	for _, childOU := range OUs {
		cost = 0
		if err := getOUCostRecursive(&cost, &unit, childOU, awsClient, timePtr); err != nil {
			return nil
		}
		fmt.Printf("Cost of %s: %f\n", *childOU.Id, cost)
	}

	return nil
}
