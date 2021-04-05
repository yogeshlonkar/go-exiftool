package exiftool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

var (
	ValidTagName   = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	ValidNamespace = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

// ExifSettool is the exiftool utility wrapper with arguments to set TAG
type ExifSettool struct {
	*Exiftool
	namespace string
	tagNames  []string
}

// NewExifSettool instantiates a new ExifSettool with namespace and custom tagNames that will be set
// namespace must be alphanumeric starting matching pattern ^[a-zA-Z0-9]+
// tagNames must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$
func NewExifSettool(namespace string, tagNames ...string) (*ExifSettool, error) {
	configFile := ""
	if namespace != "" {
		if !ValidNamespace.Match([]byte(namespace)) {
			return nil, errors.New("namespace must be alphanumeric starting matching pattern ^[a-zA-Z0-9]+$")
		}
		for _, tagName := range tagNames {
			if tagName == "" {
				return nil, errors.New("tagName must be provided to add custom metadata")
			}
			if !ValidTagName.Match([]byte(tagName)) {
				return nil, errors.New("tagName must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$")
			}
		}
		tempConfig, err := createTempConfig(namespace, tagNames...)
		if err != nil {
			return nil, err
		}
		configFile = tempConfig.Name()
	}
	e, err := NewExiftool(func(e *Exiftool) error {
		e.configFile = configFile
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ExifSettool{
		e,
		namespace,
		tagNames,
	}, nil
}

// SetUserDefinedMetadata sets user defined metadata on files. Will return error if tagName not configured with ExifSettool
func (e *ExifSettool) SetUserDefinedMetadata(tagName string, value interface{}, files ...string) ([]FileMetadata, error) {
	hasConfigure := false
	for _, configuredTagName := range e.tagNames {
		if configuredTagName == tagName {
			hasConfigure = true
			break
		}
	}
	if !hasConfigure {
		return nil, fmt.Errorf("tagName %s not configured while creating NewExifSettool", tagName)
	}
	return e.SetMetadata("xmp-"+e.namespace+":"+tagName, value, files...)
}

// SetMetadata sets metadata on files
func (e *ExifSettool) SetMetadata(name string, value interface{}, files ...string) (fms []FileMetadata, err error) {
	fms = make([]FileMetadata, len(files))
	for i, f := range files {
		fms[i].File = f
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				fms[i].Err = ErrNotExist
				continue
			}
			fms[i].Err = err
			continue
		}

		// set metadata
		fmt.Fprintln(e.stdin, fmt.Sprintf(`-%s=%v`, name, value))
		fmt.Fprintln(e.stdin, f)
		fmt.Fprintln(e.stdin, executeArg)

		if !e.scanMergedOut.Scan() {
			fms[i].Err = fmt.Errorf("nothing on stdMergedOut")
			continue
		}
		if e.scanMergedOut.Err() != nil {
			fms[i].Err = fmt.Errorf("error while writting stdMergedOut: %w", e.scanMergedOut.Err())
			continue
		}

		// read all metadata
		for _, curA := range extractArgs {
			fmt.Fprintln(e.stdin, curA)
		}
		fmt.Fprintln(e.stdin, f)
		fmt.Fprintln(e.stdin, executeArg)
		if !e.scanMergedOut.Scan() {
			fms[i].Err = fmt.Errorf("nothing on stdMergedOut")
			continue
		}
		if e.scanMergedOut.Err() != nil {
			fms[i].Err = fmt.Errorf("error while reading stdMergedOut: %w", e.scanMergedOut.Err())
			continue
		}
		var m []map[string]interface{}
		if err := json.Unmarshal(e.scanMergedOut.Bytes(), &m); err != nil {
			fms[i].Err = fmt.Errorf("error during unmarshaling (%v): %w)", string(e.scanMergedOut.Bytes()), err)
			continue
		}
		fms[i].Fields = m[0]
	}
	return fms, nil
}

func createTempConfig(namespace string, tagNames ...string) (file *os.File, err error) {
	config := `%%Image::ExifTool::UserDefined = (
    'Image::ExifTool::XMP::Main' => {
        %[1]s => {
            SubDirectory => {
                TagTable => 'Image::ExifTool::UserDefined::%[1]s',
            },
        },
    }
);
%%Image::ExifTool::UserDefined::%[1]s = (
    GROUPS => { 0 => 'XMP', 1 => 'XMP-%[1]s' },
    NAMESPACE => { '%[1]s' => 'http://ns.example.com/%[1]s/1.0/' },
    WRITABLE => 'string',
`
	file, err = ioutil.TempFile("", "exif.config")
	if err != nil {
		return nil, err
	}
	config = fmt.Sprintf(config, namespace)
	for _, tagName := range tagNames {
		config += fmt.Sprintf(`%[1]s => { },\n`, tagName)
	}
	config += `);\n`
	file.WriteString(config)
	file.Close()
	return file, nil
}
