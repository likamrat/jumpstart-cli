# ArcBox Refactoring Strategy Document

## Executive Summary

Based on the comprehensive Phase 0.1 analysis, this document outlines specific refactoring patterns, service boundaries, migration strategies, and testing approaches for breaking down the monolithic 2164-line `cmd/arcbox/arcbox.go` file into maintainable, testable modules.

## 1. Code Grouping Analysis

### 1.1 Service Boundaries Identification

#### **Core Services**
- **DeploymentService**: Orchestrates ArcBox deployments
- **DiscoveryService**: Finds and analyzes ArcBox deployments
- **MonitoringService**: Tracks deployment progress and status
- **QuotaService**: Validates resource quotas and availability
- **ValidationService**: Handles preflight checks and parameter validation

#### **Support Services**
- **DisplayService**: Handles output formatting (table, JSON, spinners)
- **AzureClientService**: Wraps Azure CLI operations
- **CacheService**: Manages quota and region data caching

#### **Utility Packages**
- **Normalizers**: Pure functions for case normalization
- **Parsers**: Duration parsing, type conversion utilities
- **Validators**: Region and parameter validation helpers

### 1.2 Function Grouping by Responsibility

#### **Deployment Group** (High Cohesion)
```
Functions: deployArcboxWithParamFile, waitForDeploymentAndShowStatus, 
          getDeploymentResourceStatus, getDeploymentProvisioningState,
          printDeploymentResourceList, getAzureDeploymentDuration
Shared Dependencies: Azure CLI, deployment tracking, resource monitoring
Data Flow: Template processing → Deployment creation → Status monitoring → Result reporting
```

#### **Discovery Group** (High Cohesion)
```
Functions: runArcBoxList, discoverArcBoxDeployments, isArcBoxResourceGroup,
          hasArcBoxSolutionTag, hasArcBoxDeployments, hasArcBoxNamingPattern,
          enrichArcBoxDeployment, detectArcBoxFlavor
Shared Dependencies: Azure CLI, subscription management, resource analysis
Data Flow: Subscription scanning → Resource group filtering → Deployment identification → Metadata enrichment
```

#### **Display Group** (Medium Cohesion)
```
Functions: outputArcBoxDeploymentsTable, outputArcBoxDeploymentsJSON,
          printDeploymentErrorDetails, getStatusIcon, scanSubscriptionWithSpinner
Shared Dependencies: Formatting utilities, table library, color output
Data Flow: Data input → Format selection → Rendering → User output
```

#### **Utility Group** (Low Cohesion - Easy to Extract)
```
Functions: normalizeFlavorCase, normalizeSqlServerEditionCase, normalizeBastionSkuCase,
          parseISO8601Duration, parseInt64, validateLocations
Shared Dependencies: None or minimal
Data Flow: Input validation → Transformation → Output
```

### 1.3 Data Flow Mapping

#### **Primary Data Flows**
1. **Command → Deployment → Monitoring**
   ```
   NewArcboxCmdWithCLI → deployArcboxWithParamFile → waitForDeploymentAndShowStatus → Azure CLI
   ```

2. **Command → Discovery → Display**
   ```
   NewArcboxCmdWithCLI → runArcBoxList → discoverArcBoxDeployments → outputArcBoxDeploymentsTable
   ```

3. **Validation → Azure → Decision**
   ```
   Preflight checks → runQuotaChecksWithOutput → arcbox.RunQuotaChecks → Azure CLI
   ```

#### **Shared Data Dependencies**
- **quotaCache**: Used by quota checking functions
- **defaultAzureCLI**: Used by all Azure operations
- **region data**: Used by validation and quota functions

## 2. Dependency Analysis

### 2.1 Azure CLI Dependencies

#### **High Azure CLI Usage** (Requires Service Abstraction)
```go
// Functions with heavy Azure CLI dependency
discoverArcBoxDeployments  // 8+ azCLI method calls
waitForDeploymentAndShowStatus  // 6+ azCLI method calls  
deployArcboxWithParamFile  // 4+ azCLI method calls
isArcBoxResourceGroup  // 3+ azCLI method calls
```

