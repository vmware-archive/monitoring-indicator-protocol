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

var _schemasYml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\x5f\x8f\xa3\x38\x0c\x7f\xe7\x53\x58\xea\x49\x95\x4e\xd0\xde\x3c\x1e\x6f\x27\xdd\xcb\x49\xf7\x36\xbb\xf3\xb0\xa3\x6a\x1a\x82\x0b\xd9\x86\x38\x4d\x4c\x67\xd9\x4f\xbf\x0a\xa5\x0c\x2d\xb4\x74\xb5\x1b\xa9\x52\x6d\x6c\xc7\xbf\x9f\xff\xc0\x02\x3e\x7b\x84\x6d\x41\x49\xa6\x4c\x2e\x58\x40\x42\x60\xf7\xc5\x5a\x78\x8f\xbc\xf6\xb2\xc4\x4a\xac\x0a\x82\xc4\xee\x0b\x68\x95\x70\x52\xfa\x55\x53\xe9\x2d\xec\x1c\x55\xc0\x25\x82\x43\x4b\xe0\x88\x18\x98\x20\xab\x4d\xae\x31\x5a\x00\x97\xca\x43\x1b\x57\x19\x26\x10\x50\x10\xec\x94\x46\xf0\x04\x5c\x0a\x06\xc5\x20\x85\x81\x0c\x41\x19\xa9\xeb\x1c\x73\x50\x06\xf0\x1b\xca\x9a\x45\xa6\xd1\xaf\xa2\x24\x49\xa2\xff\x4c\xae\xa4\x60\x72\xff\x92\xac\x2b\x34\x9c\x46\x00\xdc\x58\x4c\x81\xb2\xaf\x28\x39\x02\x70\x78\xa8\x95\xc3\x3c\x3c\x4a\x40\x58\xf5\x82\xce\x2b\x32\xad\xb8\x57\x26\x6f\xff\x78\x8b\x32\x02\xb0\x8e\x2c\x3a\x56\xe8\x83\x39\x0c\xcc\x4f\xf2\x39\xba\x67\xa7\x4c\xd1\xa9\xd0\xd4\x55\x0a\xaf\xea\x9c\x8c\x75\xc4\x24\x49\xaf\x14\xad\x8f\x4f\x9b\xd6\x2a\x5c\x34\x1f\x62\x84\xe7\xe4\x5c\x21\x8b\xc0\xd6\x65\x80\x1e\x61\x38\xd7\x89\x87\xa3\x45\x86\x7a\x20\x5f\x3a\xc2\x22\x50\x6c\x88\x41\x92\x61\xa1\x0c\x6c\x3d\xa3\xdd\xc2\x1e\x9b\xd6\x25\x50\x72\x76\xfe\xc3\xe1\x2e\x85\xe5\x62\x3d\xca\xf0\xd9\xa2\x5c\x8e\x0b\xf1\xdc\x39\xdf\x2d\x86\x75\x94\xd7\xad\xfe\x3a\xfd\xee\xc9\x1d\xc0\xc3\x48\xe1\x24\x60\x44\x85\xbd\x70\xec\x8b\x7c\x8b\x9d\x60\x3e\xe6\xe6\xa2\x2a\x2d\xf7\xca\xfc\x8f\xa6\xe0\x32\x85\xa7\x5e\x7d\xbc\xec\x89\x9f\x71\xef\xbb\xc4\x5f\x82\x13\xce\x89\xa6\xd3\x28\xc6\x6a\x90\xe9\x98\xfc\x13\xe9\xa7\x1a\x37\x54\xdf\xe3\x69\x0a\x3a\xbd\x1b\x74\xb3\xc9\xb3\x62\x3d\xcf\x50\x8e\x5e\x3a\x65\xf9\x11\x3a\x3c\xca\x60\x37\xd1\x92\x43\xf4\x13\x0c\xdc\xc0\x76\x1b\xe1\x0d\x04\x77\x4b\x75\x07\xcd\xac\xdf\xb8\xaa\x73\x08\x6f\xe2\x1c\xdf\x05\x0b\x30\x04\xef\xa2\x09\x3b\xf4\x28\xb4\xca\x05\x63\xd8\xae\x1e\xa3\x8b\x96\x98\x1d\xb8\x6e\x44\xda\xc9\xab\x0e\x7a\x62\xf0\x86\x53\x31\x81\xd7\x0a\x66\x74\x26\x85\xe5\xab\x48\xbe\xff\x93\x7c\x79\x4b\x37\xdd\xbf\xbf\x92\xbf\xdf\xd2\xcd\x9f\xcb\xf3\xfc\x56\x07\x7d\x27\xd0\x68\x2e\x5a\x9b\xd9\x05\xb9\xb7\x2a\x06\xaf\x55\x0c\xc4\x25\xba\xd3\x7a\xe4\xd2\xa1\x2f\x49\xe7\x8f\x8f\xd4\x64\x33\x5d\xaf\x94\x40\x94\xc6\x23\xea\x81\x1c\xd8\x0a\x74\x0f\x54\x47\xa1\x6b\xec\xe5\xe9\x6e\x6c\xc3\x4c\xf5\xf3\xa8\x9b\xce\x17\x3c\x64\xdc\xf3\xa2\x39\x06\xcd\x18\x43\xc1\xe1\x87\x31\xe0\x21\x06\x83\x87\xcd\xc0\xba\x4d\x74\x2a\xae\xa9\xab\x0c\xdd\xe0\x81\xd0\xe8\xf8\x97\xc7\x6f\x77\x8d\xe2\x1a\x49\xdb\xd8\x0c\xb2\x76\x0e\x0d\xeb\xa6\x6f\xee\x1c\x84\x87\xbc\x76\x82\x3f\x76\xf8\xf9\x84\x97\xd4\x6f\x0b\x6b\x1d\x7a\x34\x2c\x78\xf4\x86\x9f\x5d\xa2\xb2\x14\x8e\x3f\x0d\xba\xf6\x4e\x9d\xba\x2a\x85\xd4\x63\xc8\x84\x8b\xc1\xb3\xe0\xda\xc7\x70\xa8\x89\xc5\x47\x91\xba\x9c\x5f\xae\x2b\x75\x8a\x9b\x11\x69\x14\x1f\x84\xec\x42\xc7\xa2\x91\xcd\xd8\x54\x19\xc6\x62\x50\xd3\x5b\xdf\x02\x0f\x2f\xde\x2b\x4c\xb5\x51\x3c\x11\x6e\x60\x95\x77\x5f\x02\xb7\xb9\x8d\x7e\x04\x00\x00\xff\xff\x32\x90\x2c\x19\x67\x0a\x00\x00")

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

	info := bindataFileInfo{name: "schemas.yml", size: 2663, mode: os.FileMode(420), modTime: time.Unix(1576692547, 0)}
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
