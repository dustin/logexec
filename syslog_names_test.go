package main

import (
	"testing"
)

func TestFacilityNames(t *testing.T) {
	for name := range facilityByName {
		lf := new(logFacility)
		lf.Set(name)
		if lf.String() != name {
			t.Errorf("Error on %v, got %v", name, lf)
		}
	}
}

func TestLevelNames(t *testing.T) {
	for name := range levelByName {
		lf := new(logLevel)
		lf.Set(name)
		if lf.String() != name {
			t.Errorf("Error on %v, got %v", name, lf)
		}
	}
}