#### **Medium Azure CLI Usage** (Service Integration Candidates)
```go
// Functions with moderate Azure CLI dependency
getDeploymentResourceStatus  // 2-3 azCLI method calls
enrichArcBoxDeployment  // 2-3 azCLI method calls
checkResourceGroupExists  // 1-2 azCLI method calls
```

#### **No Azure CLI Dependencies** (Pure Functions - Easy Extraction)
```go
// Pure utility functions
normalizeFlavorCase, normalizeSqlServerEditionCase, normalizeBastionSkuCase
parseISO8601Duration, parseInt64, getStatusIcon
getSupportedRegionsList, getSupportedRegionsDisplayList
```

### 2.2 Shared Data Structure Dependencies

#### **Command Structure Dependencies**
- All command creation functions depend on `cobra.Command`
- Flag definitions shared across deploy/list/preflight commands
- Common flag validation patterns

#### **Azure Resource Dependencies**
- `ArcBoxDeployment` struct used by discovery and display functions
- `AzureSubscription` struct used by list and discovery functions
- `resourceStatus` struct used by monitoring functions

#### **Caching Dependencies**
- `quotaCache` shared between quota checking functions
- Region data shared between validation functions

### 2.3 External Interface Dependencies

#### **Critical External Contracts** (Cannot Change)
```go
// main.go dependency
func NewArcboxCmd() *cobra.Command

// Preflight dependencies  
func ValidateConditionalRequirements(cmd *cobra.Command) bool
func RunArcBoxPreflightChecks(cmd *cobra.Command) bool

// Testing dependencies
func NewArcboxCmdWithCLI(cli azurecli.AzureCLI) *cobra.Command
func SetAzureCLI(cli azurecli.AzureCLI)
```

## 3. Testing Impact Analysis

### 3.1 Current Testing Patterns

#### **Dependency Injection Pattern** (Already Established)
```go
// Current pattern successfully used
mockCLI := azurecli.NewMockAzureCLI()
SetAzureCLI(mockCLI)
cmd := NewArcboxCmdWithCLI(mockCLI)

// Pattern proven in other packages:
// - cmd/subscription: 95.9% test coverage
// - internal/preflight/arcbox: 94.3% coverage
// - 200+ comprehensive tests across refactored packages
```

#### **Mock Integration Capabilities**
- Azure CLI wrapper provides comprehensive mocking
- Error injection capabilities for failure testing
- Data manipulation for various test scenarios

### 3.2 Functions Requiring New Test Patterns

#### **Service Layer Testing** (New Requirement)
```go
// Services will need interface-based testing
type DeploymentService interface {
    Deploy(ctx DeploymentContext) (*DeploymentResult, error)
    Monitor(deploymentID string) (<-chan DeploymentStatus, error)
}

// Test pattern for services
func TestDeploymentService_Deploy(t *testing.T) {
    mockAzureCLI := azurecli.NewMockAzureCLI()
    service := NewDeploymentService(mockAzureCLI)
    // Test service methods with mock
}
```

#### **Integration Testing Requirements**
- Command-to-service integration tests
- Service-to-service interaction tests  
- End-to-end workflow tests with mocked Azure CLI

### 3.3 Testing Strategy by Component

#### **Phase 1: Models and Utils** (Low Risk)
- **Strategy**: Unit tests for pure functions
- **Coverage Target**: 95%+
- **Mock Requirements**: None for pure functions

#### **Phase 2: Services** (Medium Risk)  
- **Strategy**: Interface-based unit tests + integration tests
- **Coverage Target**: 90%+
- **Mock Requirements**: Azure CLI wrapper mocking

#### **Phase 3: Commands** (High Risk)
- **Strategy**: Behavioral tests + regression tests
- **Coverage Target**: 85%+
- **Mock Requirements**: Service layer mocking + Azure CLI mocking

