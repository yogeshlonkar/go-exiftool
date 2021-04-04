package exiftool

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomMetadata(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inFile string
		expOk  []bool
	}{
		{"single", "./testdata/20190404_131804.jpg", []bool{true}},
	}

	for _, tc := range tcs {
		tc := tc // Pin variable
		t.Run(tc.tcID, func(t *testing.T) {
			testFile := tc.inFile + "_test"
			copyFile(tc.inFile, testFile)
			defer os.Remove(testFile)
			fms, err := SetMetadata("", "Make", "yogesh", testFile)
			assert.Nilf(t, err, "error not nil: %v", err)
			if assert.Equal(t, len(tc.expOk), len(fms)) {
				for i, fm := range fms {
					t.Log(fm)
					assert.Equalf(t, tc.expOk[i], fm.Err == nil, "#%v different", i)
					assert.Equalf(t, "yogesh", fm.Fields["Make"], "custom metadata not set", i)
				}
			}
		})
	}
}
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
			fms, err := SetCustomMetadata("custom", "OriginalFilename", "this_is_long", testFile)
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

func TestSetCustomMetadataErrorName(t *testing.T) {
	_, err := SetCustomMetadata("abc", "Nam eabc", 1234, "./testdata/20190404_131804.jpg")
	assert.Equal(t, errors.New("name must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$"), err)
}

func TestSetCustomMetadataErrorNamespace(t *testing.T) {
	_, err := SetCustomMetadata("A-123", "Name", true, "./testdata/20190404_131804.jpg")
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
