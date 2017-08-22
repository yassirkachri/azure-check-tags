package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"
	"sync"
)

/******TODOs :
1- implement Action
2- implement Exclude
...
*/

// specify the parameters file name and path
var parametersFile = "./params.json"

// specift the outputs directory name and path
var outputsDir = "./outputs/"

// get Access Token to Azure using the REST API
//https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-v2-protocols-oauth-code
func getAccessToken(accName string, subscriptionID string, tenantID string, clientID string, clientSecret string, azureResource string, azureAuthentEndpoint string) (string, error) {
	sData := make(map[string]string)
	sData["grant_type"] = "client_credentials"
	sData["client_id"] = clientID
	sData["client_secret"] = clientSecret
	sData["resource"] = azureResource
	accessToken := ""
	lebody := sendRequest("POST", azureAuthentEndpoint, "/"+tenantID+"/oauth2/token/", sData, nil)
	if lebody != nil {
		//access token to be used by Azure REST API
		var larep map[string]string
		err := json.Unmarshal(lebody, &larep)
		onErrorFail(err, "json.Unmarshal(b, &larep) failed")
		accessToken = larep["token_type"] + " " + larep["access_token"]
		Trace.Println(" Token received ", accessToken)
		return accessToken, nil
	}
	return accessToken, errors.New("goinit : Problem with init function : GetToken ")

}

//get Resource Group List using https://docs.microsoft.com/en-us/rest/api/resources/ResourceGroups/List
// GET https://management.azure.com/subscriptions/{subscriptionId}/resourcegroups?api-version=2017-05-10&$filter&$top={$filter&$top}
func getListResourceGroups(sazureSubscription string, saccessToken string, managementHost string, apiversion string, nextLink string) (azResourceGroups, error) {
	var larep azResourceGroups
	var lebody []byte

	sHeader := make(map[string]string)
	sHeader["Authorization"] = saccessToken
	if nextLink == "" {
		sData := make(map[string]string)
		sData["api-version"] = apiversion
		lebody = sendRequest("GET", managementHost, "/subscriptions/"+sazureSubscription+"/resourcegroups", sData, sHeader)
	} else {
		lebody = sendGetRequest(nextLink, sHeader)
	}
	if lebody != nil {
		err := json.Unmarshal(lebody, &larep)
		onErrorFail(err, "json.Unmarshal(b, &larep) failed")
		return larep, nil
	}
	return larep, errors.New("getListResourceGroups : Problem getting ListResourceGroups")

}

// Get the resource List https://docs.microsoft.com/en-us/rest/api/resources/Resources/List
//GET https://management.azure.com/subscriptions/{subscriptionId}/resources?api-version=2017-05-10&$filter&$expand&$top={$filter&$expand&$top}
func getListResources(sazureSubscription string, saccessToken string, sfilter string, managementHost string, apiversion string, nextLink string) (azResources, error) {
	var larep azResources
	var lebody []byte
	sHeader := make(map[string]string)
	sHeader["Authorization"] = saccessToken
	sData := make(map[string]string)
	if nextLink == "" {
		sData["api-version"] = apiversion
		if sfilter != "" {
			sData["$filter"] = sfilter
		}
		sURL := "/subscriptions/" + sazureSubscription + "/resources"
		lebody = sendRequest("GET", managementHost, sURL, sData, sHeader)
	} else {
		//NextLink
		lebody = sendGetRequest(nextLink, sHeader)
	}
	if lebody != nil {
		err := json.Unmarshal(lebody, &larep)
		onErrorFail(err, "json.Unmarshal(b, &larep) failed")
		return larep, nil
	}
	return larep, errors.New("getListResources : Problem getting ListResources")
}

