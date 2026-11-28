package env

import (
	"reflect"
	"testing"
)

func TestScalarSliceSkippedDefaultKeepsIndexedOverride(t *testing.T) {
	type config struct {
		Values []scalarSliceText `env:"VALUES" envPrefix:"ITEM_" envDefault:"default"`
	}
	cfg := config{Values: []scalarSliceText{{Value: "initial"}}}
	calls := 0
	err := ParseWithOptions(&cfg, Options{
		Environment:                  map[string]string{"ITEM_0_VALUE": "indexed"},
		SetDefaultsForZeroValuesOnly: true,
		OnSet:                        func(string, interface{}, bool) { calls++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []scalarSliceText{{Value: "indexed"}}
	if !reflect.DeepEqual(cfg.Values, want) {
		t.Errorf("values = %#v, want %#v", cfg.Values, want)
	}
	if calls != 2 {
		t.Errorf("callbacks = %d, want 2", calls)
	}
}

func TestScalarSliceUnresolvedMetadataKeepsIndexedFields(t *testing.T) {
	type expanded struct {
		Values []scalarSliceText `env:"VALUES,expand" envPrefix:"ITEM_"`
	}
	type file struct {
		Values []scalarSliceText `env:"VALUES,file" envPrefix:"ITEM_"`
	}
	type skippedDefault struct {
		Values []scalarSliceText `env:"VALUES" envPrefix:"ITEM_" envDefault:"default"`
	}
	tests := []struct {
		name     string
		config   any
		value    string
		zeroOnly bool
	}{
		{"empty-expansion", &expanded{}, "${MISSING}", false},
		{"missing-file", &file{}, "/file-that-must-not-be-read", false},
		{"skipped-default", &skippedDefault{Values: []scalarSliceText{{Value: "initial"}}}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callbacks := 0
			params, err := GetFieldParamsWithOptions(tt.config, Options{
				Environment:                  map[string]string{"VALUES": tt.value, "ITEM_0_VALUE": "indexed"},
				SetDefaultsForZeroValuesOnly: tt.zeroOnly,
				OnSet:                        func(string, interface{}, bool) { callbacks++ },
			})
			if err != nil {
				t.Fatal(err)
			}
			var keys []string
			for _, p := range params {
				keys = append(keys, p.Key)
			}
			if !reflect.DeepEqual(keys, []string{"VALUES", "ITEM_0_VALUE"}) {
				t.Errorf("keys = %v, want scalar and indexed key", keys)
			}
			if callbacks != 0 {
				t.Errorf("metadata fired %d callbacks", callbacks)
			}
		})
	}
}
