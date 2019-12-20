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

var _schemasYml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\x51\xaf\xab\x36\x0c\x7e\xe7\x57\x58\xea\xa4\x4a\x13\xb4\xbb\x8f\xe3\x6d\xd2\x5e\x26\xed\xed\x6e\xf7\x61\x47\xd5\x69\x08\x2e\x64\x0d\x71\x9a\x98\x9e\xb1\x5f\x3f\x05\x28\xa5\x85\x96\x4e\xe7\x46\xaa\x44\x8c\xe3\xf8\xfb\xfc\xd9\x65\x05\x7f\x7a\x84\x7d\x41\x49\xa6\x4c\x2e\x58\x40\x42\x60\x8f\xc5\x56\x78\x8f\xbc\xf5\xb2\xc4\x4a\x6c\x0a\x82\xc4\x1e\x0b\x68\x8d\xd0\x19\xfd\xa6\xa9\xf4\x1e\x0e\x8e\x2a\xe0\x12\xc1\xa1\x25\x70\x44\x0c\x4c\x90\xd5\x26\xd7\x18\xad\x80\x4b\xe5\xa1\x8d\xab\x0c\x13\x08\x28\x08\x0e\x4a\x23\x78\x02\x2e\x05\x83\x62\x90\xc2\x40\x86\xa0\x8c\xd4\x75\x8e\x39\x28\x03\xf8\x0f\xca\x9a\x45\xa6\xd1\x6f\xa2\x24\x49\xa2\xdf\x4c\xae\xa4\x60\x72\xbf\x92\xac\x2b\x34\x9c\x46\x00\xdc\x58\x4c\x81\xb2\xbf\x51\x72\x04\xe0\xf0\x54\x2b\x87\x79\x78\x95\x80\xb0\xea\x1b\x3a\xaf\xc8\xb4\xdb\xa3\x32\x79\xfb\xe0\x2d\xca\x08\xc0\x3a\xb2\xe8\x58\xa1\x0f\xee\x30\x72\xef\xf6\x97\xe8\x9e\x9d\x32\x45\x6f\x42\x53\x57\x29\xbc\xa9\x4b\x32\xd6\x11\x93\x24\xbd\x51\xb4\x3d\x7f\xd9\xb5\x5e\xe1\xa2\xe5\x10\x13\x3c\xdd\xe1\x0a\x59\x04\xb6\x6e\x03\x0c\x08\xc3\xba\x4f\x3c\x2c\x23\x2a\xbc\xee\x66\xef\x05\xd0\x22\x43\xed\xa7\x6e\x5d\x74\x58\x85\x3a\x18\x62\x90\x64\x58\x28\x03\x7b\xcf\x68\xf7\x70\xc4\xa6\x3d\x12\x78\xbb\x1c\xfe\xc1\xe1\x21\x85\xf5\x6a\x3b\x81\xf1\xd5\xa2\x5c\x4f\xab\xf5\xb5\x3f\xfc\xb4\x62\xd6\x51\x5e\xb7\xf6\x7b\x8c\xfd\x9b\x27\xac\x8c\x23\x85\x95\xb4\x9c\x0c\x9b\xf3\xa0\x84\x4f\x51\x08\x50\x29\xf3\x3b\x9a\x82\xcb\x14\xbe\x0c\xe6\xf3\xad\x70\xfe\xcf\xf1\x41\x4a\xfe\x16\x9c\x70\x4e\x34\xbd\x45\x31\x56\xa3\x4c\xa7\xe4\x77\xa4\x77\x35\x6e\xa8\x7e\xc6\xd3\x1c\x74\xfa\x30\xe8\x16\x93\x67\xc5\x7a\x99\xa1\x1c\xbd\x74\xca\xf2\x2b\x74\x78\x94\xc1\x6f\x46\x92\x63\xf4\x33\x0c\x3c\xc0\xf6\x18\xe1\x03\x04\x4f\x4b\xf5\x04\xcd\xe2\xb9\x69\x55\x97\x10\x3e\xc4\x39\xbd\x0b\x56\x60\x08\x3e\x44\x13\x06\xed\x59\x68\x95\x0b\xc6\x30\x82\x3d\x46\x37\x92\x58\x6c\xb8\xbe\x45\xda\xce\xab\x4e\x7a\xa6\xf1\xc6\x5d\x31\x83\xd7\x0a\x66\x74\x26\x85\xf5\x9b\x48\xfe\xfd\x25\xf9\xeb\x3d\xdd\xf5\x4f\x3f\x25\x3f\xbf\xa7\xbb\x1f\xd7\x97\xfe\xad\x4e\xfa\x49\xa0\x49\x5f\xb4\x3e\x8b\x53\xf4\x68\x55\x0c\x5e\xab\x18\x88\x4b\x74\xdd\x0c\xe5\xd2\xa1\x2f\x49\xe7\xaf\xb7\xd4\xac\x98\xee\x47\x4a\x20\x4a\xe3\x19\xf5\x68\x1f\xd8\x0a\x74\x8f\x4c\x67\xa1\x6b\x1c\xf6\xf3\x6a\x6c\xc3\xcc\xe9\x79\xa2\xa6\xcb\x05\x2f\x39\x0f\xbc\x68\x8e\x41\x33\xc6\x50\x70\xf8\x61\x0c\x78\x8a\xc1\xe0\x69\x37\xf2\x6e\x13\x9d\x8b\x6b\xea\x2a\x43\x37\x7a\x21\x34\x3a\xfe\x74\xfb\x1d\xee\x51\xdc\x23\x69\x85\xcd\x20\x6b\xe7\xd0\xb0\x6e\x06\x71\xe7\x20\x3c\xe4\xb5\x13\x7c\x9d\xe1\x97\x15\xfe\xa4\xbe\x5b\x58\xeb\xd0\xa3\x61\xc1\x93\xcf\x80\xc5\x21\x2a\x4b\xe1\xf8\x8f\x91\x6a\x9f\xd4\xa9\xaf\x52\x48\x3d\x86\x4c\xb8\x18\x3c\x0b\xae\x7d\x0c\xa7\x9a\x58\x5c\x8b\xd4\xe7\xfc\xed\xbe\x52\x5d\xdc\x8c\x48\xa3\xb8\x12\x72\x08\x8a\x45\x23\x9b\xa9\xab\x32\x8c\xc5\xa8\xa6\x8f\xbe\x05\x5e\x1e\xbc\x77\x98\x6a\xa3\x78\x26\xdc\xc8\x2b\xef\xbf\x04\x1e\x73\x1b\xfd\x17\x00\x00\xff\xff\xb5\xce\x5b\xea\x8c\x0a\x00\x00")

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

	info := bindataFileInfo{name: "schemas.yml", size: 2700, mode: os.FileMode(420), modTime: time.Unix(1576881713, 0)}
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