## 4. Migration Complexity Assessment

### 4.1 Complexity Rankings

#### **Easy Migration** (Pure Functions - No Dependencies)
```go
// Complexity: LOW | Risk: LOW | Effort: 1-2 hours each
normalizeFlavorCase()          // Line 1730-1745: Pure string normalization
normalizeSqlServerEditionCase() // Line 1747-1757: Pure string normalization  
normalizeBastionSkuCase()      // Line 1759-1769: Pure string normalization
parseISO8601Duration()         // Line 1690-1728: Pure parsing logic
parseInt64()                   // Line 2157-2169: Pure type conversion
getStatusIcon()                // Line 1604-1616: Pure mapping function
```

#### **Medium Migration** (Clear Dependencies)
```go
// Complexity: MEDIUM | Risk: MEDIUM | Effort: 4-6 hours each
getSubscriptionID()            // Line 1771-1790: Environment/flag access
validateLocations()            // Line 1870-1925: Region data dependency
outputArcBoxDeploymentsTable() // Line 1558-1591: Table formatting
outputArcBoxDeploymentsJSON()  // Line 1593-1602: JSON formatting  
clearQuotaCache()              // Line 34-36: Simple state management
```

#### **Hard Migration** (Complex Dependencies)
```go
// Complexity: HIGH | Risk: HIGH | Effort: 8-12 hours each
deployArcboxWithParamFile()    // Line 540-829: Multiple Azure operations
waitForDeploymentAndShowStatus() // Line 912-1035: Complex monitoring logic
discoverArcBoxDeployments()    // Line 1245-1272: Complex discovery logic
isArcBoxResourceGroup()        // Line 1274-1300: Multiple detection strategies
runArcBoxList()                // Line 1082-1122: Orchestration with UI
```

#### **Very Hard Migration** (Massive Refactoring)
```go
// Complexity: VERY HIGH | Risk: VERY HIGH | Effort: 16-24 hours
NewArcboxCmdWithCLI()          // Line 46-538: Massive command creation
```

### 4.2 Circular Dependencies

#### **Identified Circular Dependencies**
- **Command Creation ↔ Business Logic**: Commands embed business logic directly
- **Display ↔ Business Logic**: Business functions handle their own output formatting  
- **Caching ↔ Azure Operations**: Cache management mixed with Azure calls

#### **Dependency Breaking Strategy**
1. **Extract interfaces first** to define contracts
2. **Create service abstractions** to break direct dependencies
3. **Use dependency injection** to reverse control flow
4. **Separate concerns** (business logic vs. display vs. orchestration)

### 4.3 External Usage Risk Assessment

#### **High Risk Functions** (External Dependencies)
```go
// CRITICAL: Cannot change signatures
ValidateConditionalRequirements() // Called by preflight package
RunArcBoxPreflightChecks()        // Called by preflight package  
NewArcboxCmd()                    // Called by main.go
NewArcboxCmdWithCLI()             // Called by tests
SetAzureCLI()                     // Called by tests
```

#### **Medium Risk Functions** (Internal Dependencies)
```go
// Called via arcbox prefix - can refactor with care
runQuotaChecksWithOutput()        // Called by quota command
CreateResourceProviderCommands()  // Called by preflight command
CreateStatusCommand()             // Called by preflight command
```

## 5. Proposed Service Architecture

### 5.1 Service Boundaries and Responsibilities

#### **DeploymentService**
```go
type DeploymentService interface {
    ValidateParameters(ctx *DeploymentContext) error
    PrepareTemplate(ctx *DeploymentContext) (*Template, error)
    Deploy(ctx *DeploymentContext) (*DeploymentTracker, error)
    Monitor(tracker *DeploymentTracker) <-chan DeploymentEvent
}

// Responsibilities:
// - Parameter validation and normalization
// - Template preparation and parameter building
// - Azure deployment creation
// - Deployment monitoring and status updates
```

