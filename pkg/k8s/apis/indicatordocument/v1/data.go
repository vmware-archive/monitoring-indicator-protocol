// Code generated for package v1 by go-bindata DO NOT EDIT. (@generated)
// sources:
// schemas.yml
package v1

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

var _schemasYml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x55\xdf\x6b\xdb\x40\x0c\x7e\xcf\x5f\x21\xe8\x20\x30\x9c\x75\x7d\x9c\xdf\x06\x7b\x19\xec\x6d\xa3\x0f\x2b\xa1\x55\xce\x6a\x72\xeb\xf9\xee\xa2\x93\x53\xbc\xbf\x7e\x9c\xed\xb8\x76\xfc\x2b\x85\x19\x02\x39\x59\x92\xf5\x7d\xfa\xa4\xfb\x6e\x33\xad\x50\x1c\x7f\x73\xaa\xc8\xc9\x4a\xba\x02\x90\xd2\x53\x0a\x6e\xf7\x87\x94\xac\x00\x98\x8e\x85\x66\xca\xe2\xab\x0d\xa0\xd7\xf7\xc4\x41\x3b\x5b\x1d\x5f\xb4\xcd\xaa\x3f\xc1\x93\x5a\x01\x78\x76\x9e\x58\x34\x85\xe8\x0e\x1d\xf7\xfa\x7c\xce\x1e\x84\xb5\xdd\x37\x26\xb2\x45\x9e\xc2\x83\x3e\x17\xe3\xd9\x89\x53\xce\x7c\xd2\xee\xf6\x74\xb7\xad\xbc\xe2\x87\x96\x53\x0c\xf0\xd4\xc1\x39\x09\x66\x28\xd8\x4f\xd0\x22\x8c\xcf\x65\xe1\xf1\x31\xb8\x23\xd3\x39\xf7\x03\xe1\x06\x14\x5a\xeb\x04\x94\xb3\x82\xda\xc2\x53\x10\xf2\x4f\xf0\x42\x65\x15\x12\x29\x39\x07\x7f\x60\x7a\x4e\x61\x7d\x73\x3b\xa8\xf0\xa7\x27\xb5\x5e\x8d\x9a\x17\x9b\xe1\xd9\x65\x45\x65\xbf\x2c\xbf\x79\x33\x03\xb8\x9b\x29\x3e\x1b\xb0\x98\x53\x7b\x38\xb5\x4d\x9e\x62\x27\xba\x0f\xb9\xe9\x75\xa5\xe2\x5e\xdb\x1f\x64\xf7\x72\x48\xe1\xae\x35\x9f\xfa\x9a\x78\x4f\x78\xab\x92\xd0\x07\x87\xcc\x58\x36\x16\x2d\x94\x77\x2a\x1d\x92\x5f\x93\x5e\xf7\xb8\x74\xc5\x1c\x4f\x63\xd0\xdd\xab\x25\x5e\x2c\x5e\xb4\x98\x65\x86\x32\x0a\x8a\xb5\x97\x6b\xe8\x08\xa4\xa2\xdf\x88\x24\xbb\xe8\x47\x18\x98\xc0\x36\x8d\x70\x02\xc1\x6c\xab\x66\xd0\x2c\xc6\x0d\xbb\xba\x84\x70\x12\xe7\xf0\x5b\x70\x03\xd6\xc1\x2b\x96\x20\x0e\x4e\x68\x74\x86\x42\x20\x07\x0a\xb4\xea\x49\x62\x71\xe0\x9a\x11\xa9\x26\x2f\x3f\x9a\x91\xc1\xeb\x4e\xc5\x08\x5e\x8f\x22\xc4\x36\x85\xf5\x03\x6e\xfe\x7e\xdd\xfc\x7e\x4c\xb7\xcd\xbf\xcf\x9b\x2f\x8f\xe9\xf6\xe3\xfa\x3c\xbf\xf9\xd1\xcc\x24\x1a\xcc\x45\xe5\xb3\xb8\x20\x5f\xbc\x4e\x20\x18\x9d\x80\x93\x03\x71\xbd\x1e\xe5\xc0\x14\x0e\xce\x64\xd7\x8f\xd4\xa8\x98\x2e\x57\x4a\x24\xca\xd0\x89\x4c\xe7\x1c\xd9\x8a\x74\x77\x4c\x27\x34\x05\xb5\xe7\x71\x35\x56\x69\xc6\xf4\x3c\x50\xd3\xf9\x03\x57\x39\xb7\xbc\x18\x49\xc0\x08\x25\xb0\x97\xf8\xa3\x04\xe8\x98\x80\xa5\xe3\xb6\xe3\x5d\x15\x3a\x96\xd7\x16\xf9\x8e\x6a\x48\x68\x88\xdf\xbb\x4e\x9e\xdd\xe4\x32\xa9\xa4\x2b\xa0\x0a\x66\xb2\x62\xca\x56\xbe\x19\x60\x80\xac\x60\x94\xb7\x2d\x0d\x10\x2f\xa0\xff\x90\xca\x33\x05\xb2\x82\x32\xb8\xb7\x17\xb1\xa8\x03\xb2\xfc\xea\x68\x71\x86\xfd\x86\xfb\x58\x74\x02\x3b\xe4\x04\x82\xa0\x14\x21\x81\x63\xe1\x04\xdf\xa8\x6f\x6a\xbe\xbf\xe4\xbf\xce\xbb\x73\xce\x10\xbe\x91\xf0\x1c\x75\x48\x56\x95\x43\x57\x6d\x85\xf6\xc4\x8b\x37\xfc\xd5\xeb\xb4\x83\x29\x6b\x6e\xee\x69\xd6\x56\xff\x02\x00\x00\xff\xff\x79\x55\x63\x6a\x6f\x09\x00\x00")

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

	info := bindataFileInfo{name: "schemas.yml", size: 2415, mode: os.FileMode(420), modTime: time.Unix(1568233996, 0)}
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
