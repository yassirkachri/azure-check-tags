# azure-check-tags
Check if Azure resource  are well tagged within multi-subscriptions. The conformity of tags is based on regular expressions.

# Prerequisites
* Golang 1.6 or above
* A parameters file with name *params.json* see section **Params.json  File format**
* Azure credetentials for every subscription to check. Refere to Azure Documentation on ho to  create credentials for Azure RESTT API 
[credentials for Azure RESTT API](https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-create-service-principal-portal)



# Launching script
* ***go run \*.go***
   
# How it Compare Tags

For a resource tag's
* If the key dosen't exist a ***NOTAG*** is exported to the output file 
* If the key exists (not case sensitive) and value match the regular expression the value is exported in output file
* If if the Key exists and the value is not matching the regular expression a ***TAGVALUEKO*** 

# Output
* A log file is created with name ***YYYYMMDD-HH-log.logs*** 
* The scriptis create ***outputs*** Directory if it no exists
* A file  by account is created in CSV format with the separator decalred in the ***params.json*** file. The Datas exported are : 
  * ###### Account Name
  * ###### Subscription ID
  * ###### Tenant ID
  * ###### Resource ID
  * ###### Resource Group Name
  * ###### Resource Name
  * ###### Resource Location
  * ###### TAG 1  
  * ###### TAG 2 
  * ###### TAG 3
  * ###### .....

# Params.json  File format
```json
{
   
    "Comments"  : [
                  "Name        : a common name. It will be used to create the name of the output file [Account Name+OutputFilename]",
                  "Provider    : Azure,  ***TO BE IMPLEMENTED***> AWS,GCP,...",
                  "Action      : ***TO BE IMPLEMENTED*** Action to be performed if the tags are not matching the regexp ",
                  "Exclude     : ***TO BE IMPLEMENTED*** Exclude resources from the check",
                  "Level       : RG for Resource Groups/ ALL for all resources within a subscription  / VM=Microsoft.Compute/virtualMachines ",
                  "Compare     : ***TO BE IMPLEMENTED*** yes : case sensitive / no : no case sensitive ",
                  "AutoShutdown: ***TO BE IMPLEMENTED*** Check if autoshutdown is active on  Virtual Machine  [only used with Levels :ALL and VM]",
                  "Backup      : ***TO BE IMPLEMENTED***Check if their is a Backup policy active on Virtual Machine  [only used with Levels :ALL and VM]",
                  "Credentials : for Azure : Subscription ID / Client ID / Client Secret / Tenant ID",
                  "OutputFilename  :  a YYYYMMDD-HHMM- is added for every file"
                  ],

    
    "AzureAuthentEndpoint" :"https://login.microsoftonline.com",
    "AzureResourcesEndpoint" :"https://management.azure.com",
    "AzureResourcesAPIVersion" :"2017-05-10",
    
    "OutputSeparator" :";",
    "OutputFilename" :"output.csv",

    "Accounts" : [
        {
            "Name" :  "{*A Given Name*}"  ,
            "Provider" : "{Azure}",
            "Action" : "{NOT YET IMPLEMENTED}",
            "Level"  : "{ALL | VM | RG}",
            "Exclude" : ["NOT YET IMPLEMENTED"],
            "Credentials" : {  
                "subscription_id" :  "{subscription ID}",
                "client_id" : "{ Application ID}",
                "client_secret" : "{Secret Key}",
                "tenant_id" : "{Tenant ID}"
            },
            "Tags" : [
                    {	"Key":"{your Key}","Value":"{Regular expression}" },
                    {	"Key":"{Regular expression}","Value":"{Regular expression}" },
                    {	"Key":"{Regular expression}","Value":"{Regular expression}"}
            ]
        },       
        {
            "Name" :  "{Another Given Name}"  ,
            "Provider" : "{Azure}",
            "Action" : "{NOT YET IMPLEMENTED}",
            "Level"  : "{ALL | VM | RG}",
            "Exclude" : ["NOT YET IMPLEMENTED"],
            "Credentials" : {  
                "subscription_id" :  "{subscription ID}",
                "client_id" : "{ Application ID}",
                "client_secret" : "{Secret Key}",
                "tenant_id" : "{Tenant ID}"
            },
            "Tags" : [
                    {	"Key":"{your Key}","Value":"{Regular expression}" },
                    {	"Key":"{your Key}","Value":"{Regular expression}" },
                    {	"Key":"{your Key}","Value":"{Regular expression}"}
            ]
        }   
                
        ]
}

