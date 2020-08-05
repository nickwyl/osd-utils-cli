package cost

import (
	"fmt"
	awsprovider "github.com/openshift/osd-utils-cli/pkg/provider/aws"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/organizations"
)

// getCmd represents the get command
func newCmdGet(streams genericclioptions.IOStreams) *cobra.Command {
	ops := newGetOptions(streams)
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get total cost of a given OU. If no OU given, then gets total cost of v4 OU.",
		Run: func(cmd *cobra.Command, args []string) {

			awsClient, err := opsCost.initAWSClients()
			cmdutil.CheckErr(err)

			OU := getOU(awsClient, ops.ou)

			//Store cost
			var cost float64
			var unit string

			if ops.recursive { //Get cost of given OU by aggregating costs of all (including immediate) accounts under OU
				if err := getOUCostRecursive(&cost, &unit, OU, awsClient, &ops.time); err != nil {
					log.Fatalln("Error getting cost of OU recursively:", err)
				}

				if ops.csv { //If csv option specified, print result in csv
					fmt.Printf("\n%s,%f %s\n\n", *OU.Name, cost, unit)
				} else {
					fmt.Printf("\nCost of %s OU recursively is: %f %s\n\n", *OU.Name, cost, unit)
				}
			} else { //Get cost of given OU by aggregating costs of only immediate accounts under given OU
				if err := getOUCost(&cost, &unit, OU, awsClient, &ops.time); err != nil {
					log.Fatalln("Error getting cost of OU:", err)
				}

				if ops.csv {
					fmt.Printf("\n%s,%s%f\n\n", *OU.Name, unit, cost)
				} else {
					fmt.Printf("\nCost of %s OU is: %f%s\n\n", *OU.Name, cost, unit)
				}
			}
		},
	}
	getCmd.Flags().StringVar(&ops.ou, "ou", "ou-0wd6-aff5ji37", "get OU ID (default is v4)") //Default OU is v4
	getCmd.Flags().BoolVarP(&ops.recursive, "recursive", "r", false, "recurse through OUs")
	getCmd.Flags().StringVarP(&ops.time, "time", "t", "", "set time")
	getCmd.Flags().BoolVar(&ops.csv, "csv", false, "output result as csv")

	return getCmd
}

//Store flag options for get command
type getOptions struct {
	ou        string
	recursive bool
	time      string
	csv       bool

	genericclioptions.IOStreams
}

func newGetOptions(streams genericclioptions.IOStreams) *getOptions {
	return &getOptions{
		IOStreams: streams,
	}
}

