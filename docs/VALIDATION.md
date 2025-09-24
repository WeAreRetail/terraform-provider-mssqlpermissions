# Terraform Provider Validation Guide

This guide covers validation best practices for the `terraform-provider-mssqlpermissions` provider, following industry standards from HashiCorp and the Terraform provider community.

## Validation Layers

Our provider implements a comprehensive 4-layer validation approach:

### 1. Unit Tests (Fastest - No Infrastructure)

```bash
task test:unit
```

- **Purpose**: Test individual functions and logic
- **Speed**: Very fast (< 1 minute)
- **Infrastructure**: None required
- **Coverage**: Business logic, validation functions, data transformations

### 2. Integration Tests (Fast - Local Infrastructure)

```bash
task test:integration
```

- **Purpose**: Test database operations against real MSSQL
- **Speed**: Fast (< 5 minutes)
- **Infrastructure**: Local Docker MSSQL
- **Coverage**: SQL queries, database connectivity, error handling

### 3. Acceptance Tests (Slow - Real Infrastructure)

```bash
task test:acceptance        # Local Docker
task test:acceptance:azure  # Azure Cloud
```

- **Purpose**: Test complete Terraform workflows (plan/apply/destroy)
- **Speed**: Slow (5-20 minutes)
- **Infrastructure**: Real MSSQL (local or cloud)
- **Coverage**: Full provider lifecycle, Terraform integration

### 4. Manual Validation (Complete - Production-like)

```bash
task validate               # Complete validation
task validate:local         # Local validation only
task validate:azure         # Azure validation only
```

- **Purpose**: End-to-end validation with real Terraform configurations
- **Speed**: Manual (varies)
- **Infrastructure**: Production-like environments
- **Coverage**: Real-world usage scenarios, documentation examples

## Manual Validation Workflows

### Quick Validation (Development)

```bash
# 1. Install provider locally
task install

# 2. Validate configuration syntax
task validate:plan-only

# 3. Validate all examples
task validate:examples
```

### Full Local Validation

```bash
# 1. Start local infrastructure
task infra:local:ready

# 2. Run comprehensive local validation
task validate:local

# Expected result: All resources created, validated, and destroyed
```

### Production Azure Validation

```bash
# 1. Set up Azure environment
source .azure.env

# 2. Run Azure validation
task validate:azure

# Expected result: All resources created against real Azure SQL
```

### Registry Validation

```bash
# Test provider installation from Terraform Registry
./scripts/validate-registry.sh
```

## Best Practice Guidelines

### What Other Providers Do

Based on analysis of popular providers (AWS, Azure, Google Cloud):

1. **Multi-Environment Testing**: ✅ Test against both local and cloud infrastructure
2. **Example-Driven Development**: ✅ Examples that double as tests and documentation
3. **Automated CI/CD**: ✅ Automated testing in GitHub Actions/similar
4. **Manual Validation Scripts**: ✅ Scripts for manual end-to-end testing
5. **Registry Integration**: ✅ Testing against the official registry
6. **Comprehensive Documentation**: ✅ Clear validation procedures

### Our Implementation

#### ✅ **What we do well:**

- Complete test pyramid (unit → integration → acceptance → manual)
- Multi-environment support (local Docker + Azure)
- Automated task management via Taskfile
- Real infrastructure testing
- Example configurations for validation
- Environment isolation with `.env` files

#### 🚀 **Recommended enhancements:**

- Manual validation examples (✅ **Added in this session**)
- Registry validation scripts (✅ **Added in this session**)
- Enhanced task automation (✅ **Added in this session**)

## Validation Checklist

Use this checklist before releasing a new provider version:

### Pre-Release Validation

- [ ] Unit tests pass: `task test:unit`
- [ ] Integration tests pass: `task test:integration`
- [ ] Local acceptance tests pass: `task test:acceptance`
- [ ] Azure acceptance tests pass: `task test:acceptance:azure`
- [ ] Manual local validation: `task validate:local`
- [ ] Manual Azure validation: `task validate:azure`
- [ ] All examples validate: `task validate:examples`
- [ ] Registry validation: `./scripts/validate-registry.sh`

### Post-Release Validation

- [ ] Provider installs from registry
- [ ] Documentation examples work
- [ ] Breaking changes are documented
- [ ] Version constraints are correct

## Common Validation Scenarios

### New Feature Development

1. Write unit tests first
2. Add integration tests for database operations
3. Create acceptance tests for Terraform workflows
4. Add example configuration in `examples/`
5. Run full validation suite

### Bug Fixes

1. Write failing test that reproduces the bug
2. Fix the issue
3. Ensure all tests pass
4. Run validation against affected scenarios

### Performance Testing

```bash
# Test with multiple resources
task test:acceptance
# Monitor database connections and query performance
# Check memory usage during large operations
```

### Regression Testing

```bash
# Run full test suite
task test:all

# Validate examples still work
task validate:examples

# Test against different Terraform versions
TF_ACC_TERRAFORM_VERSION=1.5.0 task test:acceptance
TF_ACC_TERRAFORM_VERSION=1.6.0 task test:acceptance
```

## Error Handling Validation

Test these common error scenarios:

1. **Database Connection Failures**
   - Invalid server FQDN
   - Wrong credentials
   - Network timeouts

2. **Permission Errors**
   - Insufficient database permissions
   - Missing authentication

3. **Resource Conflicts**
   - Duplicate user names
   - Role membership conflicts

4. **Configuration Errors**
   - Invalid Terraform syntax
   - Missing required parameters

## Continuous Integration

Our GitHub Actions workflow should include:

```yaml
# Example CI validation steps
- name: Unit Tests
  run: task test:unit

- name: Integration Tests
  run: task test:integration

- name: Acceptance Tests
  run: task test:acceptance

- name: Validate Examples
  run: task validate:examples
```

## Manual Testing Best Practices

### When to manually test

- Before major releases
- After significant refactoring
- When adding new authentication methods
- When updating dependencies
- Before publishing to registry

### What to test manually

- End-to-end workflows
- Error handling and recovery
- Documentation examples
- Cross-platform compatibility
- Different Terraform versions

### How to document results

- Record test scenarios and outcomes
- Note any manual steps required
- Document known limitations
- Update troubleshooting guides

## Conclusion

Your provider already follows most industry best practices! The validation enhancements added in this session provide:

- **Complete validation examples** for manual testing
- **Enhanced task automation** for common workflows
- **Registry validation** to ensure distribution works
- **Comprehensive documentation** of validation procedures

This approach ensures your provider is reliable, well-tested, and follows Terraform community standards.
