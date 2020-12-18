package schemainput

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"

	"github.com/stretchr/testify/require"
)

func TestToSchema(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	schemaBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "schema_input")
	cases := []struct {
		schemaInput      string
		expectProperties string
	}{
		{
			schemaInput: filepath.Join(schemaBasePath, "def_base", "def_variant1.json"),
			expectProperties: `{
	"{def_variant1}/type": {
		"state": "Covered"
	},
	"{def_variant2}":{
		"state":"Uncovered"
	}
}`,
		},
		{
			schemaInput: filepath.Join(schemaBasePath, "def_regular", "def_regular.json"),
			expectProperties: `{
	"prop_primitive": {
		"state": "Covered"
	},
	"prop_array_of_primitive":{
		"state":"Uncovered"
	},
	"prop_array_of_object":{
		"state":"Uncovered"
	},
	"prop_object":{
		"state":"Uncovered"
	}
}`,
		},
		{
			schemaInput: filepath.Join(schemaBasePath, "def_bar", "def_bar.json"),
			expectProperties: `{
	"bar": {
		"state": "Covered"
	}
}`,
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir, "")
	require.NoError(t, err)

	for idx, c := range cases {
		b, err := ioutil.ReadFile(c.schemaInput)
		require.NoError(t, err, idx)

		var schemaInput SchemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		schema, err := schemaInput.ToSchema(index)
		require.NoError(t, err, idx)

		b, err = schema.Properties.MarshalJSON()
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expectProperties, string(b), idx)
	}
}