#### **DiscoveryService**
```go
type DiscoveryService interface {
    ListDeployments(criteria *SearchCriteria) ([]ArcBoxDeployment, error)
    AnalyzeResourceGroup(rgName string) (*ArcBoxAnalysis, error)
    DetectFlavor(deployment *ArcBoxDeployment) (string, error)
    EnrichMetadata(deployment *ArcBoxDeployment) error
}

// Responsibilities:
// - Multi-subscription scanning
// - ArcBox deployment identification
// - Metadata enrichment and analysis
// - Flavor detection and classification
```

#### **QuotaService**
```go
type QuotaService interface {
    CheckQuota(region, flavor string) (*QuotaResult, error)
    ValidateAvailability(skus []string, region string) (*AvailabilityResult, error)
    GetRequirements(flavor string) (*FlavorRequirements, error)
}

// Responsibilities:
// - vCPU quota validation
// - SKU availability checking
// - Flavor requirement mapping
// - Regional quota caching
```

#### **DisplayService**
```go
type DisplayService interface {
    FormatDeployments(deployments []ArcBoxDeployment, format string) error
    ShowProgress(message string, progress <-chan ProgressEvent) error
    DisplayStatus(resources []ResourceStatus) error
    ShowErrors(errors []error) error
}

// Responsibilities:
// - Multi-format output (table, JSON, YAML)
// - Progress indication and spinners
// - Error formatting and display
// - Status reporting
```

### 5.2 Service Interaction Patterns

#### **Command → Service → Azure CLI** (Dependency Flow)
```
Commands: Handle flags, validation, user interaction
   ↓
Services: Implement business logic, coordinate operations  
   ↓
Azure CLI Wrapper: Handle Azure operations, caching, error handling
```

#### **Service Composition Pattern**
```go
type ArcBoxOrchestrator struct {
    deployment *DeploymentService
    discovery  *DiscoveryService
    quota      *QuotaService
    display    *DisplayService
    azure      azurecli.AzureCLI
}

// Commands use orchestrator to coordinate service interactions
func (cmd *deployCommand) Execute() error {
    return cmd.orchestrator.DeployArcBox(cmd.context)
}
```

### 5.3 Interface Segregation Strategy

#### **Focused Azure CLI Interfaces**
```go
// Instead of using the full azurecli.AzureCLI interface everywhere
type DeploymentAzureOps interface {
    CreateDeployment(rg, name, template string, params []string) error
    GetDeployment(rg, name string) (*azurecli.DeploymentInfo, error)
    ListResources(rg string) ([]azurecli.ResourceInfo, error)
}

type DiscoveryAzureOps interface {
    ListSubscriptions() ([]azurecli.SubscriptionInfo, error) 
    ListResourceGroups() ([]azurecli.ResourceGroupInfo, error)
    ListResources(rg string) ([]azurecli.ResourceInfo, error)
}
```

## 6. Migration Order and Dependencies

### 6.1 Phase-by-Phase Migration Plan

#### **Phase 1: Foundation (1-2 weeks)**
```
Priority: LOW RISK - HIGH IMPACT
Order:
1. Extract models (ArcBoxDeployment, AzureSubscription, etc.)
2. Extract pure utility functions (normalizers, parsers)
3. Extract helper functions (region utilities, cache management)

Dependencies: None
Risk: Very Low  
Testing: Unit tests for pure functions
Validation: Compilation success + existing tests pass
```

#### **Phase 2: Services (2-3 weeks)**
```
Priority: MEDIUM RISK - HIGH IMPACT
Order:
1. Create DisplayService (low Azure CLI dependency)
2. Create QuotaService (medium Azure CLI dependency)  
3. Create DiscoveryService (high Azure CLI dependency)
4. Create DeploymentService (very high Azure CLI dependency)

Dependencies: Phase 1 completion
Risk: Medium
Testing: Service unit tests + integration tests
Validation: All services testable in isolation
```

