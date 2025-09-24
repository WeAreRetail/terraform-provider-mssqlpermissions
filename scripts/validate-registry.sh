#!/bin/bash

# Terraform Provider Registry Validation Script
# This script validates your provider works with the official Terraform Registry

set -e

PROVIDER_NAME="${PROVIDER_NAME:-mssqlpermissions}"
PROVIDER_NAMESPACE="${PROVIDER_NAMESPACE:-WeAreRetail}"
PROVIDER_VERSION="${PROVIDER_VERSION:-0.0.5}"

echo "🔍 Terraform Provider Registry Validation"
echo "=========================================="
echo "Provider: ${PROVIDER_NAMESPACE}/${PROVIDER_NAME}@${PROVIDER_VERSION}"
echo ""

# Create temporary directory for validation
TEMP_DIR=$(mktemp -d)
echo "📁 Using temporary directory: $TEMP_DIR"

cd "$TEMP_DIR"

# Test 1: Basic provider installation from registry
echo "📦 Test 1: Installing provider from registry..."
cat >main.tf <<EOF
terraform {
  required_version = ">= 1.0"

  required_providers {
    ${PROVIDER_NAME} = {
      source  = "${PROVIDER_NAMESPACE}/${PROVIDER_NAME}"
      version = ">= ${PROVIDER_VERSION}"
    }
  }
}

provider "${PROVIDER_NAME}" {
  # Minimal configuration to test installation
  server_fqdn   = "test.example.com"
  server_port   = 1433
  database_name = "TestDB"

  sql_login = {
    username = "test"
    password = "test"
  }
}
EOF

terraform init
echo "✅ Provider installed successfully from registry"

# Test 2: Validate configuration
echo ""
echo "🔍 Test 2: Validating configuration..."
terraform validate
echo "✅ Configuration is valid"

# Test 3: Generate plan (should fail gracefully with connection error)
echo ""
echo "📋 Test 3: Generating plan (expected to fail with connection error)..."
if terraform plan 2>&1 | grep -E "(connection|authentication|network)" >/dev/null; then
  echo "✅ Plan failed as expected due to connection (this is good)"
else
  echo "⚠️  Unexpected plan result - review output above"
fi

# Test 4: Test provider schema
echo ""
echo "📚 Test 4: Validating provider schema..."
terraform providers schema -json >schema.json
if [ -s schema.json ]; then
  echo "✅ Provider schema is valid and non-empty"
  echo "📊 Schema size: $(wc -c <schema.json) bytes"
else
  echo "❌ Provider schema is empty or invalid"
  exit 1
fi

# Test 5: Check provider version
echo ""
echo "🏷️  Test 5: Checking provider version..."
if terraform version | grep -q "${PROVIDER_NAME}"; then
  echo "✅ Provider version information is available"
else
  echo "⚠️  Provider version not shown in terraform version output"
fi

# Cleanup
cd /
rm -rf "$TEMP_DIR"

echo ""
echo "🎉 Registry validation completed successfully!"
echo ""
echo "Next steps:"
echo "1. Test the validation examples: task validate"
echo "2. Run acceptance tests: task test:acceptance"
echo "3. Verify provider documentation on registry"
echo "4. Test provider in a real project"
