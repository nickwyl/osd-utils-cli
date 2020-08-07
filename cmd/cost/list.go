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

			if err := listCostsUnderOU(OU, awsClient, ops); err != nil {
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
func listCostsUnderOU(OU *organizations.OrganizationalUnit, awsClient awsprovider.Client, ops *listOptions) error {
	OUs, err := getOUsRecursive(OU, awsClient)
	if err != nil {
		return err
	}

	var cost float64
	var unit string

	fmt.Println(ops.time)
	fmt.Println(&ops.time)

	if err := getOUCostRecursive(&cost, &unit, OU, awsClient, &ops.time); err != nil {
		return err
	}
	output := outputCost1

	output(cost, unit, OU, OUs, ops)
	//Print total cost for given OU
	//if csv {
	//	fmt.Printf("\nOU,Cost(%s)\n%v,%f\n", unit, *OU.Name, cost)
	//} else {
	//	if len(OUs) != 0 {
	//		fmt.Printf("\nCost of %s: %f %s\n\nCost of child OUs:\n", *OU.Name, cost, unit)
	//	} else {
	//		fmt.Printf("\nCost of %s: %f %s\nNo child OUs.\n", *OU.Name, cost, unit)
	//	}
	//}

	//Print costs of child OUs under given OU
	for _, childOU := range OUs {
		cost = 0
		if err := getOUCostRecursive(&cost, &unit, childOU, awsClient, &ops.time); err != nil {
			return err
		}
		output(cost, unit, OU, OUs, ops)

		//if csv {
		//	fmt.Printf("%v,%f\n", *childOU.Name, cost)
		//} else {
		//	fmt.Printf("Cost of %s: %f %s\n", *childOU.Name, cost, unit)
		//}
	}

	return nil
}

func outputCost1(cost float64, unit string, OU *organizations.OrganizationalUnit, OUs []*organizations.OrganizationalUnit, ops *listOptions) (y func()) {
	var isChildOU bool

	y = func() {
		isChildOU = true
	}

	if !isChildOU {
		if ops.csv {
			fmt.Printf("\nOU,Cost(%s)\n%v,%f\n", unit, *OU.Name, cost)
		} else {
			if len(OUs) != 0 {
				fmt.Printf("\nCost of %s: %f %s\n\nCost of child OUs:\n", *OU.Name, cost, unit)
			} else {
				fmt.Printf("\nCost of %s: %f %s\nNo child OUs.\n", *OU.Name, cost, unit)
			}
		}

		//y()
	} else {
		if ops.csv {
			fmt.Printf("%v,%f\n", *OU.Name, cost)
		} else {
			fmt.Printf("Cost of %s: %f %s\n", *OU.Name, cost, unit)
		}
	}
	return
}