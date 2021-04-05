package exiftool

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetMetadata(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inFile string
		expOk  bool
	}{
		{"single", "./testdata/20190404_131804.jpg", true},
	}
	e, err := NewExifSettool("")
	assert.Nil(t, err)
	defer e.Close()
	for _, tc := range tcs {
		tc := tc // Pin variable
		t.Run(tc.tcID, func(t *testing.T) {
			testFile := tc.inFile + "_test"
			copyFile(tc.inFile, testFile)
			defer os.Remove(testFile)
			fms, err := e.SetMetadata(true, "Make", "yogesh", testFile)
			assert.Nilf(t, err, "error not nil: %v", err)
			if assert.Equal(t, 1, len(fms)) {
				fm := e.ExtractMetadata(testFile)[0]
				t.Log(fm)
				assert.Equal(t, tc.expOk, fm.Err == nil)
				assert.Equalf(t, "yogesh", fm.Fields["Make"], "custom metadata not set")
			}
		})
	}
}

func TestSetUserDefinedMetadata(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inFile string
		expOk  bool
	}{
		{"single", "./testdata/20190404_131804.jpg", true},
		{"empty", "./testdata/empty.jpg", true},
		{"three", "./testdata/easter-4945288_640.jpg", true},
		{"four", "./testdata/eggs-3216877_640.jpg", true},
		{"five", "./testdata/willow-catkin-4949064_640.jpg", true},
	}
	e, err := NewExifSettool("custom", "OriginalFilename")
	assert.Nil(t, err)
	defer e.Close()
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.tcID, func(t *testing.T) {
			testFile := tc.inFile + "_test"
			copyFile(tc.inFile, testFile)
			defer os.Remove(testFile)
			fms, err := e.SetUserDefinedMetadata(true, "OriginalFilename", "this_is_long", testFile)
			assert.Nilf(t, err, "error not nil: %v", err)
			if assert.Equal(t, 1, len(fms)) {
				fm := e.ExtractMetadata(testFile)[0]
				t.Log(fm)
				assert.Equal(t, tc.expOk, fm.Err == nil, "err not nil", err)
				assert.Equal(t, "this_is_long", fm.Fields["OriginalFilename"], "custom metadata not set file")
			}
		})
	}
}

func TestSetUserDefinedMetadataErrorTagNotConfigured(t *testing.T) {
	e, err := NewExifSettool("abc", "X1")
	defer e.Close()
	_, err = e.SetUserDefinedMetadata(true, "OriginalFilename", "this_is_long", "./testdata/empty.jpg")
	assert.Equal(t, errors.New("tagName OriginalFilename not configured while creating NewExifSettool"), err)

}

func TestNewExifSettoolErrorTagName(t *testing.T) {
	e, err := NewExifSettool("abc", "Nam eabc")
	assert.Nil(t, e)
	assert.Equal(t, errors.New("tagName must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$"), err)
}

func TestNewExifSettoolErrorTErrorNamespace(t *testing.T) {
	e, err := NewExifSettool("A-123", "Name")
	assert.Nil(t, e)
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
