package main

import (
	"io"
	"log"
)

// To be used for logging purpose
var (
	Trace    *log.Logger
	Info     *log.Logger
	Warning  *log.Logger
	Error    *log.Logger
	logsFile io.Writer
)

// the parameters strucutre
type actObject struct {
	Description              string       `json:"Description"`
	Comment                  string       `json:"Comment"`
	AzureAuthentEndpoint     string       `json:"AzureAuthentEndpoint"`
	AzureResourcesEndpoint   string       `json:"AzureResourcesEndpoint"`
	AzureResourcesAPIVersion string       `json:"AzureResourcesAPIVersion"`
	OutputSeparator          string       `json:"OutputSeparator"`
	OutputFilename           string       `json:"OutputFilename"`
	Accounts                 []actAccount `json:"Accounts"`
}

// the account object
type actAccount struct {
	Name        string         `json:"Name"`
	Provider    string         `json:"Provider"`
	Action      string         `json:"Action"`
	Level       string         `json:"Level"`
	Exclude     []string       `json:"Exclude"`
	Credentials actCredentials `json:"Credentials"`
	Tags        []actTags      `json:"Tags"`
}

// Azure credentials
type actCredentials struct {
	SubscriptionID string `json:"subscription_id"`
	ApplicationID  string `json:"client_id"`
	KeySecret      string `json:"client_secret"`
	Tenant         string `json:"tenant_id"`
}

// Account Tags
type actTags struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// Resource Groups
type azResourceGroups struct {
	NextLink string            `json:"nextLink"`
	Valeur   []azResourceGroup `json:"value"`
}

type azResourceGroup struct {
	ID         string            `json:"id"`
	Location   string            `json:"location"`
	Name       string            `json:"name"`
	Tags       map[string]string `json:"tags"`
	Properties map[string]string `json:"properties"`
}

// Resources Structure
type azResources struct {
	NextLink string       `json:"nextLink"`
	Valeur   []azResource `json:"value"`
}

type azResource struct {
	ID         string            `json:"id"`
	Location   string            `json:"location"`
	ResType    string            `json:"type"`
	Name       string            `json:"name"`
	Plan       azPlan            `json:"plan"`
	Kind       string            `json:"kind"`
	ManagedBy  string            `json:"managedBy"`
	Sku        azSku             `json:"sku"`
	Identity   azIdentity        `json:"identity"`
	Tags       map[string]string `json:"tags"`
	Properties map[string]string `json:"properties"`
}

type azPlan struct {
	Name          string `json:"name"`
	Publisher     string `json:"publisher"`
	Product       string `json:"product"`
	PromotionCode string `json:"promotionCode"`
}

type azSku struct {
	Name     string `json:"name"`
	Tier     string `json:"tier"`
	Size     string `json:"size"`
	Family   string `json:"family"`
	Model    string `json:"model"`
	Capacity int32  `json:"capacity"`
}

type azIdentity struct {
	PrincipalID string `json:"principalId"`
	TenantID    string `json:"tenantId"`
	ResType     string `json:"type"`
}

type azOutputLine struct {
	AccountName       string
	TenantID          string
	SubscriptionName  string
	ResourceName      string
	ResourceID        string
	ResourceGroupName string
	ResourceType      string
	ResourceLocation  string
	Tags              map[string]string
}