//Get account IDs of immediate accounts under given OU
func getAccounts(OU *organizations.OrganizationalUnit, awsClient awsprovider.Client) ([]*string, error) {
	var accountSlice []*string
	var nextToken *string

	//Populate accountSlice with accounts by looping until accounts.NextToken is null
	for {
		accounts, err := awsClient.ListAccountsForParent(&organizations.ListAccountsForParentInput{
			ParentId:  OU.Id,
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(accounts.Accounts); i++ {
			accountSlice = append(accountSlice, accounts.Accounts[i].Id)
		}

		if accounts.NextToken == nil {
			break
		}
		nextToken = accounts.NextToken //If NextToken != nil, keep looping
	}

	return accountSlice, nil
}

//Get the account IDs of all (not only immediate) accounts under OU
func getAccountsRecursive(OU *organizations.OrganizationalUnit, awsClient awsprovider.Client) ([]*string, error) {
	var accountsIDs []*string

	//Populate OUs
	OUs, err := getOUs(OU, awsClient)
	if err != nil {
		return nil, err
	}

	//Loop through all child OUs to get account IDs from the accounts that comprise the OU
	for _, childOU := range OUs {
		accountsIDsOU, _ := getAccountsRecursive(childOU, awsClient)
		accountsIDs = append(accountsIDs, accountsIDsOU...)
	}
	//Get account
	accountsIDsOU, err := getAccounts(OU, awsClient)
	if err != nil {
		return nil, err
	}

	return append(accountsIDs, accountsIDsOU...), nil
}

//Get immediate OUs (child nodes) directly under given OU
func getOUs(OU *organizations.OrganizationalUnit, awsClient awsprovider.Client) ([]*organizations.OrganizationalUnit, error) {
	var OUSlice []*organizations.OrganizationalUnit
	var nextToken *string

	//Populate OUSlice with OUs by looping until OUs.NextToken is null
	for {
		OUs, err := awsClient.ListOrganizationalUnitsForParent(&organizations.ListOrganizationalUnitsForParentInput{
			ParentId:  OU.Id,
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		//Add OUs to slice
		for childOU := 0; childOU < len(OUs.OrganizationalUnits); childOU++ {
			OUSlice = append(OUSlice, OUs.OrganizationalUnits[childOU])
		}

		if OUs.NextToken == nil {
			break
		}
		nextToken = OUs.NextToken //If NextToken != nil, keep looping
	}

	return OUSlice, nil
}

//Get the account IDs of all (not only immediate) accounts under OU
func getOUsRecursive(OU *organizations.OrganizationalUnit, awsClient awsprovider.Client) ([]*organizations.OrganizationalUnit, error) {
	var OUs []*organizations.OrganizationalUnit

	//Populate OUs by getting immediate OUs (direct nodes)
	currentOUs, err := getOUs(OU, awsClient)
	if err != nil {
		return nil, err
	}

	//Loop through all child OUs. Append the child OU, then append the OUs of the child OU
	for _, currentOU := range currentOUs {
		OUs = append(OUs, currentOU)

		OUsRecursive, _ := getOUsRecursive(currentOU, awsClient)
		OUs = append(OUs, OUsRecursive...)
	}

	return OUs, nil
}

//Get cost of given account
func getAccountCost(accountID *string, unit *string, awsClient awsprovider.Client, timePtr *string, cost *float64) error {

	start, end := getTimePeriod(timePtr)
	granularity := "MONTHLY"
	metrics := []string{
		"NetUnblendedCost",
	}

	//Get cost information for chosen account
	costs, err := awsClient.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
		Filter: &costexplorer.Expression{
			Dimensions: &costexplorer.DimensionValues{
				Key: aws.String("LINKED_ACCOUNT"),
				Values: []*string{
					accountID,
				},
			},
		},
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: aws.String(granularity),
		Metrics:     aws.StringSlice(metrics),
	})
	if err != nil {
		return err
	}

	//Loop through month-by-month cost and increment to get total cost
	for month := 0; month < len(costs.ResultsByTime); month++ {
		monthCost, err := strconv.ParseFloat(*costs.ResultsByTime[month].Total["NetUnblendedCost"].Amount, 64)
		if err != nil {
			return err
		}
		*cost += monthCost
	}

	//Save unit
	*unit = *costs.ResultsByTime[0].Total["NetUnblendedCost"].Unit

	return nil
}

//Get cost of given OU by aggregating costs of only immediate accounts under given OU
func getOUCost(cost *float64, unit *string, OU *organizations.OrganizationalUnit, awsClient awsprovider.Client, timePtr *string) error {
	//Populate accounts
	accounts, err := getAccounts(OU, awsClient)
	if err != nil {
		return err
	}

	//Increment costs of accounts
	for _, account := range accounts {
		if err := getAccountCost(account, unit, awsClient, timePtr, cost); err != nil {
			return err
		}
	}

	return nil
}

//Get cost of given OU by aggregating costs of all (including immediate) accounts under OU
func getOUCostRecursive(cost *float64, unit *string, OU *organizations.OrganizationalUnit, awsClient awsprovider.Client, timePtr *string) error {
	//Populate OUs
	OUs, err := getOUs(OU, awsClient)
	if err != nil {
		return err
	}

	//Loop through all child OUs, get their costs, and store it to cost of current OU
	for _, childOU := range OUs {
		if err := getOUCostRecursive(cost, unit, childOU, awsClient, timePtr); err != nil {
			return err
		}
	}

	//Return cost of child OUs + cost of immediate accounts under current OU
	if err := getOUCost(cost, unit, OU, awsClient, timePtr); err != nil {
		return err
	}

	return nil
}

func getTimePeriod(timePtr *string) (string, string) {
	t := time.Now()

	//Starting from the 1st of the current month last year i.e. if today is 2020-06-29, then start date is 2019-06-01
	start := fmt.Sprintf("%d-%02d-%02d", t.Year()-1, t.Month(), 01)
	end := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())

	switch *timePtr {
	case "LM": //Last Month
		start = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month()-1, 01)
		end = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), 01)
	case "MTD":
		start = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), 01)
	case "YTD":
		start = fmt.Sprintf("%d-%02d-%02d", t.Year(), 01, 01)
	case "3M":
		if month := t.Month(); month > 3 {
			start = t.AddDate(0, -3, 0).Format("2006-01-02")
		} else {
			start = t.AddDate(-1, 9, 0).Format("2006-01-02")
		}
	case "6M":
		if month, _ := strconv.Atoi(time.Now().Format("01")); month > 6 {
			start = t.AddDate(0, -6, 0).Format("2006-01-02")
		} else {
			start = t.AddDate(-1, 6, 0).Format("2006-01-02")
		}
	case "1Y":
		start = t.AddDate(-1, 0, 0).Format("2006-01-02")
	}

	return start, end
}