#### **Phase 3: Command Refactoring (1-2 weeks)**
```
Priority: HIGH RISK - MEDIUM IMPACT
Order:
1. Refactor simple commands (list, preflight)
2. Refactor complex commands (deploy)
3. Update command creation logic

Dependencies: Phase 1-2 completion  
Risk: High (external interface dependencies)
Testing: Command integration tests + regression tests
Validation: All external contracts preserved
```

### 6.2 Dependency Resolution Strategy

#### **Circular Dependency Breaking**
1. **Interface First**: Define service interfaces before implementation
2. **Constructor Injection**: Pass dependencies through constructors
3. **Factory Pattern**: Use factories to manage service creation and wiring
4. **Event-Driven**: Use channels/events for decoupled communication

#### **Legacy Compatibility Maintenance**
```go
// Maintain existing functions as wrappers during transition
func deployArcboxWithParamFile(cmd *cobra.Command, args []string, ...) error {
    // Create service instances
    orchestrator := arcbox.NewOrchestrator(defaultAzureCLI)
    
    // Convert parameters
    ctx := convertToDeploymentContext(cmd, args)
    
    // Delegate to service
    return orchestrator.Deploy(ctx)
}
```

### 6.3 Risk Mitigation Strategies

#### **Breaking Changes Prevention**
- Keep all exported functions with original signatures
- Use internal refactoring without changing public APIs
- Add comprehensive regression tests
- Validate external package integrations

#### **Rollback Planning**
- Git branching strategy with atomic commits per service
- Feature flags for gradual service adoption
- Parallel implementation during transition
- Automated testing for rapid validation

## 7. Testing Strategy

### 7.1 Testing Approach by Phase

#### **Phase 1: Foundation Testing**
```go
// Pure function testing - no mocks needed
func TestNormalizeFlavorCase(t *testing.T) {
    tests := []struct {
        input    string
        expected string  
    }{
        {"itpro", "ITPro"},
        {"DEVOPS", "DevOps"},
        {"dataops", "DataOps"},
    }
    
    for _, tt := range tests {
        result := normalizeFlavorCase(tt.input)
        assert.Equal(t, tt.expected, result)
    }
}
```

#### **Phase 2: Service Testing**
```go
// Service testing with Azure CLI mocking
func TestDeploymentService_Deploy(t *testing.T) {
    mockAzureCLI := azurecli.NewMockAzureCLI()
    service := NewDeploymentService(mockAzureCLI)
    
    // Setup mock expectations
    mockAzureCLI.SetMockDeploymentResult("rg1", "deploy1", &azurecli.DeploymentInfo{
        ProvisioningState: "Succeeded",
    })
    
    // Test deployment
    ctx := &DeploymentContext{
        ResourceGroup: "rg1",
        Template:      "template.json",
    }
    
    tracker, err := service.Deploy(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, tracker)
}
```

#### **Phase 3: Integration Testing**
```go
// Command integration testing
func TestArcBoxDeployCommand_Integration(t *testing.T) {
    mockAzureCLI := azurecli.NewMockAzureCLI()
    cmd := NewArcboxCmdWithCLI(mockAzureCLI)
    
    // Setup comprehensive mock scenario
    setupDeploymentMocks(mockAzureCLI)
    
    // Execute command
    deployCmd := getSubcommand(cmd, "deploy")
    deployCmd.SetArgs([]string{
        "--resource-group", "test-rg",
        "--location", "eastus", 
        "--windows-user", "admin",
        "--windows-password", "SecurePass123!",
    })
    
    err := deployCmd.Execute()
    assert.NoError(t, err)
    
    // Verify Azure CLI interactions
    assert.True(t, mockAzureCLI.WasMethodCalled("CreateDeployment"))
}
```

### 7.2 Mock Strategy and Test Patterns

