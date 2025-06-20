# ArcBox Deployment Monitoring Fix - GitHub Copilot Prompts

## Overview

This document contains a structured set of GitHub Copilot prompts to fix the deployment monitoring regression in the `js arcbox deploy` command. The issues identified include inefficient polling, resource state caching flaws, excessive Azure CLI calls, and race conditions in status updates.

---

## Phase 1: Optimize Azure CLI Resource Querying

### Prompt 1.1: Batch Resource State Retrieval

```text
<!-- START PROMPT 1.1 -->
I need to optimize the getDeploymentResourceStatus method in cmd/arcbox/display/deployment_display.go. Currently it makes individual Azure CLI calls for each resource to get provisioning state, creating an N+1 query problem.

Current inefficient approach:
- Calls `az resource list` to get all resources
- Then calls `az resource show --ids <resourceId>` for each individual resource

Please refactor this to:
1. Use a single batched Azure CLI call to get all resource details with provisioning states
2. Implement parallel querying where batching isn't possible
3. Add error handling for individual resource query failures
4. Maintain the same return type: []models.ResourceStatus

The method should be more efficient and reduce the total number of Azure CLI calls from N+1 to ideally 1-2 calls total.
<!-- END PROMPT 1.1 -->
```

### Prompt 1.2: Implement Parallel Resource Queries

```text
<!-- START PROMPT 1.2 -->
In cmd/arcbox/display/deployment_display.go, I need to implement parallel resource querying as a fallback when batch operations aren't available.

Requirements:
1. Create a worker pool pattern to query multiple resources concurrently
2. Limit concurrency to avoid Azure API throttling (max 5-8 concurrent calls)
3. Use context with timeout for each individual resource query
4. Collect results and handle partial failures gracefully
5. Maintain original order of resources in the response
6. Add proper error logging for failed individual queries

The parallel querying should be used when the batch approach fails or isn't available, ensuring we still get better performance than sequential queries.
<!-- END PROMPT 1.2 -->
```

---

## Phase 2: Fix State Change Detection Logic

### Prompt 2.1: Redesign State Tracking

```text
<!-- START PROMPT 2.1 -->
I need to fix the state change detection logic in the WaitForDeploymentAndShowStatus method in cmd/arcbox/display/deployment_display.go.

Current problems:
1. State mapping is reset after each display, causing missed transitions
2. Resources showing repeatedly with identical states
3. Fast state transitions are not captured properly

Please redesign the state tracking to:
1. Maintain persistent state history across polling cycles
2. Detect meaningful state changes (not just any state difference)
3. Only trigger re-display when there are actual meaningful changes
4. Track state transition timestamps to detect rapid changes
5. Add logic to detect "stuck" resources that haven't changed in a while

The goal is to eliminate duplicate displays and capture all meaningful state transitions accurately.
<!-- END PROMPT 2.1 -->
```

### Prompt 2.2: Add Smart State Diffing

```text
<!-- START PROMPT 2.2 -->
In cmd/arcbox/display/deployment_display.go, I need to implement intelligent state diffing that can detect when to update the console display.

Create a smart diffing algorithm that:
1. Compares current resource states with previous states
2. Identifies new resources that appeared
3. Detects meaningful state transitions (Creating→Running→Succeeded)
4. Ignores transient or duplicate states
5. Groups related changes together for batch display updates
6. Handles edge cases like resources being deleted/recreated

The diffing should return a structured result indicating:
- New resources to display
- Changed resources with their state transitions
- Whether a full refresh is needed
- Summary of changes for logging

This will eliminate the repetitive resource listings seen in the current output.
<!-- END PROMPT 2.2 -->
```

---

## Phase 3: Implement Intelligent Polling Strategy

### Prompt 3.1: Adaptive Polling Intervals

