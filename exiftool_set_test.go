package exiftool

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetCustomMetadata(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inFile string
		expOk  []bool
	}{
		{"single", "./testdata/20190404_131804.jpg", []bool{true}},
		{"empty", "./testdata/empty.jpg", []bool{true}},
	}

	for _, tc := range tcs {
		tc := tc // Pin variable
		t.Run(tc.tcID, func(t *testing.T) {
			testFile := tc.inFile + "_test"
			copyFile(tc.inFile, testFile)
			defer os.Remove(testFile)
			fms, err := SetCustomMetadata("OriginalFilename", "this_is_long", testFile)
			assert.Nilf(t, err, "error not nil: %v", err)
			if assert.Equal(t, len(tc.expOk), len(fms)) {
				for i, fm := range fms {
					t.Log(fm)
					assert.Equalf(t, tc.expOk[i], fm.Err == nil, "#%v different", i)
					assert.Equalf(t, "this_is_long", fm.Fields["OriginalFilename"], "custom metadata not set", i)
				}
			}
		})
	}
}

func TestSetMetadata(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inFile string
		expOk  []bool
	}{
		{"single", "./testdata/20190404_131804.jpg", []bool{true}},
		{"empty", "./testdata/empty.jpg", []bool{true}},
	}

	for _, tc := range tcs {
		tc := tc // Pin variable
		t.Run(tc.tcID, func(t *testing.T) {
			testFile := tc.inFile + "_test"
			copyFile(tc.inFile, testFile)
			defer os.Remove(testFile)
			fms, err := SetMetadata("abc", "Nameabc", "this is long", testFile)
			assert.Nilf(t, err, "error not nil: %v", err)
			if assert.Equal(t, len(tc.expOk), len(fms)) {
				for i, fm := range fms {
					t.Log(fm)
					os.Rename(fm.File+"_original", fm.File)
					assert.Equalf(t, tc.expOk[i], fm.Err == nil, "#%v different", i)
					assert.Equalf(t, "this is long", fm.Fields["Nameabc"], "custom metadata not set", i)
				}
			}
		})
	}
}

func TestSetMetadataErrorName(t *testing.T) {
	_, err := SetMetadata("abc", "Nam eabc", 1234, "./testdata/20190404_131804.jpg")
	assert.Equal(t, errors.New("name must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$"), err)
}

func TestSetMetadataErrorNamespace(t *testing.T) {
	_, err := SetMetadata("A-123", "Name", true, "./testdata/20190404_131804.jpg")
	assert.Equal(t, errors.New("namespace must be alphanumeric starting matching pattern ^[a-zA-Z0-9]+$"), err)
}

func copyFile(fromFile, toFile string) error {
	from, err := os.Open(fromFile)
	if err != nil {
		return err
	}
	defer from.Close()
	to, err := os.OpenFile(toFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()
	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}
