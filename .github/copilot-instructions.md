# Development Guidelines and Instructions

## Overview

This document provides guidelines to maintain code quality and prevent common development mistakes in the terraform-provider-mssqlpermissions project.

## Testing Guidelines

### 1. Always Test Real Production Code

#### ❌ WRONG: Testing Fake Helper Functions

```go
// DON'T DO THIS - Testing functions that don't exist in production code
func TestUserResourceErrorHandling(t *testing.T) {
    // Testing analyzeUserReadError() that only exists in tests
    shouldRemove, shouldAddError := analyzeUserReadError(err)
    // This gives false confidence - it's not testing real code!
}
```

#### ✅ CORRECT: Test Actual Production Code

```go
// DO THIS - Test the actual functions used in production
func TestHandleUserReadError(t *testing.T) {
    // Testing HandleUserReadError() that is actually used by user_resource.go
    result := HandleUserReadError(err)
    // This tests the real logic that runs in production!
}
```

**Key Principle:** Every test must verify the behavior of code that actually runs in production. If you create helper functions for testing, those same functions must be used by the production code.

### 2. Mandatory Unit Tests for Bug Fixes

**When you fix a bug, you MUST:**

1. **Create a unit test that would have caught the bug** before implementing the fix
2. **Verify the test fails** with the buggy code
3. **Apply the fix** and verify the test passes
4. **Document the scenario** in the test to prevent regression

**Example from the "external deletion" bug:**

```go
func TestHandleUserReadError(t *testing.T) {
    tests := []struct {
        name                  string
        err                   error
        expectedShouldRemove  bool  // This test would have caught the bug!
        expectedShouldAddError bool
    }{
        {
            name:                   "User not found - should remove from state",
            err:                    errors.New("user not found"),
            expectedShouldRemove:   true,  // Original bug: this was false
            expectedShouldAddError: false,
        },
        // ... more test cases
    }
}
```

### 3. Extract Logic for Better Testability

**Instead of inline error handling:**

```go
// Hard to test - logic is embedded in framework calls
func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    user, err := connector.GetUser(ctx, db, user)
    if err != nil && err.Error() == "user not found" {
        resp.State.RemoveResource(ctx)  // Hard to test this logic
        return
    }
    // ... rest of method
}
```

**Extract testable helper functions:**

```go
// Easy to test - logic is in pure functions
func HandleUserReadError(err error) ErrorHandlingResult {
    // Pure function - easy to test with various inputs
    if err == nil {
        return ErrorHandlingResult{ShouldRemoveFromState: false, ShouldAddError: false}
    }
    if err.Error() == "user not found" {
        return ErrorHandlingResult{ShouldRemoveFromState: true, ShouldAddError: false}
    }
    return ErrorHandlingResult{ShouldRemoveFromState: false, ShouldAddError: true}
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    user, err := connector.GetUser(ctx, db, user)
    result := HandleUserReadError(err)  // Use the testable function
    if result.ShouldRemoveFromState {
        resp.State.RemoveResource(ctx)
        return
    }
    // ... rest of method
}
```

## Code Review Checklist

### For Bug Fixes

- [ ] **Unit test added** that would have caught the bug
- [ ] **Test fails** without the fix
- [ ] **Test passes** with the fix
- [ ] **Similar code patterns** checked for the same issue
- [ ] **Documentation updated** if the behavior changed

### For New Features

- [ ] **Unit tests** cover the main logic paths
- [ ] **Error handling** is properly tested
- [ ] **Edge cases** are considered and tested
- [ ] **Integration tests** added if needed

### For Tests

- [ ] **Tests verify actual production code** (not test-only helper functions)
- [ ] **Test names** clearly describe the scenario being tested
- [ ] **Test cases** include both success and failure scenarios
- [ ] **Error conditions** are explicitly tested

### For Resource/Data Source Changes