```text
<!-- START PROMPT 3.1 -->
I need to replace the fixed 5-second polling interval in WaitForDeploymentAndShowStatus with an adaptive polling strategy.

Current issue: Fixed 5-second intervals cause either too many API calls or missed rapid state changes.

Implement adaptive polling that:
1. Starts with faster polling (2-3 seconds) during active deployment phases
2. Reduces frequency (10-15 seconds) when resources are stable
3. Increases frequency temporarily when new resources appear
4. Backs off exponentially when getting repeated identical states
5. Has different intervals for deployment state vs resource state polling
6. Respects Azure API rate limits and throttling signals

The adaptive strategy should:
- Reduce total API calls by 40-60%
- Capture state changes more reliably
- Automatically adjust based on deployment activity level
- Include configurable min/max intervals for testing
<!-- END PROMPT 3.1 -->
```

### Prompt 3.2: Add Deployment Phase Detection

```text
<!-- START PROMPT 3.2 -->
In cmd/arcbox/display/deployment_display.go, I need to add deployment phase detection to optimize polling behavior for different stages of ArcBox deployment.

Analyze the resource states and deployment progress to detect phases like:
1. **Initialization** (0-10% resources created) - Fast polling needed
2. **Infrastructure** (Basic resources: storage, networking, VMs) - Medium polling
3. **Configuration** (VM extensions, software installation) - Slower polling, longer timeouts
4. **Completion** (Final validations) - Fast polling for final state

For each phase:
- Adjust polling intervals appropriately
- Set different timeout expectations
- Customize progress indicators
- Handle phase-specific errors differently

Create a deployment phase detector that:
- Analyzes current resources and their types
- Estimates completion percentage
- Identifies current deployment phase
- Returns recommended polling strategy for that phase
- Provides user-friendly phase descriptions for display
<!-- END PROMPT 3.2 -->
```

---

## Phase 4: Enhance Error Handling and Resilience

### Prompt 4.1: Implement Robust Error Recovery

```text
<!-- START PROMPT 4.1 -->
I need to add robust error handling and recovery mechanisms to the deployment monitoring in cmd/arcbox/display/deployment_display.go.

Current issues:
- API timeouts cause complete monitoring failure
- Network hiccups interrupt the entire process
- No retry logic for transient failures

Implement comprehensive error handling:
1. **Retry Logic**: Exponential backoff for transient Azure CLI failures
2. **Partial Failure Handling**: Continue monitoring even if some resources fail to query
3. **Network Resilience**: Detect and recover from network issues
4. **API Throttling**: Detect Azure API throttling and back off appropriately
5. **Graceful Degradation**: Fall back to simplified monitoring if full monitoring fails

Error recovery should:
- Distinguish between retryable and permanent errors
- Log appropriate error levels (debug vs warning vs error)
- Maintain user experience with helpful error messages
- Continue monitoring other resources when individual queries fail
- Provide clear guidance when manual intervention is needed
<!-- END PROMPT 4.1 -->
```

### Prompt 4.2: Add Monitoring Health Checks

```text
<!-- START PROMPT 4.2 -->
In cmd/arcbox/display/deployment_display.go, I need to add health checks for the monitoring system itself to detect when monitoring is falling behind or failing.

Implement monitoring health checks that detect:
1. **Polling Lag**: When polls are taking longer than expected intervals
2. **API Health**: When Azure CLI calls are failing frequently
3. **State Staleness**: When resource states haven't updated in expected timeframes
4. **Missing Resources**: When expected resources don't appear
5. **Monitoring Performance**: Track monitoring overhead and efficiency

Health check system should:
- Run alongside the main monitoring loop
- Provide early warnings for monitoring issues
- Suggest remediation actions (e.g., "check network connectivity")
- Automatically adjust monitoring strategy when problems detected
- Log health metrics for debugging
- Display health status to user when appropriate

Include configurable thresholds for:
- Maximum acceptable polling delay
- API failure rate tolerance
- Resource state staleness limits
- Expected resource creation timeframes
<!-- END PROMPT 4.2 -->
```

---

## Phase 5: Optimize Console Output and User Experience

### Prompt 5.1: Redesign Progress Display

