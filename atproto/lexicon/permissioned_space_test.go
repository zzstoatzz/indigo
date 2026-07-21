package lexicon

import (
	"encoding/json"
	"testing"
)

func parseAndCheckSchema(t *testing.T, raw string) SchemaFile {
	t.Helper()
	var schema SchemaFile
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	if err := schema.FinishParse(); err != nil {
		t.Fatalf("finish parsing schema: %v", err)
	}
	if err := schema.CheckSchema(); err != nil {
		t.Fatalf("check schema: %v", err)
	}
	return schema
}

func TestPermissionedSpaceSchemas(t *testing.T) {
	space := parseAndCheckSchema(t, `{
		"lexicon": 1,
		"id": "com.atmoboards.forum",
		"defs": {
			"main": {
				"type": "space",
				"description": "A discussion forum",
				"key": "any",
				"name": "AtmoBoards Forum",
				"name:lang": {"es": "Foro AtmoBoards"},
				"collections": ["com.atmoboards.thread", "com.atmoboards.reply"]
			}
		}
	}`)
	if _, ok := space.Defs["main"].Inner.(SchemaSpace); !ok {
		t.Fatalf("main definition is %T, want SchemaSpace", space.Defs["main"].Inner)
	}

	permissions := parseAndCheckSchema(t, `{
		"lexicon": 1,
		"id": "com.atmoboards.authForum",
		"defs": {
			"main": {
				"type": "permission-set",
				"title": "AtmoBoards",
				"permissions": [{
					"type": "permission",
					"resource": "space",
					"spaceType": "com.atmoboards.forum",
					"authority": "*",
					"skey": "*",
					"collection": ["com.atmoboards.thread", "*"],
					"action": ["read", "create"],
					"manage": ["update"]
				}]
			}
		}
	}`)
	main := permissions.Defs["main"].Inner.(SchemaPermissionSet)
	if got := main.Permissions[0].SpaceType; got != "com.atmoboards.forum" {
		t.Fatalf("spaceType = %q", got)
	}
}

func TestPermissionedSpaceSchemaValidation(t *testing.T) {
	tests := []struct {
		name string
		def  SchemaDef
	}{
		{
			name: "space without collections",
			def:  SchemaDef{Inner: SchemaSpace{Type: "space", Key: "any", Name: "Empty"}},
		},
		{
			name: "space permission with wildcard type",
			def: SchemaDef{Inner: SchemaPermission{
				Type: "permission", Resource: "space", SpaceType: "*",
			}},
		},
		{
			name: "space permission with unsupported action",
			def: SchemaDef{Inner: SchemaPermission{
				Type: "permission", Resource: "space", SpaceType: "com.atmoboards.forum", Action: []string{"share"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.def.CheckSchema(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