#### **Azure CLI Mock Patterns** (Already Proven)
```go
// Pattern established in other refactored packages
mockCLI := azurecli.NewMockAzureCLI()

// Data setup
mockCLI.SetMockSubscriptions([]azurecli.SubscriptionInfo{...})
mockCLI.SetMockResourceGroups([]azurecli.ResourceGroupInfo{...})

// Error injection for failure testing  
mockCLI.SetErrorForGetCurrentSubscription(errors.New("not logged in"))

// Behavior verification
assert.True(t, mockCLI.WasMethodCalled("ListResourceGroups"))
```

#### **Service Mock Strategy**
```go
// Service interfaces enable easy mocking
type MockDeploymentService struct {
    deploymentResults map[string]*DeploymentResult
    errors           map[string]error
}

func (m *MockDeploymentService) Deploy(ctx *DeploymentContext) (*DeploymentResult, error) {
    if err, exists := m.errors[ctx.ResourceGroup]; exists {
        return nil, err
    }
    return m.deploymentResults[ctx.ResourceGroup], nil
}
```

### 7.3 Quality Gates and Validation

#### **Coverage Targets** (Based on Proven Results)
- **Pure functions**: 95%+ coverage (easy to achieve)
- **Services**: 90%+ coverage (following established patterns)
- **Commands**: 85%+ coverage (focusing on critical paths)
- **Integration**: 80%+ coverage (end-to-end scenarios)

#### **Validation Checkpoints**
1. **Phase Completion Gates**: All tests pass + coverage targets met
2. **External Contract Validation**: Preflight integration still works
3. **Regression Testing**: All existing functionality preserved
4. **Performance Validation**: No significant performance degradation

## 8. Risk Assessment and Mitigation

### 8.1 Technical Risks

#### **High Risk: External Interface Changes**
- **Risk**: Breaking main.go or preflight package integration
- **Mitigation**: Comprehensive integration tests + parallel implementation
- **Detection**: Automated build validation across dependent packages

#### **Medium Risk: Service Interaction Complexity**
- **Risk**: Services not integrating properly or circular dependencies  
- **Mitigation**: Interface-first design + dependency injection
- **Detection**: Integration test coverage + dependency analysis tools

#### **Low Risk: Utility Function Extraction**
- **Risk**: Missing edge cases in pure function extraction
- **Mitigation**: Comprehensive unit test coverage
- **Detection**: Unit test execution + coverage analysis

### 8.2 Process Risks

#### **Timeline Risk: Underestimating Complexity**
- **Mitigation**: Phase-by-phase approach with explicit completion criteria
- **Buffer**: 20% additional time allocation per phase
- **Monitoring**: Daily progress tracking + blocker identification

#### **Quality Risk: Regression Introduction**
- **Mitigation**: Automated regression test suite + comprehensive mocking
- **Validation**: Continuous integration with full test execution
- **Rollback**: Git branching strategy enables rapid rollback

### 8.3 Success Metrics

#### **Objective Metrics**
- **Code Reduction**: Main file reduced from 2164 lines to <500 lines
- **Test Coverage**: Overall package coverage increased to 90%+
- **Function Count**: Main file functions reduced from 59 to <15
- **Dependency Count**: External dependencies clearly documented and maintained

#### **Subjective Metrics**
- **Maintainability**: New features can be added without touching main file
- **Testability**: All business logic can be unit tested in isolation
- **Readability**: Service responsibilities are clear and well-defined
- **Reusability**: Services can be reused by other commands (agora, localbox)

## 9. Conclusion

This refactoring strategy leverages proven patterns from the successful Azure CLI wrapper integration across other packages. The systematic approach prioritizes risk mitigation while achieving significant architectural improvements:

1. **Foundation First**: Extract low-risk pure functions and models
2. **Services Second**: Create testable business logic services  
3. **Commands Last**: Refactor command structure with preserved interfaces

The strategy ensures zero breaking changes while modernizing the architecture for better maintainability, testability, and reusability. The comprehensive testing approach, based on established patterns, ensures quality throughout the migration process.