```text
<!-- START PROMPT 5.1 -->
I need to redesign the console output in cmd/arcbox/display/deployment_display.go to eliminate repetitive resource listings and provide better progress visualization.

Current problems:
- Same resources printed multiple times with identical states
- No clear progress indication
- Overwhelming output for users
- No deployment phase context

Redesign the display to:
1. **Progressive Updates**: Only show changes, not full resource lists each time
2. **Progress Summary**: Show completion percentage and estimated time remaining
3. **Phase Indicators**: Display current deployment phase clearly
4. **Grouped Updates**: Group related resource changes together
5. **Clean Transitions**: Clear previous output and update in-place when appropriate

New display format should include:
- Overall progress bar or percentage
- Current phase description
- Recently changed resources only
- Summary counters (X/Y resources succeeded)
- Estimated time remaining
- Clear visual separation between updates

The goal is to provide informative updates without overwhelming the user with repetitive information.
<!-- END PROMPT 5.1 -->
```

### Prompt 5.2: Add Interactive Progress Indicators

```text
<!-- START PROMPT 5.2 -->
In cmd/arcbox/display/deployment_display.go, I need to add interactive progress indicators that provide better feedback during long-running deployments.

Enhance the user experience with:
1. **Smart Spinner**: Context-aware spinner that changes based on current activity
2. **Progress Bars**: Show completion progress for different deployment phases
3. **Resource Counters**: Live counters showing resources in different states
4. **Time Indicators**: Elapsed time, estimated remaining time, phase durations
5. **Activity Feed**: Scrolling feed of recent significant changes

Interactive elements should:
- Update in real-time without overwhelming the console
- Provide meaningful context about what's currently happening
- Show both immediate progress and overall deployment progress
- Include helpful tips or next steps during long waits
- Be responsive to terminal size and capabilities
- Gracefully handle terminal resize events

Design the progress indicators to:
- Be informative but not distracting
- Work well in both interactive and CI/CD environments
- Provide appropriate detail level based on verbosity settings
- Include accessibility considerations for screen readers
<!-- END PROMPT 5.2 -->
```

---

## Phase 6: Performance Optimization and Caching

### Prompt 6.1: Implement Intelligent Caching

```text
<!-- START PROMPT 6.1 -->
I need to add intelligent caching to the deployment monitoring system in cmd/arcbox/display/deployment_display.go to reduce redundant Azure CLI calls.

Implement a caching strategy that:
1. **Resource State Caching**: Cache resource states with TTL based on resource type
2. **Deployment Info Caching**: Cache deployment-level information separately
3. **Smart Invalidation**: Invalidate cache when state changes are detected
4. **Selective Refresh**: Only refresh specific resources that are likely to have changed
5. **Cache Warming**: Pre-populate cache with expected resources

Caching logic should:
- Different TTL for different resource types (VMs vs Storage vs Extensions)
- Shorter cache for resources in transitional states
- Longer cache for resources in stable states (Succeeded/Failed)
- Invalidate cache intelligently based on deployment phase
- Handle cache misses gracefully
- Provide cache hit/miss metrics for optimization

The caching should reduce API calls by 50-70% while maintaining accuracy of state reporting.
<!-- END PROMPT 6.1 -->
```

### Prompt 6.2: Add Performance Metrics and Monitoring

```text
<!-- START PROMPT 6.2 -->
In cmd/arcbox/display/deployment_display.go, I need to add performance metrics collection to measure and optimize the monitoring system's efficiency.

Implement metrics collection for:
1. **API Call Metrics**: Track number, duration, and success rate of Azure CLI calls
2. **Polling Performance**: Measure actual vs intended polling intervals
3. **Cache Effectiveness**: Track cache hit rates and performance improvements
4. **Resource Query Times**: Monitor how long different resource types take to query
5. **State Change Detection**: Measure how quickly state changes are detected and displayed

Metrics system should:
- Collect metrics in real-time during deployment monitoring
- Calculate efficiency scores and performance indicators
- Identify bottlenecks and optimization opportunities
- Log performance data for analysis and debugging
- Provide optional verbose output for performance troubleshooting
- Export metrics in structured format for analysis

Include configurable metrics collection levels:
- Basic: Essential performance counters only
- Detailed: Comprehensive metrics for optimization
- Debug: Full tracing for development and troubleshooting

The metrics will help validate that the optimizations are working and identify further improvement opportunities.
<!-- END PROMPT 6.2 -->
```