// Processing an Account
func processAccount(accountsParams actObject, loopIndex int) {

	fmt.Println("Processing  ", loopIndex+1, "/", len(accountsParams.Accounts), " : ", accountsParams.Accounts[loopIndex].Name, " on ", accountsParams.Accounts[loopIndex].Provider)
	Trace.Println("Processing  ", loopIndex+1, "/", len(accountsParams.Accounts), " : ", accountsParams.Accounts[loopIndex].Name, " on ", accountsParams.Accounts[loopIndex].Provider)
	if accountsParams.Accounts[loopIndex].Provider == "Azure" {

		appID := accountsParams.Accounts[loopIndex].Credentials.ApplicationID
		secret := accountsParams.Accounts[loopIndex].Credentials.KeySecret
		subscriptionID := accountsParams.Accounts[loopIndex].Credentials.SubscriptionID
		tenant := accountsParams.Accounts[loopIndex].Credentials.Tenant
		accname := accountsParams.Accounts[loopIndex].Name
		//get Azure Token for the current Account
		accessToken, err := getAccessToken(accname, subscriptionID, tenant, appID, secret, accountsParams.AzureResourcesEndpoint+"/", accountsParams.AzureAuthentEndpoint)
		onErrorFail(err, "actTnit failed")

		var linesOut []azOutputLine
		var lineOut azOutputLine

		//first line with titles
		lineOut.AccountName = "Account Name"
		lineOut.SubscriptionName = "Subscription ID"
		lineOut.TenantID = "Tenant ID"
		lineOut.ResourceID = "Resource ID"
		lineOut.ResourceGroupName = "Resource Group Name"
		lineOut.ResourceName = "Resource Name"
		lineOut.ResourceLocation = "Resource Location"
		lineOut.ResourceType = "Resource Type"
		lineOut.Tags = make(map[string]string)
		for k := 0; k < len(accountsParams.Accounts[loopIndex].Tags); k++ {
			lineOut.Tags[accountsParams.Accounts[loopIndex].Tags[k].Key] = accountsParams.Accounts[loopIndex].Tags[k].Key + "(" + accountsParams.Accounts[loopIndex].Tags[k].Value + ")"
		}
		linesOut = append(linesOut, lineOut)

		azLevel := accountsParams.Accounts[loopIndex].Level

		// check if the level is  VMs or ALL
		//  it's possible to use filter when calling Azure API but I was not be able to find the correct way to filte for the Classic and ARM Virtual Machines
		// I have tried several filters
		// sfilter = "resourceType eq 'Microsoft.Compute/virtualMachines'"  or  resourceType eq 'Microsoft.ClassicCompute/virtualMachines'"
		//
		if (azLevel == "VM") || (azLevel == "ALL") {
			var larep azResources
			sfilter := ""
			/*
				sfilter := ""
				if accountsParams.Accounts[loopIndex].Level == "VM" {
					sfilter = "resourceType eq 'Microsoft.Compute/virtualMachines'"  or  resourceType eq 'Microsoft.ClassicCompute/virtualMachines'"
				}
			*/
			larep, err = getListResources(subscriptionID, accessToken, sfilter, accountsParams.AzureResourcesEndpoint, accountsParams.AzureResourcesAPIVersion, "")
			onErrorFail(err, "getListResources  failed")

			// Loop while to manage pagination ( Azure have a  max limit  line : 1000 per response)
		while_res_1:
			for {
				for j := 0; j < len(larep.Valeur); j++ {
					//Trace.Println(accname, " : ", larep.Valeur[j].Name, " : ", larep.Valeur[j].ResType)
					if (azLevel == "ALL") || (azLevel == "VM" && ((larep.Valeur[j].ResType == "Microsoft.Compute/virtualMachines") || (larep.Valeur[j].ResType == "Microsoft.ClassicCompute/virtualMachines"))) {
						lineOut.AccountName = accname
						lineOut.SubscriptionName = subscriptionID
						lineOut.TenantID = tenant
						lineOut.ResourceID = larep.Valeur[j].ID
						lineOut.ResourceGroupName = getRessourceGroupFromID(larep.Valeur[j].ID)
						lineOut.ResourceType = larep.Valeur[j].ResType
						lineOut.ResourceName = larep.Valeur[j].Name
						lineOut.ResourceLocation = larep.Valeur[j].Location
						lineOut.Tags = make(map[string]string)
						for k := 0; k < len(accountsParams.Accounts[loopIndex].Tags); k++ {
							isoktag := isValideTags(accountsParams.Accounts[loopIndex].Tags[k].Key, accountsParams.Accounts[loopIndex].Tags[k].Value, larep.Valeur[j].Tags)
							lineOut.Tags[accountsParams.Accounts[loopIndex].Tags[k].Key] = isoktag
						}
						linesOut = append(linesOut, lineOut)
					}
				}
				// check if there is more results
				if len(larep.NextLink) > 0 {
					larep, err = getListResources(subscriptionID, accessToken, sfilter, accountsParams.AzureResourcesEndpoint, accountsParams.AzureResourcesAPIVersion, larep.NextLink)
					onErrorFail(err, "getListResources  failed")
				} else {
					break while_res_1
				}

			}

		}
		// if the requested level is Resource Group
		if azLevel == "RG" {
			var larep azResourceGroups
			larep, err = getListResourceGroups(subscriptionID, accessToken, accountsParams.AzureResourcesEndpoint, accountsParams.AzureResourcesAPIVersion, "")
			onErrorFail(err, "getListResourceGroups NextLink failed")
			// Loop while to manage pagination ( Azure have a  max limit  line : 1000 per response)
		while_res_2:
			for {
				for j := 0; j < len(larep.Valeur); j++ {
					lineOut.AccountName = accname
					lineOut.SubscriptionName = subscriptionID
					lineOut.TenantID = tenant
					lineOut.ResourceID = larep.Valeur[j].ID
					lineOut.ResourceGroupName = larep.Valeur[j].Name
					lineOut.ResourceType = ""
					lineOut.ResourceName = larep.Valeur[j].Name
					lineOut.ResourceLocation = larep.Valeur[j].Location
					lineOut.Tags = make(map[string]string)
					for k := 0; k < len(accountsParams.Accounts[loopIndex].Tags); k++ {
						isoktag := isValideTags(accountsParams.Accounts[loopIndex].Tags[k].Key, accountsParams.Accounts[loopIndex].Tags[k].Value, larep.Valeur[j].Tags)
						lineOut.Tags[accountsParams.Accounts[loopIndex].Tags[k].Key] = isoktag
					}
					linesOut = append(linesOut, lineOut)
				}
				// check if ther is more results
				if len(larep.NextLink) > 0 {
					larep, err = getListResourceGroups(subscriptionID, accessToken, accountsParams.AzureResourcesEndpoint, accountsParams.AzureResourcesAPIVersion, larep.NextLink)
					onErrorFail(err, "getListResourceGroups NextLink failed")
				} else {

					break while_res_2
				}

			}

		}

		fmt.Println(loopIndex+1, "/", len(accountsParams.Accounts), " - Writing   ", len(linesOut)-1, " lines to ", accname+"-"+accountsParams.OutputFilename, " : ", accountsParams.Accounts[loopIndex].Name, " on ", accountsParams.Accounts[loopIndex].Provider)
		Trace.Println(loopIndex+1, "/", len(accountsParams.Accounts), " - Writing   ", len(linesOut)-1, " lines to ", accname+"-"+accountsParams.OutputFilename, " : ", accountsParams.Accounts[loopIndex].Name, " on ", accountsParams.Accounts[loopIndex].Provider)
		writeOutputFileFromLines(linesOut, accname+"-"+accountsParams.OutputFilename, accountsParams.OutputSeparator)
		fmt.Println("End Processing  ", loopIndex+1, "/", len(accountsParams.Accounts), " : ", accountsParams.Accounts[loopIndex].Name, " on ", accountsParams.Accounts[loopIndex].Provider)
		Trace.Println("End Processing  ", loopIndex+1, "/", len(accountsParams.Accounts), " : ", accountsParams.Accounts[loopIndex].Name, " on ", accountsParams.Accounts[loopIndex].Provider)

	}

}

func main() {
	runtime.GOMAXPROCS(4)
	// init Loggers
	initLog()

	Trace.Println("Start......................................")
	fmt.Println("Start......................................")
	//Read parameters file
	paramsFileBody, err := ioutil.ReadFile(parametersFile)
	onErrorFail(err, "Reading Parameters file failed")
	// Parse the Json File into actObject
	var accountsParams actObject
	json.Unmarshal(paramsFileBody, &accountsParams)
	// use Go Routine and paralize processing accounts

	var wg sync.WaitGroup
	wg.Add(len(accountsParams.Accounts))
	for i := 0; i < len(accountsParams.Accounts); i++ {
		ndxLoop := i
		go func() {
			defer wg.Done()
			processAccount(accountsParams, ndxLoop)
		}()
	}
	wg.Wait()

	Trace.Println("End......................................")
	fmt.Println("End..........................................")

}
