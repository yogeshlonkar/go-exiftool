package exiftool

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
)

var (
	ValidTagName   = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	ValidNamespace = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

func createTempConfig(namespace, name string) (file *os.File, err error) {
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
    %[2]s => { },
);
`
	file, err = ioutil.TempFile("", "exif.config")
	if err != nil {
		return nil, err
	}
	file.WriteString(fmt.Sprintf(config, namespace, name))
	file.Close()

	return file, nil
}

// SetMetadata set metadata on files. namespace must be alphanumeric starting matching pattern ^[a-zA-Z0-9]+ and name must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$
func SetCustomMetadata(name string, value interface{}, files ...string) ([]FileMetadata, error) {
	return SetMetadata("custom", name, value, files...)
}

// SetMetadata set metadata on files. namespace must be alphanumeric starting matching pattern ^[a-zA-Z0-9]+ and name must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$
func SetMetadata(namespace, name string, value interface{}, files ...string) ([]FileMetadata, error) {
	if namespace == "" {
		return nil, errors.New("namespace must be provided to add custom metadata")
	}
	if !ValidNamespace.Match([]byte(namespace)) {
		return nil, errors.New("namespace must be alphanumeric starting matching pattern ^[a-zA-Z0-9]+$")
	}
	if name == "" {
		return nil, errors.New("name must be provided to add custom metadata")
	}
	if !ValidTagName.Match([]byte(name)) {
		return nil, errors.New("name must be alphanumeric starting with capital letter matching pattern ^[A-Z][a-zA-Z0-9]*$")
	}
	tempConfig, err := createTempConfig(namespace, name)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempConfig.Name())

	e := Exiftool{}
	defer e.Close()
	args := []string{"-config", tempConfig.Name(), "-overwrite_original"}
	args = append(args, initArgs...)
	cmd := exec.Command(binary, args...)
	r, w := io.Pipe()
	e.stdMergedOut = r
	cmd.Stdout = w
	cmd.Stderr = w
	if e.stdin, err = cmd.StdinPipe(); err != nil {
		return nil, fmt.Errorf("error when piping stdin: %w", err)
	}
	e.scanMergedOut = bufio.NewScanner(r)
	if e.bufferSet {
		e.scanMergedOut.Buffer(e.buffer, e.bufferMaxSize)
	}
	e.scanMergedOut.Split(splitReadyToken)
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("error when executing commande: %w", err)
	}
	e.lock.Lock()
	defer e.lock.Unlock()

	fms := make([]FileMetadata, len(files))

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

		fmt.Fprintln(e.stdin, fmt.Sprintf(`-xmp-%s:%s=%v`, namespace, name, value))
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
