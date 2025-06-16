# ArcBox Codebase Analysis Report

## Executive Summary
- **Current file size**: 2164 lines
- **Number of functions**: 59 functions
- **Number of structs**: 4 structs
- **Key architectural patterns**: Monolithic command structure with embedded business logic, dependency injection via Azure CLI wrapper, centralized deployment orchestration

## Function Inventory

### Command Creation Functions
| Function Name | Line Range | Purpose | External Usage |
|---------------|------------|---------|----------------|
| NewArcboxCmd | 40-43 | Main command entry point (wrapper) | main.go |
| NewArcboxCmdWithCLI | 46-538 | Main command creation with DI | NewArcboxCmd, tests |
| deployArcboxWithParamFile | 540-829 | Deploy command implementation | NewArcboxCmdWithCLI |

### Business Logic Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| runArcBoxList | 1082-1122 | List deployment orchestration | Azure CLI, subscription management | NewArcboxCmdWithCLI |
| discoverArcBoxDeployments | 1245-1272 | Discover deployments in subscription | Azure CLI, resource group analysis | runArcBoxList |
| isArcBoxResourceGroup | 1274-1300 | Identify ArcBox resource groups | Azure CLI, resource analysis | discoverArcBoxDeployments |
| enrichArcBoxDeployment | 1350-1359 | Add metadata to deployments | Resource counting, flavor detection | discoverArcBoxDeployments |
| detectArcBoxFlavor | 1390-1430 | Determine ArcBox flavor from deployment | Azure CLI, deployment analysis | enrichArcBoxDeployment |
| detectArcBoxFlavorFallback | 1432-1470 | Fallback flavor detection | Resource type analysis | detectArcBoxFlavor |
| waitForDeploymentAndShowStatus | 912-1035 | Monitor deployment progress | Azure CLI, resource status | deployArcboxWithParamFile |
| scanSubscriptionWithSpinner | 1815-1850 | Scan subscription with UI feedback | Spinner animation, deployment discovery | runArcBoxList |

### Utility Functions
| Function Name | Line Range | Purpose | Pure Function? | External Usage |
|---------------|------------|---------|----------------|----------------|
| normalizeFlavorCase | 1730-1745 | Normalize flavor names | Yes | deployArcboxWithParamFile |
| normalizeSqlServerEditionCase | 1747-1757 | Normalize SQL edition names | Yes | deployArcboxWithParamFile |
| normalizeBastionSkuCase | 1759-1769 | Normalize Bastion SKU names | Yes | deployArcboxWithParamFile |
| getSubscriptionID | 1771-1790 | Get subscription from various sources | No (env access) | Multiple functions |
| setAzureSubscription | 1852-1857 | Set Azure CLI subscription | No (Azure CLI) | Multiple functions |
| checkResourceGroupExists | 1859-1868 | Check resource group existence | No (Azure CLI) | Multiple functions |
| validateLocations | 1870-1925 | Validate Azure regions | No (region data access) | Quota command |
| parseISO8601Duration | 1690-1728 | Parse Azure duration format | Yes | getAzureDeploymentDuration |
| parseInt64 | 2157-2169 | Safe type conversion | Yes | Quota functions |

### Display Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| printDeploymentResourceList | 883-910 | Display resource status list | Color formatting | waitForDeploymentAndShowStatus |
| outputArcBoxDeploymentsTable | 1558-1591 | Format deployments as table | Table utility | runArcBoxList |
| outputArcBoxDeploymentsJSON | 1593-1602 | Format deployments as JSON | JSON marshaling | runArcBoxList |
| getStatusIcon | 1604-1616 | Get status emoji | None | outputArcBoxDeploymentsTable |
| printDeploymentErrorDetails | 1634-1647 | Display deployment errors | Azure CLI | waitForDeploymentAndShowStatus |

### Azure Integration Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| getDeploymentResourceStatus | 835-861 | Get resource status list | Azure CLI | waitForDeploymentAndShowStatus |
| getDeploymentProvisioningState | 1618-1625 | Get deployment state | Azure CLI | waitForDeploymentAndShowStatus |
| getAzureDeploymentDuration | 1649-1688 | Get deployment duration | Azure CLI | waitForDeploymentAndShowStatus |
| getAllSubscriptions | 1124-1139 | List all subscriptions | Azure CLI | runArcBoxList |
| getCurrentSubscription | 1141-1152 | Get current subscription | Azure CLI | runArcBoxList |
| getSubscription | 1154-1165 | Get specific subscription | Azure CLI | runArcBoxList |

