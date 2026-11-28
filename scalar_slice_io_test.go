package env

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScalarSliceFileBindingDoesNotReadTwice(t *testing.T) {
	type config struct {
		Values []scalarSliceText `env:"VALUES,file,expand" envPrefix:"ITEM_"`
	}
	filename := filepath.Join(t.TempDir(), "values")
	if err := os.WriteFile(filename, []byte("one,two"), 0o600); err != nil {
		t.Fatal(err)
	}
	var cfg config
	var hooks []string
	err := ParseWithOptions(&cfg, Options{
		Environment: map[string]string{
			"VALUES": "$SOURCE", "SOURCE": filename,
			"ITEM_0_VALUE": "wrong", "ITEM_1_VALUE": "also-wrong",
		},
		UseFieldNameByDefault: true,
		RequiredIfNoDef:       true,
		OnSet: func(key string, _ interface{}, _ bool) {
			hooks = append(hooks, key)
			if key == "VALUES" {
				if err := os.Remove(filename); err != nil {
					t.Error(err)
				}
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []scalarSliceText{{"one", 1}, {"two", 1}}
	if !reflect.DeepEqual(cfg.Values, want) {
		t.Errorf("values = %#v, want %#v", cfg.Values, want)
	}
	if !reflect.DeepEqual(hooks, []string{"VALUES"}) {
		t.Errorf("hooks = %v, want only VALUES", hooks)
	}
}

func TestScalarSliceFieldParamsDoNotTraverseElements(t *testing.T) {
	type config struct {
		Values []scalarSliceText `setting:"VALUES" itemPrefix:"ITEM_"`
	}
	var cfg config
	params, err := GetFieldParamsWithOptions(&cfg, Options{
		TagName: "setting", PrefixTagName: "itemPrefix", Prefix: "APP_",
		Environment: map[string]string{
			"APP_VALUES": "one,two", "APP_ITEM_0_VALUE": "wrong",
		},
		UseFieldNameByDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(params) != 1 || params[0].Key != "APP_VALUES" {
		t.Errorf("params = %#v, want only APP_VALUES", params)
	}
	if cfg.Values != nil {
		t.Errorf("metadata lookup initialized the scalar-bound slice: %#v", cfg.Values)
	}
}

func TestScalarSliceEmptyFileKeepsIndexedContainer(t *testing.T) {
	type config struct {
		Values []scalarSliceText `env:"VALUES,file,expand" envPrefix:"ITEM_"`
	}
	filename := filepath.Join(t.TempDir(), "empty values")
	if err := os.WriteFile(filename, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var cfg config
	rootHooks := 0
	err := ParseWithOptions(&cfg, Options{
		Environment:           map[string]string{"VALUES": "$SOURCE", "SOURCE": filename, "ITEM_0_VALUE": "indexed"},
		UseFieldNameByDefault: true,
		OnSet: func(key string, _ interface{}, _ bool) {
			if key == "VALUES" {
				rootHooks++
				if err := os.Remove(filename); err != nil {
					t.Error(err)
				}
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []scalarSliceText{{"indexed", 0}}
	if !reflect.DeepEqual(cfg.Values, want) {
		t.Errorf("values = %#v, want %#v", cfg.Values, want)
	}
	if rootHooks != 1 {
		t.Errorf("root hooks = %d, want one", rootHooks)
	}
}

func TestScalarSliceAbsentValueMetadataKeepsIndexedFields(t *testing.T) {
	type config struct {
		Values []scalarSliceText `setting:"VALUES" itemPrefix:"ITEM_"`
	}
	var cfg config
	params, err := GetFieldParamsWithOptions(&cfg, Options{
		TagName: "setting", PrefixTagName: "itemPrefix", Prefix: "APP_",
		Environment:           map[string]string{"APP_ITEM_0_VALUE": "indexed"},
		UseFieldNameByDefault: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, param := range params {
		keys = append(keys, param.Key)
	}
	if !reflect.DeepEqual(keys, []string{"APP_VALUES", "APP_ITEM_0_VALUE", "APP_ITEM_0_CALLS"}) {
		t.Errorf("keys = %v, want scalar key and indexed child keys", keys)
	}
}
