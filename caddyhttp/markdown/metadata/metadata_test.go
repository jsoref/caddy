// Copyright 2015 Light Code Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metadata

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

var TOML = [5]string{`
title = "A title"
template = "default"
name = "value"
positive = true
negative = false
number = 1410
float = 1410.07
`,
	`+++
title = "A title"
template = "default"
name = "value"
positive = true
negative = false
number = 1410
float = 1410.07
+++
Page content
	`,
	`+++
title = "A title"
template = "default"
name = "value"
positive = true
negative = false
number = 1410
float = 1410.07
	`,
	`title = "A title" template = "default" [variables] name = "value"`,
	`+++
title = "A title"
template = "default"
name = "value"
positive = true
negative = false
number = 1410
float = 1410.07
+++
`,
}

var YAML = [5]string{`
title : A title
template : default
name : value
positive : true
negative : false
number : 1410
float : 1410.07
`,
	`---
title : A title
template : default
name : value
positive : true
negative : false
number : 1410
float : 1410.07
---
	Page content
	`,
	`---
title : A title
template : default
name : value
number : 1410
float : 1410.07
	`,
	`title : A title template : default variables : name : value : positive : true : negative : false`,
	`---
title : A title
template : default
name : value
positive : true
negative : false
number : 1410
float : 1410.07
---
`,
}

var JSON = [5]string{`
	"title" : "A title",
	"template" : "default",
	"name" : "value",
	"positive" : true,
	"negative" : false,
	"number": 1410,
	"float": 1410.07
`,
	`{
	"title" : "A title",
	"template" : "default",
	"name" : "value",
	"positive" : true,
	"negative" : false,
	"number" : 1410,
	"float": 1410.07
}
Page content
	`,
	`
{
	"title" : "A title",
	"template" : "default",
	"name" : "value",
	"positive" : true,
	"negative" : false,
	"number" : 1410,
	"float": 1410.07
	`,
	`
{
	"title" :: "A title",
	"template" : "default",
	"name" : "value",
	"positive" : true,
	"negative" : false,
	"number" : 1410,
	"float": 1410.07
}
	`,
	`{
	"title" : "A title",
	"template" : "default",
	"name" : "value",
	"positive" : true,
	"negative" : false,
	"number" : 1410,
	"float": 1410.07
}
`,
}

func TestParsers(t *testing.T) {
	expected := Metadata{
		Title:    "A title",
		Template: "default",
		Variables: map[string]interface{}{
			"name":     "value",
			"title":    "A title",
			"template": "default",
			"number":   1410,
			"float":    1410.07,
			"positive": true,
			"negative": false,
		},
	}
	compare := func(m Metadata) bool {
		if m.Title != expected.Title {
			return false
		}
		if m.Template != expected.Template {
			return false
		}
		for k, v := range m.Variables {
			if fmt.Sprintf("%v", v) != fmt.Sprintf("%v", expected.Variables[k]) {
				return false
			}
		}

		varLenOK := len(m.Variables) == len(expected.Variables)
		return varLenOK
	}

	data := []struct {
		parser   Parser
		testData [5]string
		name     string
	}{
		{&JSONParser{}, JSON, "JSON"},
		{&YAMLParser{}, YAML, "YAML"},
		{&TOMLParser{}, TOML, "TOML"},
	}

	for _, v := range data {
		// metadata without identifiers
		if v.parser.Init(bytes.NewBufferString(v.testData[0])) {
			t.Fatalf("Expected error for invalid metadata for %v", v.name)
		}

		// metadata with identifiers
		if !v.parser.Init(bytes.NewBufferString(v.testData[1])) {
			t.Fatalf("Metadata failed to initialize, type %v", v.parser.Type())
		}
		md := v.parser.Markdown()
		if !compare(v.parser.Metadata()) {
			t.Fatalf("Expected %v, found %v for %v", expected, v.parser.Metadata(), v.name)
		}
		if "Page content" != strings.TrimSpace(string(md)) {
			t.Fatalf("Expected %v, found %v for %v", "Page content", string(md), v.name)
		}
		// Check that we find the correct metadata parser type
		if p := GetParser([]byte(v.testData[1])); p.Type() != v.name {
			t.Fatalf("Wrong parser found, expected %v, found %v", v.name, p.Type())
		}

		// metadata without closing identifier
		if v.parser.Init(bytes.NewBufferString(v.testData[2])) {
			t.Fatalf("Expected error for missing closing identifier for %v parser", v.name)
		}

		// invalid metadata
		if v.parser.Init(bytes.NewBufferString(v.testData[3])) {
			t.Fatalf("Expected error for invalid metadata for %v", v.name)
		}

		// front matter but no body
		if !v.parser.Init(bytes.NewBufferString(v.testData[4])) {
			t.Fatalf("Unexpected error for valid metadata but no body for %v", v.name)
		}
	}
}

func TestLargeBody(t *testing.T) {

	var JSON = `{
"template": "chapter"
}


	`
	var TOML = `+++
template = "chapter"
+++


	`
	var YAML = `---
template : chapter
---


	`
	var NONE = `


	`
`

	data := []struct {
		pType    string
		testData string
	}{
		{"JSON", JSON},
		{"TOML", TOML},
		{"YAML", YAML},
		{"None", NONE},
	}
	for _, v := range data {
		p := GetParser([]byte(v.testData))
		if v.pType != p.Type() {
			t.Fatalf("Wrong parser type, expected %v, got %v", v.pType, p.Type())
		}
		md := p.Markdown()
		if strings.TrimSpace(string(md)) != strings.TrimSpace(expectedBody) {
			t.Log("Provided:", v.testData)
			t.Log("Returned:", p.Markdown())
			t.Fatalf("Error, mismatched body in expected type %v, matched type %v", v.pType, p.Type())
		}
	}
}