### Resource Analysis Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| hasArcBoxSolutionTag | 1302-1313 | Check for ArcBox solution tags | Azure CLI | isArcBoxResourceGroup |
| hasArcBoxDeployments | 1315-1327 | Check for ArcBox deployment names | Azure CLI | isArcBoxResourceGroup |
| hasArcBoxNamingPattern | 1329-1341 | Check for ArcBox naming patterns | Azure CLI | isArcBoxResourceGroup |
| hasArcBoxResources | 1343-1375 | Check for ArcBox resource types | Azure CLI | isArcBoxResourceGroup (unused) |
| getResourceCount | 1361-1368 | Count resources in RG | Azure CLI | enrichArcBoxDeployment |
| getResourceGroupCreationDate | 1370-1388 | Get RG creation date | Azure CLI | enrichArcBoxDeployment |
| getDeploymentStatus | 1472-1516 | Determine deployment status | Azure CLI | enrichArcBoxDeployment |

### Quota Management Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| runQuotaChecksWithOutput | 1984-2081 | Run quota checks with flexible output | Preflight quota module | Quota command |
| runQuotaChecksWithTable | 2083-2087 | Run quota checks with table output | runQuotaChecksWithOutput | Legacy (unused) |
| getFlavorSKUs | 2089-2099 | Get SKUs for ArcBox flavor | None | Quota functions |
| getRequiredVCPUForSKU | 2101-2112 | Get vCPU requirement for SKU | None | Quota functions |
| mapSKUToFamilyQuotaName | 2114-2142 | Map SKU to quota family | None | Quota functions |

### Cached Data Management
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| clearQuotaCache | 34-36 | Clear quota cache | None | Quota command |
| SetAzureCLI | 25-27 | Set Azure CLI for testing | None | Tests |

### Helper Functions
| Function Name | Line Range | Purpose | Dependencies | External Usage |
|---------------|------------|---------|--------------|----------------|
| getSupportedRegionsList | 1927-1933 | Get supported region names | None | validateLocations |
| getSupportedRegionsDisplayList | 1935-1941 | Get supported region display names | None | validateLocations |

## Critical External Dependencies

### Functions Called by Other Packages
| Function | Called From | Required Signature | Notes |
|----------|-------------|-------------------|-------|
| ValidateConditionalRequirements | internal/preflight/arcbox/arcbox.go | func(cmd *cobra.Command) bool | Must remain public, validates flavor-specific requirements |
| RunArcBoxPreflightChecks | internal/preflight/arcbox/arcbox.go | func(cmd *cobra.Command) bool | Must remain public, comprehensive preflight validation |
| RunQuotaChecks | internal/preflight/arcbox/quota.go | func(cli azurecli.AzureCLI, location, flavor string, subscription string) ([]QuotaCheckResult, error) | Called via arcbox prefix |
| CreateResourceProviderCommands | internal/preflight/arcbox/rp.go | func(cli azurecli.AzureCLI) *cobra.Command | Called via arcbox prefix |
| CreateStatusCommand | internal/preflight/arcbox/status.go | func() *cobra.Command | Called via arcbox prefix |

### Testing Dependencies
| Function | Purpose | Required For |
|----------|---------|--------------|
| NewArcboxCmdWithCLI | Dependency injection | All command tests |
| SetAzureCLI | Mock injection | Test setup and isolation |

## Data Flow Analysis

### Shared Data Structures
- **quotaCache**: Global map[string][]azurecli.VMUsageInfo - caches quota data per region
- **defaultAzureCLI**: Global azurecli.AzureCLI - default Azure CLI instance
- **resourceStatus**: Local struct for deployment monitoring
- **ArcBoxDeployment**: Local struct for deployment listing
- **AzureSubscription**: Local struct for subscription management

### Function Dependencies
- **Command Creation Chain**: NewArcboxCmd → NewArcboxCmdWithCLI → (subcommand functions)
- **Deployment Chain**: deployArcboxWithParamFile → waitForDeploymentAndShowStatus → getDeploymentResourceStatus/printDeploymentResourceList
- **List Chain**: runArcBoxList → scanSubscriptionWithSpinner → discoverArcBoxDeployments → isArcBoxResourceGroup → (detection functions)
- **Validation Chain**: Preflight functions → ValidateConditionalRequirements/RunArcBoxPreflightChecks
- **Quota Chain**: Quota command → runQuotaChecksWithOutput → arcbox.RunQuotaChecks

## Risk Assessment

### High-Risk Functions (require careful handling)
- **NewArcboxCmdWithCLI** (lines 46-538): Massive command creation function with complex flag definitions and embedded logic
- **deployArcboxWithParamFile** (lines 540-829): Complex deployment orchestration with multiple parameter handling branches
- **waitForDeploymentAndShowStatus** (lines 912-1035): Complex monitoring logic with spinner animation and resource tracking
- **discoverArcBoxDeployments** (lines 1245-1272): Complex subscription scanning with resource filtering
- **isArcBoxResourceGroup** (lines 1274-1300): Multiple detection strategies with performance implications

