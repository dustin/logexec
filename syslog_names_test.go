package main

import (
	"testing"
)

func TestFacilityNames(t *testing.T) {
	for name := range facilityByName {
		var lf logFacility
		lf.Set(name)
		if lf.String() != name {
			t.Errorf("Error on %v, got %v", name, lf)
		}
	}
}

func TestLevelNames(t *testing.T) {
	for name := range levelByName {
		var lf logLevel
		lf.Set(name)
		if lf.String() != name {
			t.Errorf("Error on %v, got %v", name, lf)
		}
	}
}
