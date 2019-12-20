// Code generated for package asset by go-bindata DO NOT EDIT. (@generated)
// sources:
// schemas.yml
package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _schemasYml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\xc1\x8e\xe3\x36\x0c\xbd\xfb\x2b\x1e\x90\x02\x03\x14\xf6\x4c\xf7\x58\xdf\x0a\xf4\x52\xa0\xb7\x6d\xf7\xd0\xc5\x60\x23\xcb\x8c\xad\x46\x96\x14\x89\xce\xd6\xfd\xfa\x42\xb6\x93\x38\xb1\x13\xa7\xdd\x0a\x18\x20\xa2\x28\x8a\xef\xf1\x91\xe3\x0d\x7e\x0f\x84\x6d\x65\xb3\x42\x99\x52\xb0\x40\x66\xe1\xf6\xd5\x9b\x08\x81\xf8\x2d\xc8\x9a\x1a\xf1\x5a\x59\x64\x6e\x5f\xa1\x37\x62\x30\x86\xd7\xae\xd1\x5b\xec\xbc\x6d\xc0\x35\xc1\x93\xb3\xf0\xd6\x32\xd8\xa2\x68\x4d\xa9\x29\xd9\x80\x6b\x15\xd0\xc7\x55\x86\x2d\x04\x2a\x8b\x9d\xd2\x84\x60\xc1\xb5\x60\x28\x86\x14\x06\x05\x41\x19\xa9\xdb\x92\x4a\x28\x03\xfa\x8b\x64\xcb\xa2\xd0\x14\x5e\x93\x2c\xcb\x92\x5f\x4c\xa9\xa4\x60\xeb\x7f\xb6\xb2\x6d\xc8\x70\x9e\x00\xdc\x39\xca\x61\x8b\x3f\x49\x72\x02\x78\x3a\xb4\xca\x53\x19\x8f\x32\x08\xa7\x3e\x91\x0f\xca\x9a\x7e\xbb\x57\xa6\xec\x7f\x04\x47\x32\x01\x9c\xb7\x8e\x3c\x2b\x0a\xd1\x1d\x13\xf7\x61\x7f\x8a\x1e\xd8\x2b\x53\x8d\x26\x32\x6d\x93\xe3\xb3\x3a\x25\xe3\xbc\x65\x2b\xad\x7e\x55\xf6\xed\xf8\xe1\xbd\xf7\x8a\x0f\xad\x87\x98\xe1\x19\x2e\x37\xc4\x22\xb2\x75\x1d\xe0\x8c\x30\xae\x29\xca\xb8\x32\x18\xd1\xd0\xb8\xb9\x45\x15\x57\x3c\xbe\xec\x16\x93\x02\xb4\x28\x48\x87\xb9\xdb\xf0\x34\x36\xb1\x48\xc6\x32\xa4\x35\x2c\x94\xc1\x36\x30\xb9\x2d\xf6\xd4\xf5\x57\x22\xa9\xa7\xcb\xdf\x79\xda\xe5\x78\xd9\xbc\xcd\x30\x7e\x74\x24\x5f\xe6\xa5\xfc\x38\x5e\x7e\x58\x4e\xe7\x6d\xd9\xf6\xf6\x5b\x8c\xe3\xc9\x7f\xa5\x2c\xc3\xf1\x2c\x93\x6f\xa2\x10\x68\x94\xf9\x95\x4c\xc5\x75\x8e\x0f\x67\xf3\xf1\x5a\x55\xff\xe6\xfa\x59\x67\xe1\x1a\x9c\xf0\x5e\x74\xa3\x45\x31\x35\x93\x4c\xe7\xe4\x0f\xa4\x0f\x35\xee\x6c\xfb\x88\xa7\x25\xe8\xf6\xab\x21\xbf\x9a\x3c\x2b\xd6\xeb\x0c\x95\x14\xa4\x57\x8e\x9f\xa1\x23\x90\x8c\x7e\x0b\x92\x9c\xa2\x5f\x60\xe0\x0e\xb6\xfb\x08\xef\x20\x78\x58\xaa\x07\x68\x56\xef\xcd\xab\xba\x86\xf0\x2e\xce\xf9\x5b\xd8\xc0\x58\x7c\x15\x5d\x9c\xc2\x47\xa1\x55\x29\x98\xe2\x7c\x0e\x94\x5c\x49\x62\xb5\xe1\xc6\x16\xe9\x3b\xaf\x39\xe8\x85\xc6\x9b\x76\xc5\x02\x5e\x27\x98\xc9\x9b\x1c\x2f\x9f\x45\xf6\xf7\x4f\xd9\x1f\x5f\xf2\xf7\xf1\xd7\x0f\xd9\x8f\x5f\xf2\xf7\xef\x5f\x4e\xfd\xdb\x1c\xf4\x83\x40\xb3\xbe\xe8\x7d\x56\x47\xec\xde\xa9\x14\x41\xab\x14\x96\x6b\xf2\xc3\x80\xe5\xda\x53\xa8\xad\x2e\x9f\x6f\xa9\x45\x31\xdd\x8e\x94\x48\x94\xa6\x23\xe9\xc9\x3e\xb2\x15\xe9\x9e\x98\x8e\x42\xb7\x74\xde\x2f\xab\xb1\x0f\xb3\xa4\xe7\x99\x9a\x4e\x0f\x3c\xe5\x7c\xe6\x45\x73\x0a\xcd\x94\xa2\xe2\xf8\x47\x29\xe8\x90\xc2\xd0\xe1\x7d\xe2\xdd\x27\xba\x14\xd7\xb4\x4d\x41\x7e\x72\x20\x34\x79\xfe\xe6\xf6\xdb\xdd\xa2\xb8\x45\xd2\x0b\x9b\x21\x5b\xef\xc9\xb0\xee\xce\xe2\x2e\x21\x02\xca\xd6\x0b\xbe\xcc\xf0\xd3\x8a\xff\xa4\xfe\xb7\xb0\xce\x53\x20\xc3\x82\x67\xdf\x08\xab\x43\x54\xd6\xc2\xf3\x6f\x13\xd5\x3e\xa8\xd3\x58\xa5\x98\x7a\x8a\x42\xf8\x14\x81\x05\xb7\x21\xc5\xa1\xb5\x2c\x2e\x45\x1a\x73\xfe\x74\x5b\xa9\x21\x6e\x61\xad\x26\x71\x21\x64\x17\x15\x4b\x46\x76\x73\x57\x65\x98\xaa\x49\x4d\xef\x7d\x0b\x3c\x3d\x78\x6f\x30\xb5\x46\xf1\x42\xb8\x89\x57\x39\x7e\x09\xdc\xe7\x36\xf9\x27\x00\x00\xff\xff\xdc\x59\x1f\x35\xa9\x0a\x00\x00")

func schemasYmlBytes() ([]byte, error) {
	return bindataRead(
		_schemasYml,
		"schemas.yml",
	)
}

func schemasYml() (*asset, error) {
	bytes, err := schemasYmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schemas.yml", size: 2729, mode: os.FileMode(420), modTime: time.Unix(1576874780, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"schemas.yml": schemasYml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"schemas.yml": &bintree{schemasYml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
