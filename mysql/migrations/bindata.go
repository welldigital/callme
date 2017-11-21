// Code generated by go-bindata.
// sources:
// 20171115153101_create_initial.down.sql
// 20171115153101_create_initial.up.sql
// bindata.go
// DO NOT EDIT!

package migrations

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

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var __20171115153101_create_initialDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\xc8\xca\x4f\x4a\xb0\xe6\x02\x04\x00\x00\xff\xff\xe2\xbe\x7c\x92\x12\x00\x00\x00")

func _20171115153101_create_initialDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__20171115153101_create_initialDownSql,
		"20171115153101_create_initial.down.sql",
	)
}

func _20171115153101_create_initialDownSql() (*asset, error) {
	bytes, err := _20171115153101_create_initialDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "20171115153101_create_initial.down.sql", size: 18, mode: os.FileMode(420), modTime: time.Unix(1510931792, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __20171115153101_create_initialUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x54\x4d\x6f\xdb\x30\x0c\xbd\xe7\x57\xf0\x98\x00\x39\x0c\xc3\x36\x0c\xe8\xc9\x71\xd4\xce\x98\x2d\x0f\xaa\x32\xb4\x27\x7f\x89\x45\x9d\xba\x56\x20\x3b\x5d\xf2\xef\x87\x28\xfe\x8a\x23\xab\x18\xd6\x5b\xc2\x47\xf3\xf1\x3d\x52\x5c\x91\x3b\x8f\xde\xcc\x66\x2e\x23\x0e\x27\xc0\x9d\x95\x4f\x20\xde\xca\x34\x86\xf9\x0c\x20\xce\x85\xfe\xed\x51\x0e\x34\xe4\x40\x37\xbe\x0f\xce\x86\x87\x91\x47\x5d\x46\x02\x42\xf9\xf2\x9c\x56\x65\xcf\x28\xf6\x05\x36\xb9\x1b\xdf\xd7\xc0\x9f\x67\x2c\x63\x58\x3b\x9c\x70\x2f\x20\xf3\x6f\x8b\xae\x8c\x86\x13\x55\xc6\xf0\xdb\x61\xee\x0f\x87\xcd\x3f\x7f\xfa\xf2\x7d\xd1\x63\xbb\xe4\x58\xc8\x44\xc4\x10\x90\xb5\xb7\x09\x38\x79\xe8\x3f\xfc\xc5\xbc\xc0\x61\x8f\xf0\x93\x3c\xc2\xbc\x69\x72\xb1\x30\xe9\x50\x58\xed\x64\x59\xe1\x50\x4f\x1f\x7b\x5f\x57\x81\xc9\x38\x71\xd9\x15\xca\x85\x01\xa9\xf3\x57\xb4\x48\xee\xc9\x87\xba\x2e\x4a\x57\xa8\x94\x54\x31\xac\xbc\x11\xd2\xc4\x27\x3e\x34\x98\xd2\x91\x19\xcc\xe9\x27\x36\x37\x8e\xd0\x62\x0b\x1e\x6a\x54\x65\x52\x9c\xf4\x77\xd3\xfb\x3a\x56\x9a\x1e\x6d\xa8\x1e\xfd\xb5\x78\xe3\xd4\x87\x09\x99\xc2\xa4\x46\x61\x71\xf8\x49\xc9\x57\xdb\xce\x65\x75\xfe\x86\x06\x77\x05\x6a\xe8\x54\x5d\x24\xf5\xe5\x0c\x4d\xfe\x76\x7e\x19\xcc\xcd\x94\x2c\xeb\xa4\x7f\x45\xdd\xff\x7f\x7e\x49\x97\xda\x9b\x2a\xd3\xb6\xee\x14\xbe\xe5\x72\x5f\x59\x0c\x28\xf1\x50\x5b\xe0\x22\xa9\xea\xfd\x4e\xd8\x4d\x1e\x7b\xd1\x76\x76\x6d\x85\x7e\x40\xda\x86\xe6\x31\xbd\x6b\x41\x7d\xdc\xa1\x45\x62\x21\xb3\x17\x14\xe9\xd1\xb6\x5b\x36\x7d\xfb\xb2\xce\x8b\x69\x5c\x61\x95\xe5\xa5\x40\x71\xb5\x20\x63\xcd\xe7\xd3\x30\x54\xec\xd1\x35\x79\x80\x5c\x1c\x22\x8d\x45\x27\x25\x91\xe6\x83\x90\xb6\x4e\xc4\x1d\x43\xbc\x6c\xc4\x2e\xdb\xae\x4e\xb5\x1c\x9f\x13\x36\xbc\xc4\x33\x00\x00\x67\xbd\x06\x37\xa4\xf7\x9c\x39\x27\xff\x9e\x5e\xa2\xad\x4c\xa3\x7e\x5b\x40\x67\xdd\x86\x8c\x78\x77\xf4\xdc\x62\x0f\x2e\x80\x91\x5b\xc2\x08\x75\xc9\x3d\xb4\xc1\x21\x3e\xa2\x1d\x5c\x8e\x69\xf2\x36\x23\x6a\xe7\x6a\xea\x40\x23\x17\xf4\x3a\xd2\x21\xff\x49\xbc\x95\xa9\x89\x55\x9f\xe6\x0b\xd6\xad\x4c\xcf\xf1\x31\x63\xb3\xb8\x30\x41\xd7\xc0\x1f\x69\xb4\x1b\x06\x81\xc7\x6f\xfe\x06\x00\x00\xff\xff\x38\xe4\x48\x77\x7a\x07\x00\x00")

func _20171115153101_create_initialUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__20171115153101_create_initialUpSql,
		"20171115153101_create_initial.up.sql",
	)
}