- [ ] **Examples updated** - Check and update all relevant files in the `/examples` folder
- [ ] **Schema changes reflected** - Update example configurations when attributes are added/removed/modified
- [ ] **Output references updated** - Update any outputs.tf files that reference changed attributes
- [ ] **Documentation consistency** - Ensure README files and documentation match the new schema
- [ ] **Breaking changes documented** - Note any breaking changes in examples and provide migration guidance
- [ ] **Examples validated** - Run `task validate:examples:all` to ensure all examples work with the updated schema

**Key Examples to Check:**
- `/examples/resources/mssqlpermissions_[resource_name]/` - Resource-specific examples
- `/examples/complete-validation/` - End-to-end validation example
- `/examples/data-sources/` - Data source examples
- Any outputs or references in other example files

**After making changes, always validate:**
```bash
task validate:examples:all
```
This command validates all example configurations to ensure they work correctly with the updated provider schema.

## Common Anti-Patterns to Avoid

### 1. Test Theater

```go
// ❌ BAD: Testing logic that doesn't exist in production
func analyzeError(err error) (bool, bool) {
    // This function only exists in tests!
    return shouldRemove, shouldAddError
}
```

### 2. Missing Error Handling Tests

```go
// ❌ BAD: Only testing the happy path
func TestUserRead(t *testing.T) {
    // Only tests when everything works
    result := HandleUserReadError(nil)
    assert.False(t, result.ShouldAddError)
}

// ✅ GOOD: Testing error scenarios too
func TestUserRead(t *testing.T) {
    testCases := []struct{
        name string
        err error
        expected ErrorHandlingResult
    }{
        {"no error", nil, ErrorHandlingResult{...}},
        {"user not found", errors.New("user not found"), ErrorHandlingResult{...}},
        {"database error", errors.New("connection failed"), ErrorHandlingResult{...}},
    }
    // Test all scenarios!
}
```

### 3. Untestable Code

```go
// ❌ BAD: Logic mixed with framework calls
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Complex logic mixed with Terraform framework calls - hard to test
    if complexCondition1 && complexCondition2 {
        if anotherComplexCheck() {
            resp.State.RemoveResource(ctx)
        } else {
            resp.Diagnostics.AddError("Error", "message")
        }
    }
}

// ✅ GOOD: Extract logic into testable functions
func determineAction(condition1, condition2 bool, check func() bool) Action {
    // Pure function - easy to test
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    action := determineAction(condition1, condition2, anotherComplexCheck)
    // Apply the action using framework calls
}
```

## Example: How to Add Tests for Bug Fixes

When you discover a bug like the "external deletion" issue:

### Step 1: Understand the Bug

```text
Problem: When a resource is deleted externally, Terraform tries to delete it again instead of recreating it.
Root Cause: Read method sets empty state instead of calling RemoveResource()
```

### Step 2: Create a Failing Test

```go
func TestHandleUserReadError(t *testing.T) {
    // This test would FAIL with the buggy code
    result := HandleUserReadError(errors.New("user not found"))
    assert.True(t, result.ShouldRemoveFromState) // Would be false with bug
    assert.False(t, result.ShouldAddError)
}
```

### Step 3: Fix the Bug

```go
func HandleUserReadError(err error) ErrorHandlingResult {
    if err != nil && err.Error() == "user not found" {
        return ErrorHandlingResult{ShouldRemoveFromState: true}  // Fix!
    }
    // ... rest of logic
}
```

### Step 4: Verify Test Passes

```bash
go test ./internal/provider/ -run TestHandleUserReadError
# Should now pass
```

### Step 5: Check for Similar Issues

Search codebase for similar patterns and fix them consistently.

## Summary

**Remember:** The goal is not just to have tests, but to have **meaningful tests that actually protect against real bugs**. Every test should verify that production code behaves correctly under various conditions.

**When in doubt, ask:** "If I change this production code, will my test catch the regression?"

If the answer is no, the test needs to be improved or you need additional tests.
