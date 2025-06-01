// supported_vm_embed.go - Embeds supported VM SKU lists for ArcBox and others
package vm

import _ "embed"

// ArcBox ITPro supported VM SKUs
//
//go:embed arcbox/itpro/arcbox_supported_vm_sku.json
var ArcboxItproSupportedVmSku []byte

// ArcBox DevOps supported VM SKUs
//
//go:embed arcbox/devops/arcbox_supported_vm_sku.json
var ArcboxDevopsSupportedVmSku []byte

// ArcBox DataOps supported VM SKUs
//
//go:embed arcbox/dataops/arcbox_supported_vm_sku.json
var ArcboxDataopsSupportedVmSku []byte