func _20171115153101_create_initialUpSql() (*asset, error) {
	bytes, err := _20171115153101_create_initialUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "20171115153101_create_initial.up.sql", size: 1914, mode: os.FileMode(420), modTime: time.Unix(1511248925, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _bindataGo = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\xdd\x8e\xdb\x46\xd2\xbd\x26\x9f\xa2\x23\x20\x81\xf4\x41\x9f\x86\xcd\x7f\x0a\x98\x9b\xd8\x5e\xc0\x17\xeb\x00\x1b\xe7\x6a\x6b\x31\x68\x92\xdd\x5a\x62\x25\x51\xa6\xa8\xb8\x66\x8c\x79\xf7\xc5\xe9\x6a\x8d\xc7\xd9\xc4\x9e\x2c\xf6\x42\x12\x7f\xba\xab\x4e\xd5\x39\xa7\x5a\x37\x37\xea\xd5\xd8\x5b\xb5\xb3\x47\x3b\x99\xd9\xf6\xaa\xbd\x57\xbb\xf1\xff\xdb\xe1\xd8\x9b\xd9\x6c\xe2\x9b\x1b\x75\x1e\x2f\x53\x67\xcf\x5b\x5c\xa7\x89\xae\xb4\xd6\x85\x2e\x32\x9d\xe8\xbb\x6e\xb2\x66\xb6\x77\xc3\x71\x98\x07\xb3\xdf\xf4\xe3\xc7\xe3\xe6\xfc\x61\xff\xed\x95\x97\xd3\x75\xdd\x35\xd3\x6e\xc4\xdd\xeb\x9f\xd4\xbb\x9f\xde\xab\x37\xaf\xdf\xbe\xff\x2e\x8e\x4f\xa6\xfb\x97\xd9\x59\x75\x18\x76\x93\x99\x87\xf1\x78\x8e\xe3\xe1\x70\x1a\xa7\x59\x2d\xe3\x68\xd1\xde\xcf\xf6\xbc\x88\xa3\x45\x37\x1e\x4e\x93\x3d\x9f\x6f\x76\x0f\xc3\x09\x0f\xdc\x61\xc6\xcf\x30\xca\xf7\xcd\x30\x5e\xe6\x61\x8f\x9b\xd1\x6f\x38\x99\xf9\x9f\x37\x6e\xd8\x5b\x5c\xe0\xc1\x79\x9e\x86\xe3\xce\xbf\x9b\x87\x83\x5d\xc4\xab\x38\x76\x97\x63\x77\x85\xf7\x37\x6b\xfa\x25\x2e\xd4\xdf\xff\x81\xb4\x6b\x75\x34\x07\xab\x64\xdb\x4a\x2d\xaf\x4f\xed\x34\x8d\xd3\x4a\x7d\x8a\xa3\xdd\x83\xbf\x53\xdb\x5b\x05\x54\x9b\x77\xf6\x23\x82\xd8\x69\xe9\x61\xe3\xfe\xc7\x8b\x73\x76\xf2\x61\x57\xab\x38\x1a\x9c\xdf\xf0\xdd\xad\x3a\x0e\x7b\x84\x88\x26\x3b\x5f\xa6\x23\x6e\xd7\xca\x1d\xe6\xcd\x1b\x44\x77\xcb\x05\x02\xa9\xef\x3f\x6c\xd5\xf7\xbf\x2e\x04\x89\xcf\xb5\x8a\xa3\xc7\x38\x8e\x7e\x35\x93\x6a\x2f\x4e\x49\x1e\x49\x12\x47\x77\x02\xe7\x56\x0d\xe3\xe6\xd5\x78\xba\x5f\xfe\xd0\x5e\xdc\x5a\xed\x1e\x56\x71\xd4\xed\xdf\x5c\x91\x6e\x5e\xed\xc7\xb3\x5d\xae\xe2\xff\x15\x1e\x84\x91\xf8\x7f\x10\xc8\x4e\x93\xe0\x0e\x0f\xdb\x8b\xdb\xfc\x08\xe8\xcb\xd5\x1a\x2b\xe2\xc7\x38\x9e\xef\x4f\x56\x99\xf3\xd9\xce\x68\xf9\xa5\x9b\x11\xc5\xd7\x17\xf8\x88\xa3\xe1\xe8\x46\xa5\xc6\xf3\xe6\x2f\xc3\xde\xbe\x3d\xba\xf1\x69\x5f\xa0\xf0\xfa\xfc\x59\x04\xcf\xa1\x52\x81\xc6\x38\x3a\x0f\x0f\xfe\x7e\x38\xce\x65\x1e\x47\x07\x38\x43\x3d\x05\xfd\xeb\xd8\x5b\xff\xf0\xfd\x70\xb0\x0a\x32\xd9\xe0\x0a\x79\xbc\x54\x96\x6e\xf8\x6d\xae\x95\x7a\x67\x0e\x76\xb9\x0a\x19\x90\x33\x54\xe9\x86\x0d\xb2\xc7\x8f\x5f\xd9\xfb\xf3\xf0\x80\xbd\x1e\xcd\x97\x5b\x01\xf4\xab\x5b\x81\x75\xb9\x7a\x8e\xfc\xcb\x00\x28\xed\x5b\x01\x50\xdc\x72\xf5\xb9\xd0\xff\x88\x10\xaa\xff\xe3\x20\x6f\xcf\xaf\x87\x69\xb9\x52\xed\x38\xee\x9f\xef\x36\xfb\xf3\x37\x2a\xbf\x3f\x4b\xe1\x76\x72\xa6\xb3\x9f\x1e\x9f\xed\x0e\x92\x80\xca\xef\xee\xbe\x3a\x65\x5e\x8f\x1f\x8f\x3f\x7f\xd8\xab\xdb\x20\x92\xe5\x82\x58\x3b\xe2\xba\x25\x4e\x6a\xe2\x24\xf9\xfd\x8f\x73\xc4\x55\x4a\x9c\x34\xc4\x0e\xbf\x8e\xb8\x48\x64\x4f\xa5\x89\xab\x5c\x9e\xe3\xba\x28\x88\xf3\x9a\xb8\xc3\xc7\x10\xe7\x8e\x38\x37\xc4\x6d\x42\x6c\x4b\xe2\x04\xfb\xf3\x2f\x63\xe3\x63\x53\xe2\xd6\x12\x57\x1d\x71\x93\x12\xeb\xf4\x39\x86\xc5\x75\xfc\xbc\xa8\xc0\xe0\x96\xdf\x9b\x42\x57\x4f\x3d\x9b\x62\x71\x14\xbd\xac\x6f\xeb\x38\x8a\x16\x2f\x1b\xf8\x8b\x75\x1c\xad\x9e\x8c\xf0\xa2\xe8\xc0\xfb\x7f\xde\xd2\xcf\xf1\x7a\x4f\x3f\x0d\xce\x3f\x53\xfd\xb7\xe6\xd5\xd3\x98\xf1\x83\x62\x7b\xfb\x5b\xd1\x7d\x82\x1d\xb7\xea\xc5\x05\x2b\x78\x70\xab\x74\xbd\x56\x30\xd3\xf6\xb9\xd7\x96\x79\x9a\xac\xfc\x73\x58\x64\x2b\x16\xfa\xe5\x38\xf0\x52\x17\x3a\x69\x32\x5d\x35\xe9\x5a\x25\xab\xc7\x38\x32\x80\xf2\x83\xef\xc3\x27\x5f\xfc\x56\x85\x1e\x00\xe7\xd6\x7f\x3f\x3e\xf1\x68\xd6\x7f\x42\xfe\xbf\x9c\xfe\x5b\xf1\x9b\x8e\xb8\xc8\x89\xf3\x9e\xb8\x74\xc4\x7d\x4b\x9c\xe1\x7d\x47\xdc\xf6\xc4\xb6\x22\x2e\x2a\x62\x97\x10\x37\x21\x56\xd6\xc8\xfb\x2e\x23\xce\x4a\xb9\xb6\x30\x45\x23\x46\xe9\x73\xe2\xce\xca\xfa\xb4\x17\x53\x19\x43\x9c\xc1\x08\x39\x71\x5a\x11\x57\xc0\xd7\x10\xe7\x05\x71\xd3\x13\xb7\x86\xb8\x28\x89\x53\xc4\x6f\x89\x8b\x5e\x8c\x67\xb1\xae\x22\x4e\x6b\x62\x67\x89\x6b\x43\x9c\x66\xc4\xa6\x25\xd6\x35\x71\x5f\x12\x17\x2d\x71\x97\x12\xe7\xc0\x99\x11\x3b\x4d\x9c\xf5\xc4\x45\x4a\x5c\xc0\x74\x5a\x62\xd6\xa8\xcf\x12\x77\x1d\x71\x89\x5c\x56\x62\x25\x56\x30\x75\x89\x60\x69\x0a\x31\x37\xb0\xf8\xf5\xc0\x9e\x13\xd7\x25\xb1\x0b\xb5\xe3\x1d\x6a\xac\x0b\xc1\x65\x91\x4f\x4b\x2c\xac\xb5\xe8\x29\x6a\xa9\xc4\xfc\xe8\x81\x5f\x5f\x4a\x3d\xc0\x04\xbc\xa8\x33\x0f\xc3\x23\x4f\x25\x3e\xea\x6e\x3a\xe9\x47\x59\x10\x77\x2e\xd4\x8f\x75\x85\xf4\xbc\x6d\x88\x75\x4b\xdc\xa3\xa6\x2b\x76\x47\x5c\xa2\x57\xa8\x2f\x23\x2e\x6a\xa9\x1b\xb1\xaa\x84\x38\x0d\x35\xa1\x47\x65\xd0\x87\x01\xa6\x4e\x70\xe9\x4c\x86\x5c\x57\x0a\xdf\xd0\x42\xa9\x89\x6d\xc0\x0f\xdc\x5d\x4f\x9c\x39\xe1\xcf\x99\x30\x1c\x7b\xe2\x5e\x4b\xce\xb6\x95\xda\x8b\x30\x28\x73\x68\x21\x27\xd6\xc0\x97\x10\xb7\x05\x71\x5b\xc9\xc0\xcd\x30\x64\x1b\xd1\x0e\x62\x3a\x68\x0a\xef\x3b\xa9\x07\xb9\xc1\x99\xd7\x1e\xb4\xd6\x09\xcf\x78\x5f\x36\x32\xb8\x5b\x2d\x7a\xb5\x8d\x0c\x6e\xe4\x05\x17\x25\x30\x20\xb6\x96\xe7\xe0\x13\x9e\xa8\xa0\x2d\x27\x9a\xae\xb5\x68\x36\x0b\xc3\xbe\x47\x4f\x0b\xe1\x1b\xbc\x26\x9a\xd8\x34\xd2\x2f\xac\x87\x76\x71\x00\xa0\x1f\xd8\x87\x7e\x40\xb7\xd0\x11\x0e\x02\xac\x31\x35\x71\x93\x8b\xa7\x32\x2d\xfe\x02\x5e\x8d\xf8\xa9\xf4\x02\x5a\xcb\x82\x4e\xe0\x11\x3c\x87\x1e\x74\x23\xda\x44\x2d\x7e\x4d\x29\xfa\xa9\x2d\x71\x8f\x7e\xe0\xe0\x80\xee\x2d\x71\x69\x24\x07\xf4\x51\x04\xbd\xb8\x9c\xb8\xaa\x88\x7b\x78\x00\x3e\x0e\x9e\x6a\x4c\xd8\xd3\x0b\x3e\x5c\xbb\x5e\xb8\xa8\x50\x4f\x26\x9e\xad\x03\x2f\x4d\x23\x7d\x36\xb9\x68\x13\x3c\xa0\x3f\x0e\x7c\x36\x52\x3b\x6a\x85\x3e\xa0\x69\x60\xa8\x82\x0f\xa0\xa3\xa4\x14\x1c\xd0\x2a\x70\x82\x5f\x60\x85\xd6\x53\x1c\xa8\x05\xb1\x2d\xc4\x47\x98\x3b\xd8\x57\x61\x8f\xfd\xdc\x03\xe8\x0c\x7d\x84\x96\x6d\x2b\x58\xa0\x39\xcc\x8b\xbe\x0f\x07\xb7\x93\x3d\xc0\xd4\x00\x8f\x21\x6e\x5a\xe2\xd4\x48\x0f\xda\x52\xb8\xd2\xb9\x78\x10\x39\xa1\x19\xcf\xff\xd5\xc7\xb5\xcc\x0b\xaf\x1d\x68\x17\x3e\x49\x85\x77\x6b\xa4\x4f\x39\xfa\x5f\x0b\x56\xf4\x0e\x1a\x82\xde\x3d\xe6\x52\xfa\x0a\xbd\x78\x2c\x89\xe0\x40\x1f\xf0\xa7\x00\x1a\xc0\x1c\x2d\x5b\xe2\x5c\x07\xaf\x74\xc4\x46\x4b\x2d\xe0\x34\xc5\x9a\x8c\x58\x57\x82\x15\xfc\x23\x7e\x1b\x66\x1b\x74\x80\x7d\xe0\xb4\x4d\xc3\x1c\x69\xc5\x03\x98\x6f\xe0\xa7\x09\x3d\x35\x85\x60\x00\x5f\xe0\x17\x7e\x87\x2f\xbb\x30\xcb\xd1\x17\xf8\x05\x7c\xd8\x4e\xf2\x64\x85\xf8\xb0\x4e\x24\x9e\xee\xa4\x07\x75\x2f\x18\xa1\xc5\xb4\x10\x8d\xa2\xa6\x3a\x0b\x5e\x2e\xa5\x37\xde\xdf\x58\x9b\x05\x6f\x76\x41\xe7\x56\x34\x82\x19\x83\x75\xc0\x83\xd8\x98\x53\x3a\x9c\x1d\x6d\x58\x9b\x65\x9f\xcf\x25\xcc\x30\xf4\x2c\x09\xfa\x07\xf7\xae\x12\x8d\x67\xc1\x1b\xe0\xbf\xb1\xc4\x85\x25\x36\xe0\x0b\xf9\xd1\xd7\x4c\x78\x40\xdf\xf2\x10\x0b\x7a\xf1\x73\xb7\x13\xbd\x43\x9b\xf0\x0a\x78\x28\xc3\x9f\x3e\x60\x45\xfd\xa8\xf1\x7a\x9e\xf8\x3f\x81\x45\x98\x11\xbd\xf8\xbf\xd3\xc2\x57\x16\xf2\xa2\x6e\xf4\x0c\x1e\x2d\xc3\xec\x46\x5d\x38\x57\xbc\xf6\xab\xf0\x6b\x04\x4f\x1a\x78\x46\x7e\xf8\xb3\x0f\xfc\xa3\x1e\xe8\x18\xfd\x00\x97\xd0\x1f\xce\x47\xcc\x63\xf4\xcb\xff\xc1\x6c\xe5\xbc\xbb\xd6\x8a\x73\xc1\x9f\x75\xd7\xd9\x8b\x4f\x2d\xfc\x42\x6b\xd7\x73\x01\x73\xaf\x6c\xfe\x1d\x00\x00\xff\xff\x67\xc3\x3c\x14\x00\x10\x00\x00")

func bindataGoBytes() ([]byte, error) {
	return bindataRead(
		_bindataGo,
		"bindata.go",
	)
}

func bindataGo() (*asset, error) {
	bytes, err := bindataGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "bindata.go", size: 8192, mode: os.FileMode(420), modTime: time.Unix(1511248933, 0)}
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
	"20171115153101_create_initial.down.sql": _20171115153101_create_initialDownSql,
	"20171115153101_create_initial.up.sql": _20171115153101_create_initialUpSql,
	"bindata.go": bindataGo,
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
	"20171115153101_create_initial.down.sql": &bintree{_20171115153101_create_initialDownSql, map[string]*bintree{}},
	"20171115153101_create_initial.up.sql": &bintree{_20171115153101_create_initialUpSql, map[string]*bintree{}},
	"bindata.go": &bintree{bindataGo, map[string]*bintree{}},
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

