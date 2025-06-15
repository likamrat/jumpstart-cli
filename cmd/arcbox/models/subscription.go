package models

// This file contains subscription related types and structs for ArcBox

// AzureSubscription represents an Azure subscription with ID and name
// Used for subscription management and discovery operations
type AzureSubscription struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
