// SPDX-FileCopyrightText: 2024 AWARE - Altogether We Are Retailers
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestUserResource_ObjectIDRequiresReplace asserts that object_id forces replacement.
//
// object_id cannot be read back from the database: GetUser selects from sys.database_principals,
// which has no object_id column. It is therefore preserved from prior state rather than refreshed,
// and UpdateUser has no way to rebind an existing user to a different principal — ALTER USER only
// touches DEFAULT_SCHEMA, DEFAULT_LANGUAGE and PASSWORD. Planning an in-place update for a changed
// object_id would silently leave the user bound to the old principal, so the only correct action is
// to replace it.
func TestUserResource_ObjectIDRequiresReplace(t *testing.T) {
	r := NewUserResource()
	ctx := context.Background()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("Expected schema attributes to be defined")
	}

	attr, exists := resp.Schema.Attributes["object_id"]
	if !exists {
		t.Fatal("Expected object_id attribute to exist")
	}

	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("Expected object_id to be a StringAttribute, got %T", attr)
	}

	if len(stringAttr.PlanModifiers) == 0 {
		t.Fatal("Expected object_id to declare a RequiresReplace plan modifier, but it has none: " +
			"a changed object_id would be planned as an in-place update that UpdateUser cannot apply")
	}

	// The framework's RequiresReplace modifier describes itself as destroying and recreating the
	// resource; assert on that rather than invoking the modifier, which no-ops unless it is handed a
	// fully populated plan/state.
	found := false
	for _, m := range stringAttr.PlanModifiers {
		if strings.Contains(strings.ToLower(m.Description(ctx)), "destroy and recreate") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected one of object_id's plan modifiers to require destroy-and-recreate")
	}
}