---

## Phase 7: Testing and Validation

### Prompt 7.1: Create Comprehensive Test Suite

```text
<!-- START PROMPT 7.1 -->
I need to create comprehensive tests for the deployment monitoring improvements in cmd/arcbox/display/deployment_display.go and related files.

Create test coverage for:
1. **State Change Detection**: Test various state transition scenarios
2. **Polling Optimization**: Verify adaptive polling behaves correctly
3. **Error Handling**: Test recovery from various failure scenarios
4. **Performance**: Benchmark API call reduction and response times
5. **Edge Cases**: Handle unusual deployment scenarios and race conditions

Test scenarios should include:
- Fast-changing resources (quick state transitions)
- Slow deployments with long stable periods
- Network failures and API timeouts
- Partial resource query failures
- Azure API throttling simulation
- Large deployments with many resources
- Deployment failures at various stages

Tests should verify:
- Accuracy of state reporting
- Reduction in API calls (50-70% improvement target)
- Elimination of duplicate console output
- Proper error recovery and user messaging
- Performance improvements in different scenarios

Include both unit tests and integration tests that can run with mock Azure CLI responses.
<!-- END PROMPT 7.1 -->
```

### Prompt 7.2: Add Regression Testing Framework

```text
<!-- START PROMPT 7.2 -->
In the test suite for deployment monitoring, I need to add a regression testing framework that can validate the fixes against the specific issues identified in the original problem.

Create regression tests that specifically validate:
1. **No Duplicate Listings**: Ensure same resources aren't printed repeatedly with identical states
2. **Timely State Updates**: Verify that completed Azure resources are reported promptly
3. **Reduced API Calls**: Measure and assert on API call reduction targets
4. **State Transition Accuracy**: Ensure all meaningful state changes are captured
5. **Performance Benchmarks**: Validate that monitoring keeps up with deployment progress

Regression test framework should:
- Simulate the exact problematic scenario from the original issue
- Use recorded Azure CLI responses from real deployments
- Measure key performance indicators (API calls, display accuracy, timing)
- Provide clear pass/fail criteria for each optimization
- Generate detailed reports comparing before/after behavior
- Be runnable in CI/CD pipeline for continuous validation

Include test data from:
- The original problematic deployment scenario
- Various ArcBox flavors and configurations
- Different Azure regions and resource types
- Edge cases that previously caused issues

The regression tests should definitively prove that the identified issues have been resolved.
<!-- END PROMPT 7.2 -->
```

---

## Implementation Guidelines

### Development Approach

1. **Incremental Implementation**: Implement each phase separately to allow testing and validation
2. **Backward Compatibility**: Ensure changes don't break existing functionality
3. **Feature Flags**: Use configuration options to enable/disable optimizations during testing
4. **Monitoring**: Add extensive logging to track the effectiveness of each optimization

### Success Criteria

- **API Call Reduction**: Achieve 50-70% reduction in Azure CLI calls
- **Display Accuracy**: Eliminate duplicate resource listings
- **Responsiveness**: Reduce time gap between Azure backend completion and CLI reporting
- **User Experience**: Provide clearer, more informative progress indicators
- **Reliability**: Handle network issues and API throttling gracefully

### Testing Strategy

- Unit tests for individual components
- Integration tests with mock Azure CLI
- Performance benchmarks comparing before/after
- Regression tests for specific identified issues
- Manual testing with real Azure deployments

---

## Notes

- Each prompt should be used independently with GitHub Copilot
- Prompts are designed to build upon each other sequentially
- Include the relevant file paths and context in each Copilot session
- Test each phase thoroughly before proceeding to the next
- Consider the impact on other ArcBox commands when making changes
