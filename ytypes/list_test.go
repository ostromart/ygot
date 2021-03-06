// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ytypes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
)

var validListSchema = &yang.Entry{
	Kind:     yang.DirectoryEntry,
	ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
	Key:      "key_field_name",
	Config:   yang.TSTrue,
	Dir: map[string]*yang.Entry{
		"key_field_name": {
			Kind: yang.LeafEntry,
			Name: "key_field_name",
			Type: &yang.YangType{Kind: yang.Ystring},
		},
	},
}

func TestValidateListSchema(t *testing.T) {
	tests := []struct {
		desc    string
		schema  *yang.Entry
		wantErr bool
	}{
		{
			desc:   "success",
			schema: validListSchema,
		},
		{
			desc:    "nil schema",
			schema:  nil,
			wantErr: true,
		},
		{
			desc:    "bad schema type",
			schema:  &yang.Entry{Name: "nil-type-schema", Kind: yang.LeafEntry},
			wantErr: true,
		},
		{
			desc: "missing dir",
			schema: &yang.Entry{
				Name:   "missing-dir-schema",
				Kind:   yang.DirectoryEntry,
				Key:    "key_field_name",
				Config: yang.TSTrue,
			},
			wantErr: true,
		},
		{
			desc: "missing key field",
			schema: &yang.Entry{
				Name:     "missing-key-field-schema",
				Kind:     yang.DirectoryEntry,
				ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
				Config:   yang.TSTrue,
				Dir: map[string]*yang.Entry{
					"key_field_name": {
						Kind: yang.LeafEntry,
						Name: "key_field_name",
						Type: &yang.YangType{Kind: yang.Ystring},
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "missing key leaf",
			schema: &yang.Entry{
				Name:     "missing-key-leaf-schema",
				Kind:     yang.DirectoryEntry,
				ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
				Key:      "key_field_name",
				Config:   yang.TSTrue,
				Dir: map[string]*yang.Entry{
					"other_name": {
						Kind: yang.LeafEntry,
						Name: "other_name",
						Type: &yang.YangType{Kind: yang.Ystring},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		err := validateListSchema(test.schema)
		if got, want := (err != nil), test.wantErr; got != want {
			t.Errorf("%s: validateListSchema(%v) got error: %v, wanted error? %v", test.desc, test.schema, err, test.wantErr)
		}
		testErrLog(t, test.desc, err)
	}
}

func TestValidateListNoKey(t *testing.T) {
	listSchema := &yang.Entry{
		Name:     "list-schema",
		Kind:     yang.DirectoryEntry,
		ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
		Dir: map[string]*yang.Entry{
			"leaf-name": {
				Kind: yang.LeafEntry,
				Name: "LeafName",
				Type: &yang.YangType{Kind: yang.Ystring},
			},
		},
	}

	type StringListElemStruct struct {
		LeafName *string `path:"leaf-name"`
	}
	type BadElemStruct struct {
		UnknownName *string `path:"unknown-name"`
	}

	tests := []struct {
		desc    string
		schema  *yang.Entry
		val     interface{}
		wantErr bool
	}{
		{
			desc:   "success",
			schema: listSchema,
			val:    []*StringListElemStruct{{LeafName: ygot.String("elem1_leaf_name")}},
		},
		{
			desc:    "bad schema",
			schema:  nil,
			val:     []*StringListElemStruct{{LeafName: ygot.String("elem1_leaf_name")}},
			wantErr: true,
		},
		{
			desc:    "bad field",
			schema:  listSchema,
			val:     []*BadElemStruct{{UnknownName: ygot.String("elem1_leaf_name")}},
			wantErr: true,
		},
	}

	for _, test := range tests {
		err := Validate(test.schema, test.val)
		if got, want := (err != nil), test.wantErr; got != want {
			t.Errorf("%s: b.Validate(%v) got error: %v, wanted error? %v", test.desc, test.val, err, test.wantErr)
		}
		testErrLog(t, test.desc, err)
	}
}

func TestValidateListSimpleKey(t *testing.T) {
	listSchema := &yang.Entry{
		Name:     "list-schema",
		Kind:     yang.DirectoryEntry,
		ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
		Key:      "keyfield-name",
		Config:   yang.TSTrue,
		Dir: map[string]*yang.Entry{
			"keyfield-name": {
				Kind: yang.LeafEntry,
				Name: "keyfield-name",
				Type: &yang.YangType{Kind: yang.Ystring},
			},
			"leaf-name": {
				Kind: yang.LeafEntry,
				Name: "leaf-name",
				Type: &yang.YangType{Kind: yang.Ystring},
			},
		},
	}

	type StringListElemStruct struct {
		KeyFieldName *string `path:"keyfield-name"`
		LeafName     *string `path:"leaf-name"`
	}
	type BadElemStruct struct {
		LeafName *string
	}

	tests := []struct {
		desc    string
		val     interface{}
		wantErr bool
	}{
		{
			desc: "success",
			val: map[string]*StringListElemStruct{
				"elem1_key_val": {
					KeyFieldName: ygot.String("elem1_key_val"),
					LeafName:     ygot.String("elem1_leaf_name"),
				},
			},
		},
		{
			desc: "missing key",
			val: map[string]*BadElemStruct{
				"elem1": {
					LeafName: ygot.String("elem1_leaf_name"),
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		err := Validate(listSchema, test.val)
		if got, want := (err != nil), test.wantErr; got != want {
			t.Errorf("%s: b.Validate(%v) got error: %v, wanted error? %v", test.desc, test.val, err, test.wantErr)
		}
		testErrLog(t, test.desc, err)
	}
}

func TestValidateListStructKey(t *testing.T) {
	listSchemaStructKey := &yang.Entry{
		Name:     "list-schema-struct-key",
		Kind:     yang.DirectoryEntry,
		ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
		Key:      "Key1 Key2",
		Config:   yang.TSTrue,
		Dir: map[string]*yang.Entry{
			"key1": {
				Kind: yang.LeafEntry,
				Name: "Key1",
				Type: &yang.YangType{Kind: yang.Ystring},
			},
			"key2": {
				Kind: yang.LeafEntry,
				Name: "Key2",
				Type: &yang.YangType{Kind: yang.Yint32},
			},
			"leaf-name": {
				Kind: yang.LeafEntry,
				Name: "LeafName",
				Type: &yang.YangType{Kind: yang.Ystring},
			},
		},
	}

	type KeyStruct struct {
		Key1 string
		Key2 int32
	}
	type StringListElemStruct struct {
		Key1     *string `path:"key1"`
		Key2     *int32  `path:"key2"`
		LeafName *string `path:"leaf-name"`
	}
	type BadElemStruct1 struct {
		Key1     *string `path:"key1"`
		LeafName *string `path:"leaf-name"`
	}
	type BadElemStruct2 struct {
		Key1       *string `path:"key1"`
		Key2       *int32  `path:"key2"`
		ExtraField *string `path:"extra-name"`
		LeafName   *string `path:"leaf-name"`
	}

	tests := []struct {
		desc    string
		val     interface{}
		wantErr bool
	}{
		{
			desc: "success",
			val: map[KeyStruct]*StringListElemStruct{
				{"elem1_key_val", 1}: {
					Key1:     ygot.String("elem1_key_val"),
					Key2:     ygot.Int32(1),
					LeafName: ygot.String("elem1_leaf_name"),
				},
			},
		},
		{
			desc: "bad key value",
			val: map[KeyStruct]*StringListElemStruct{
				{"elem1_key_val", 1}: {
					Key1:     ygot.String("elem1_key_val"),
					Key2:     ygot.Int32(2),
					LeafName: ygot.String("elem1_leaf_name"),
				},
			},
			wantErr: true,
		},
		{
			desc: "missing key",
			val: map[KeyStruct]*BadElemStruct1{
				{"elem1_key_val", 0}: {
					Key1:     ygot.String("elem1_key_val"),
					LeafName: ygot.String("elem1_leaf_name"),
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		err := Validate(listSchemaStructKey, test.val)
		if got, want := (err != nil), test.wantErr; got != want {
			t.Errorf("%s: b.Validate(%v) got error: %v, wanted error? %v", test.desc, test.val, err, test.wantErr)
		}
		testErrLog(t, test.desc, err)
	}
}

func TestUnmarshalUnkeyedList(t *testing.T) {
	containerWithLeafListSchema := &yang.Entry{
		Name: "container",
		Kind: yang.DirectoryEntry,
		Dir: map[string]*yang.Entry{
			"struct-list": &yang.Entry{
				Name:     "struct-list",
				Kind:     yang.DirectoryEntry,
				ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
				Dir: map[string]*yang.Entry{
					"leaf-field": {
						Kind: yang.LeafEntry,
						Name: "leaf-field",
						Type: &yang.YangType{Kind: yang.Yint32},
					},
					"enum-leaf-field": {
						Kind: yang.LeafEntry,
						Name: "enum-leaf-field",
						Type: &yang.YangType{Kind: yang.Yenum},
					},
				},
			},
		},
	}

	type ListElemStruct struct {
		LeafName *int32   `path:"leaf-field"`
		EnumLeaf EnumType `path:"enum-leaf-field"`
	}
	type ContainerStruct struct {
		StructList []*ListElemStruct `path:"struct-list"`
	}

	tests := []struct {
		desc    string
		json    string
		want    ContainerStruct
		wantErr string
	}{
		{
			desc: "success",
			json: `{"struct-list" : [ { "leaf-field" : 42, "enum-leaf-field" : "E_VALUE_FORTY_TWO"} ] }`,
			want: ContainerStruct{
				StructList: []*ListElemStruct{
					&ListElemStruct{
						LeafName: ygot.Int32(42),
						EnumLeaf: 42,
					},
				},
			},
		},
	}

	var jsonTree interface{}
	for _, test := range tests {
		var parent ContainerStruct

		if err := json.Unmarshal([]byte(test.json), &jsonTree); err != nil {
			t.Fatal(fmt.Sprintf("%s : %s", test.desc, err))
		}

		err := Unmarshal(containerWithLeafListSchema, &parent, jsonTree)
		if got, want := errToString(err), test.wantErr; got != want {
			t.Errorf("%s: Unmarshal got error: %v, wanted error? %v", test.desc, got, want)
		}
		testErrLog(t, test.desc, err)
		if err == nil {
			if got, want := parent, test.want; !reflect.DeepEqual(got, want) {
				t.Errorf("%s: Unmarshal got:\n%v\nwant:\n%v\n", test.desc, pretty.Sprint(got), pretty.Sprint(want))
			}
		}
	}
}

func TestUnmarshalKeyedList(t *testing.T) {
	containerWithLeafListSchema := &yang.Entry{
		Name: "container",
		Kind: yang.DirectoryEntry,
		Dir: map[string]*yang.Entry{
			"key-list": &yang.Entry{
				Name:     "key-list",
				Kind:     yang.DirectoryEntry,
				ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
				Key:      "key",
				Config:   yang.TSTrue,
				Dir: map[string]*yang.Entry{
					"key": {
						Kind: yang.LeafEntry,
						Name: "key",
						Type: &yang.YangType{Kind: yang.Ystring},
					},
					"leaf-field": {
						Kind: yang.LeafEntry,
						Name: "leaf-field",
						Type: &yang.YangType{Kind: yang.Yint32},
					},
				},
			},
		},
	}

	type ListElemStruct struct {
		Key       *string `path:"key"`
		LeafField *int32  `path:"leaf-field"`
	}
	type ContainerStruct struct {
		KeyList map[string]*ListElemStruct `path:"key-list"`
	}

	tests := []struct {
		desc    string
		json    string
		want    ContainerStruct
		wantErr string
	}{
		{
			desc: "success",
			json: `{ "key-list" : [ { "key" : "forty-two", "leaf-field" : 42} ] }`,
			want: ContainerStruct{
				KeyList: map[string]*ListElemStruct{
					"forty-two": &ListElemStruct{
						Key:       ygot.String("forty-two"),
						LeafField: ygot.Int32(42),
					},
				},
			},
		},
		{
			desc:    "bad field",
			json:    `{ "key-list" : [ { "key" : "forty-two", "bad-field" : 42} ] }`,
			wantErr: `parent container key-list (type map[string]*ytypes.ListElemStruct): JSON contains unexpected field bad-field`,
		},
	}

	var jsonTree interface{}
	for _, test := range tests {
		var parent ContainerStruct

		if err := json.Unmarshal([]byte(test.json), &jsonTree); err != nil {
			t.Fatal(fmt.Sprintf("%s : %s", test.desc, err))
		}

		err := Unmarshal(containerWithLeafListSchema, &parent, jsonTree)
		if got, want := errToString(err), test.wantErr; got != want {
			t.Errorf("%s: Unmarshal got error: %v, wanted error? %v", test.desc, got, want)
		}
		testErrLog(t, test.desc, err)
		if err == nil {
			if got, want := parent, test.want; !reflect.DeepEqual(got, want) {
				t.Errorf("%s: Unmarshal got:\n%v\nwant:\n%v\n", test.desc, pretty.Sprint(got), pretty.Sprint(want))
			}
		}
	}
}

func TestUnmarshalStructKeyedList(t *testing.T) {
	containerWithLeafListSchema := &yang.Entry{
		Name: "container",
		Kind: yang.DirectoryEntry,
		Dir: map[string]*yang.Entry{
			"struct-key-list": &yang.Entry{
				Name:     "struct-key-list",
				Kind:     yang.DirectoryEntry,
				ListAttr: &yang.ListAttr{MinElements: &yang.Value{Name: "0"}},
				Key:      "key1 key2 key3",
				Config:   yang.TSTrue,
				Dir: map[string]*yang.Entry{
					"key1": {
						Kind: yang.LeafEntry,
						Name: "key1",
						Type: &yang.YangType{Kind: yang.Ystring},
					},
					"key2": {
						Kind: yang.LeafEntry,
						Name: "key2",
						Type: &yang.YangType{Kind: yang.Yint32},
					},
					"key3": {
						Kind: yang.LeafEntry,
						Name: "key3",
						Type: &yang.YangType{Kind: yang.Yenum},
					},
					"leaf-field": {
						Kind: yang.LeafEntry,
						Name: "leaf-field",
						Type: &yang.YangType{Kind: yang.Yint32},
					},
				},
			},
		},
	}

	type KeyStruct struct {
		Key1    string
		Key2    int32
		EnumKey EnumType
	}
	type ListElemStruct struct {
		Key1     *string  `path:"key1"`
		Key2     *int32   `path:"key2"`
		EnumKey  EnumType `path:"key3"`
		LeafName *int32   `path:"leaf-field"`
	}
	type ContainerStruct struct {
		StructKeyList map[KeyStruct]*ListElemStruct `path:"struct-key-list"`
	}

	tests := []struct {
		desc    string
		json    string
		want    ContainerStruct
		wantErr string
	}{
		{
			desc: "success",
			json: `{ "struct-key-list" : [ { "key1" : "forty-two", "key2" : 42, "key3" : "E_VALUE_FORTY_TWO", "leaf-field" : 43} ] }`,
			want: ContainerStruct{
				StructKeyList: map[KeyStruct]*ListElemStruct{
					{"forty-two", 42, 42}: &ListElemStruct{
						Key1:     ygot.String("forty-two"),
						Key2:     ygot.Int32(42),
						EnumKey:  42,
						LeafName: ygot.Int32(43),
					},
				},
			},
		},
	}

	var jsonTree interface{}
	for _, test := range tests {
		var parent ContainerStruct

		if err := json.Unmarshal([]byte(test.json), &jsonTree); err != nil {
			t.Fatal(fmt.Sprintf("%s : %s", test.desc, err))
		}

		err := Unmarshal(containerWithLeafListSchema, &parent, jsonTree)
		if got, want := errToString(err), test.wantErr; got != want {
			t.Errorf("%s: Unmarshal got error: %v, wanted error? %v", test.desc, got, want)
		}
		testErrLog(t, test.desc, err)
		if err == nil {
			if got, want := parent, test.want; !reflect.DeepEqual(got, want) {
				t.Errorf("%s: Unmarshal got:\n%v\nwant:\n%v\n", test.desc, pretty.Sprint(got), pretty.Sprint(want))
			}
		}
	}
}