### Medium-Risk Functions
- **runArcBoxList** (lines 1082-1122): Orchestration logic but clearer boundaries
- **enrichArcBoxDeployment** (lines 1350-1359): Metadata gathering with multiple dependencies
- **detectArcBoxFlavor** (lines 1390-1430): Business logic but isolated scope
- **runQuotaChecksWithOutput** (lines 1984-2081): Quota logic with flexible output formatting

### Low-Risk Functions (safe to extract first)
- **normalizeFlavorCase, normalizeSqlServerEditionCase, normalizeBastionSkuCase**: Pure normalization functions
- **parseISO8601Duration**: Pure parsing function
- **parseInt64**: Pure utility function
- **getStatusIcon**: Pure display function
- **clearQuotaCache**: Simple cache management

## Proposed Migration Strategy

### Phase 1 Targets (Low Risk)
- **Models**: ArcBoxDeployment, AzureSubscription, resourceStatus structs
- **Normalizers**: normalizeFlavorCase, normalizeSqlServerEditionCase, normalizeBastionSkuCase
- **Parsers**: parseISO8601Duration, parseInt64
- **Simple utilities**: getStatusIcon, clearQuotaCache

### Phase 2 Targets (Medium Risk)
- **Azure helpers**: getSubscriptionID, setAzureSubscription, checkResourceGroupExists
- **Region utilities**: validateLocations, getSupportedRegionsList, getSupportedRegionsDisplayList
- **Quota utilities**: getFlavorSKUs, getRequiredVCPUForSKU, mapSKUToFamilyQuotaName

### Phase 3 Targets (High Risk)
- **Display services**: printDeploymentResourceList, outputArcBoxDeploymentsTable, outputArcBoxDeploymentsJSON
- **Azure services**: getDeploymentResourceStatus, getDeploymentProvisioningState, getAllSubscriptions
- **Discovery services**: Resource analysis functions (hasArcBoxSolutionTag, etc.)

### Phase 4 Targets (Highest Risk - Core Logic)
- **Deployment service**: deployArcboxWithParamFile, waitForDeploymentAndShowStatus
- **List service**: runArcBoxList, discoverArcBoxDeployments, scanSubscriptionWithSpinner
- **Command service**: NewArcboxCmdWithCLI refactoring

## Testing Strategy

### Current Test Coverage
- **Test patterns**: Dependency injection via SetAzureCLI and NewArcboxCmdWithCLI
- **Mock usage**: azurecli.MockAzureCLI for Azure CLI operations
- **Command testing**: Cobra command structure validation and flag testing
- **Coverage areas**: Basic command creation, some deployment functions

### Required New Tests
- **Service isolation**: Each extracted service needs comprehensive unit tests
- **Interface mocking**: New service interfaces need mock implementations
- **Integration testing**: Ensure refactored components work together
- **Regression testing**: Verify no behavioral changes during refactoring

## Architecture Improvements Required

### Dependency Injection
- **Current**: Global defaultAzureCLI variable with SetAzureCLI for testing
- **Target**: Service constructors accepting interfaces

### Interface Segregation
- **Current**: Direct azurecli.AzureCLI usage throughout
- **Target**: Focused interfaces for specific operations (DeploymentService, ListService, etc.)

### Error Handling
- **Current**: Mix of os.Exit calls and error returns
- **Target**: Consistent error return patterns in services, exit handling in commands only

### State Management
- **Current**: Global quotaCache variable
- **Target**: Service-managed state with proper lifecycle

## External Integration Points

### Main Package Integration
- **Location**: main.go line 61
- **Usage**: `rootCmd.AddCommand(arcbox.NewArcboxCmd())`
- **Requirement**: NewArcboxCmd() must remain unchanged

### Preflight Integration
- **Location**: internal/preflight/arcbox/arcbox.go lines 119, 129
- **Functions**: ValidateConditionalRequirements, RunArcBoxPreflightChecks
- **Requirement**: Function signatures must remain exactly the same

### Examples Integration
- **Location**: internal/examples/examples.go
- **Usage**: Examples referenced by keys like "arcbox.deploy", "arcbox.delete", "arcbox.list"
- **Requirement**: Example keys must remain unchanged

### Quota Integration
- **Location**: internal/preflight/arcbox/quota.go
- **Usage**: arcbox.RunQuotaChecks function called with specific signature
- **Requirement**: Function must remain accessible via arcbox package

## Conclusion

The ArcBox package is a complex monolithic structure that requires careful phase-by-phase refactoring. The main challenges are:

1. **High coupling**: Many functions directly depend on Azure CLI and global state
2. **Mixed responsibilities**: Commands, business logic, utilities, and display logic are intermingled
3. **External dependencies**: Several external packages rely on specific function signatures
4. **Complex state management**: Global variables and caching complicate extraction

The proposed approach prioritizes risk mitigation by starting with pure functions and gradually extracting more complex components while maintaining all external contracts.
