# Refactor Work Name Generation - Common Base Name Format

## Requirements
- Refactor `getWorkNamePrefixFromSnapshotName` function to use a common base name format for all work name generation
- Use `namespace.name` format for namespaced placements and just `name` for cluster-scoped placements consistently
- Remove redundant logic between cluster-scoped and namespaced placement handling
- Fix existing bug where `namespace` variable is used instead of `resourceSnapshot.GetNamespace()`
- Update tests to match the new unified format (tests currently expect `namespace-name` but should expect `namespace.name`)
- Ensure all existing functionality continues to work

## Additional comments from user
- The goal is to prevent naming conflicts between cluster-scoped and namespace-scoped placements
- Want to use a unified approach for work name generation
- Remove redundant logic and simplify the function

## Plan

### Phase 1: Analyze Current Implementation and Fix Existing Bug
- [x] Review current implementation of `getWorkNamePrefixFromSnapshotName`
- [x] Identify the bug where `namespace` is used instead of `resourceSnapshot.GetNamespace()`
- [x] Review existing format constants in `commons.go`
- [x] Review existing test cases to understand current expectations

### Phase 2: Fix Bug and Simplify Implementation
- [ ] Task 2.1: Fix the existing bug where `namespace` variable is used incorrectly
- [ ] Task 2.2: Simplify the function logic to remove redundant if/else blocks
- [ ] Task 2.3: Use a common base name generation approach for both scenarios

### Phase 3: Update Tests
- [ ] Task 3.1: Update test cases to expect the correct `namespace.name` format instead of `namespace-name`
- [ ] Task 3.2: Add additional test cases to ensure comprehensive coverage
- [ ] Task 3.3: Run tests to verify all changes work correctly

### Phase 4: Validation
- [ ] Task 4.1: Run unit tests to ensure all functionality works
- [ ] Task 4.2: Review all usages of the function to ensure no breaking changes
- [ ] Task 4.3: Verify the change achieves the goal of preventing naming conflicts

## Success Criteria
- `getWorkNamePrefixFromSnapshotName` function uses a unified approach for work name generation
- All existing tests pass with updated expectations
- No naming conflicts between cluster-scoped and namespaced placements
- Code is cleaner and more maintainable
- Bug with `namespace` variable is fixed

## Decisions
- Use the existing format constants that already follow the `namespace.name` pattern
- Keep the same basic logic flow but simplify the branching
- Update tests to match the intended behavior rather than the current buggy behavior

## Implementation Details

### Current Issues Found
1. Bug in line where `namespace` is used instead of `resourceSnapshot.GetNamespace()`
2. Redundant if/else logic that can be simplified
3. Tests expect `namespace-name` format but constants define `namespace.name` format
4. Function has duplicate logic for handling subindex scenarios

### Proposed Changes
1. Fix the namespace variable bug
2. Simplify to use a common base name generation approach
3. Update tests to expect correct format
4. Remove redundant branching logic

## Changes Made
- [ ] Fixed namespace variable bug in `getWorkNamePrefixFromSnapshotName`
- [ ] Simplified function logic to use unified approach
- [ ] Updated test cases to expect correct format
- [ ] Verified all tests pass

## Before/After Comparison
**Before**: Redundant logic with separate handling for cluster-scoped vs namespaced, bug with namespace variable
**After**: Unified approach with common base name format, bug fixed, cleaner code

## References
- `pkg/controllers/workgenerator/controller.go` - main function location
- `apis/placement/v1beta1/commons.go` - format constants
- `pkg/controllers/workgenerator/controller_test.go` - test cases
- User requirement to use common base name format and prevent naming conflicts
