package schemainput_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"

	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"

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
		shiftVersion     string
		expectProperties string
		err              error
	}{
		// 0
		{
			schemaInput: filepath.Join(schemaBasePath, "def_base", "def_variant1.json"),
			expectProperties: `{
	"{def_base}/type": {
		"state": "Uncovered"
	},
	"{def_variant1}/type": {
		"state": "Covered",
		"app_property_maps": [
		  {
		    "name": "app1",
			"addr": "type"
		  },
		  {
		    "name": "app2",
			"addr": "type"
		  }
		]
	},
	"{def_variant2}/type":{
		"state":"Uncovered"
	}
}`,
		},
		// 1
		{
			schemaInput: filepath.Join(schemaBasePath, "def_regular", "def_regular.json"),
			expectProperties: `{
	"prop_primitive": {
		"state": "Covered"
	},
	"prop_array_of_primitive":{
		"state":"Uncovered"
	},
	"prop_array_of_object/prop_nested1":{
		"state":"Uncovered"
	},
	"prop_array_of_object/prop_nested2":{
		"state":"Uncovered"
	},
	"prop_object/prop_nested":{
		"state":"Uncovered"
	}
}`,
		},
		// 2. New API version removes some property in old API
		{
			schemaInput:  filepath.Join(schemaBasePath, "def_regular", "def_regular.json"),
			shiftVersion: "2021-01-01",
			expectProperties: `{
	"prop_primitive": {
		"state": "Covered"
	},
	"prop_array_of_object/prop_nested1":{
		"state":"Uncovered"
	},
	"prop_array_of_object/prop_nested2":{
		"state":"Uncovered"
	},
	"prop_object/prop_nested":{
		"state":"Uncovered"
	}
}`,
		},
		// 3
		{
			schemaInput: filepath.Join(schemaBasePath, "def_bar", "def_bar.json"),
			expectProperties: `{
	"bar": {
		"state": "Covered"
	}
}`,
		},
		// 4. schema-less operation
		{
			schemaInput: filepath.Join(schemaBasePath, "def_foo", "def_foo_delete.json"),
		},
		// 5. New API version removed the API path
		{
			schemaInput:  filepath.Join(schemaBasePath, "def_bar", "def_bar.json"),
			shiftVersion: "2021-01-01",
			err:          workspace.ErrFailedToLocateSwagger,
		},
		{
			schemaInput: filepath.Join(schemaBasePath, "def_foo", "def_foo.json"),
			expectProperties: `{
	"foo": {
		"state": "Covered"
	}
}`,
		},
		// 6. New API version removed the API operation
		{
			schemaInput:  filepath.Join(schemaBasePath, "def_foo", "def_foo.json"),
			shiftVersion: "2021-01-01",
			err:          workspace.ErrFailedToLocateSwagger,
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir)
	require.NoError(t, err)

	for idx, c := range cases {
		b, err := ioutil.ReadFile(c.schemaInput)
		require.NoError(t, err, idx)

		var schemaInput schemainput.SchemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		if c.shiftVersion != "" {
			schemaInput.ApiVersion = c.shiftVersion
		}
		opInfo, err := schemaInput.ToOperationInfo(index, c.shiftVersion != "", false, nil)
		if c.err == nil {
			require.NoError(t, err, idx)
		} else {
			require.Condition(t, func() bool { return errors.Is(err, c.err) }, idx)
			continue
		}

		if opInfo.Model == nil {
			continue
		}

		b, err = opInfo.Model.Properties.MarshalJSON()
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expectProperties, string(b), idx)
	}
}

func TestSchemaInputMarshal(t *testing.T) {
	input := schemainput.SchemaInput{
		OperationLocator: workspace.OperationLocator{
			SpecLocator: workspace.SpecLocator{
				ApiVersion: "v1",
				ApiPath:    "p1",
			},
			Operation: "GET",
		},
		Properties: []schemainput.SchemaInputProperty{
			{
				Addr: "addr1",
			},
			{
				Addr: "addr2",
				AppPropertyMaps: []schemainput.AppPropertyMap{
					{
						Name: "app1",
						Addr: "app_addr",
					},
					{
						Name: "app2",
						Addr: "app_addr",
					},
				},
			},
		},
	}

	expect := `{
"api_version": "v1",
"api_path":"p1",
"operation":"GET",
"properties": [
  "addr1",
  {
	"addr": "addr2",
	"app_property_maps": [
	  {
	    "name": "app1",
		"addr": "app_addr"
	  },
	  {
	    "name": "app2",
		"addr": "app_addr"
	  }
	]
  }
]
}`

	b, err := json.Marshal(input)
	require.NoError(t, err)
	require.JSONEq(t, expect, string(b))
}

func TestSchemaInputUnmarshal(t *testing.T) {
	input := `{
"api_version": "v1",
"api_path":"p1",
"operation":"GET",
"properties": [
  "addr1",
  {
	"addr": "addr2",
	"app_property_maps": [
	  {
	    "name": "app1",
		"addr": "app_addr"
	  },
	  {
	    "name": "app2",
		"addr": "app_addr"
	  }
	]
  },
  {
	"addr": "addr3"
  }
]
}`
	expect := schemainput.SchemaInput{
		OperationLocator: workspace.OperationLocator{
			SpecLocator: workspace.SpecLocator{
				ApiVersion: "v1",
				ApiPath:    "p1",
			},
			Operation: "GET",
		},
		Properties: []schemainput.SchemaInputProperty{
			{
				Addr: "addr1",
			},
			{
				Addr: "addr2",
				AppPropertyMaps: []schemainput.AppPropertyMap{
					{
						Name: "app1",
						Addr: "app_addr",
					},
					{
						Name: "app2",
						Addr: "app_addr",
					},
				},
			},
			{
				Addr: "addr3",
			},
		},
	}

	var si schemainput.SchemaInput
	err := json.Unmarshal([]byte(input), &si)
	require.NoError(t, err)
	require.Equal(t, expect, si)
}
