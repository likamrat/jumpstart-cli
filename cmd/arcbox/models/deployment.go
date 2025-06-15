package models

// This file contains deployment related types and structs for ArcBox

// ResourceStatus holds resource information for status output
// Used by deployment status reporting and resource listing functions
type ResourceStatus struct {
	Name  string
	Type  string
	State string
}

// ArcBoxDeployment represents an ArcBox deployment discovered in a resource group
type ArcBoxDeployment struct {
	ResourceGroupName string
	SubscriptionID    string
	SubscriptionName  string
	Location          string
	CreatedDate       string
	Status            string
	ResourceCount     int
	Flavor            string
	NamingPrefix      string
}
