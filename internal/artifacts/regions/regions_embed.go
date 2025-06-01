// regions_embed.go - Embeds supported region lists for ArcBox and others
package regions

import _ "embed"

// ArcBox supported regions
//
//go:embed arcbox_supported_regions.json
var ArcboxSupportedRegionsData []byte
