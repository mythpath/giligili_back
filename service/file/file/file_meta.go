package file

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// /test/file.txt should have .meta: /test/.meta/file.txt.meta
// /test/dir should have .meta: /test/.meta/dir.meta
func dirCheckMeta(fs FileSystem, path string) error {
	fDir, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer fDir.Close()
	filebs, err := dirList(fDir, path)
	if err != nil {
		return err
	}
	mDirPath := filepath.Join(path, metaDirConstName)
	mDir, err := fs.Open(mDirPath)
	if os.IsNotExist(err) {
		err = fs.MkdirAll(mDirPath, os.ModePerm)
		if err != nil {
			return err
		}
		for _, file := range filebs {
			_, err = createMeta(fs, file.Url, file.Type, defaultTimeStamp, "")
			if err != nil {
				return err
			}
		}
		mDir, _ = fs.Open(mDirPath)
	} else if err != nil {
		return err
	}
	defer mDir.Close()
	metafilebs, err := dirList(mDir, filepath.Join(path, metaDirConstName))
	if err != nil {
		return err
	}
	/* simple implements */
	// delete garbage meta
	for _, m := range metafilebs {
		isFound := false
		for _, f := range filebs {
			if f.Name+metaTokenConstSuffix == m.Name {
				isFound = true
				break
			}
		}
		if !isFound {
			fs.Delete(m.Url)
		}
	}
	// add meta
	for _, f := range filebs {
		isFound := false
		for _, m := range metafilebs {
			if m.Name == f.Name+metaTokenConstSuffix {
				isFound = true
				break
			}
		}
		if !isFound {
			_, err = createMeta(fs, f.Url, f.Type, defaultTimeStamp, "")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func dirListMeta(fs FileSystem, path string) ([]FileDesc, error) {
	dirpath := filepath.Join(path, metaDirConstName)
	fmeta, err := fs.Open(dirpath)
	if err != nil {
		return nil, err
	}
	metaFDs, err := fmeta.Readdir(-1)
	if err != nil {
		return nil, err
	}
	dataDir, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	databs, err := dirList(dataDir, path)
	if err != nil {
		return nil, err
	}
	sort.Sort(byName(metaFDs))
	filejs := []FileDesc{}
	for _, item := range metaFDs {
		iname := item.Name()
		if iname[0] == '.' {
			continue
		}
		ipath := filepath.Join(dirpath, iname)
		buf, err := readMetaFile(ipath, fs)
		if err != nil {
			return nil, err
		}
		filej := FileDesc{}
		err = json.Unmarshal(buf, &filej)
		if err != nil {
			return nil, err
		}
		for i, _ := range databs {
			if filej.Name == databs[i].Name {
				filej.Url = databs[i].Url
			}
		}
		filejs = append(filejs, filej)
	}

	return filejs, nil
}

func readMetaFile(path string, fs FileSystem) ([]byte, error) {
	var fmeta *os.File
	if _, err := fs.Stat(path); err == nil {
		fmeta, err = fs.OpenFile(path, os.O_RDWR, os.ModePerm)
		if err != nil {
			return nil, err
		}
		defer fmeta.Close()
	} else {
		return nil, err
	}
	buf, err := ioutil.ReadAll(fmeta)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func writeMetaFile(path string, fs FileSystem, buf []byte) error {
	fmeta, err := fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fmeta.Close()
	_, err = fmeta.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func createMeta(fs FileSystem, fPath, fType, cTime, cName string) (*FileDesc, error) {
	fileMeta := &FileMeta{}
	fileMeta.Type = fType
	fileMeta.Name = filepath.Base(fPath)
	fileMeta.CreatedBy = cName
	if cTime == "" {
		fileMeta.CreatedAt = strconv.FormatInt(time.Now().Unix(), 10)
	} else {
		fileMeta.CreatedAt = cTime
	}
	fileMeta.UpdatedBy = cName
	fileMeta.UpdatedAt = fileMeta.CreatedAt
	buf, err := json.Marshal(fileMeta)
	if err != nil {
		return nil, err
	}

	if _, err := fs.Stat(filepath.Join(filepath.Dir(fPath), metaDirConstName)); os.IsNotExist(err) {
		err = fs.MkdirAll(filepath.Join(filepath.Dir(fPath), metaDirConstName), os.ModePerm)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	metaFilePath := filepath.Join(filepath.Dir(fPath), metaDirConstName,
		filepath.Base(fPath)+metaTokenConstSuffix)
	err = writeMetaFile(metaFilePath, fs, buf)
	if err != nil {
		return nil, err
	}
	jsonFile := fileMeta.ToFileJson()
	jsonFile.Url = fPath
	return jsonFile, nil
}

func updateMeta(fs FileSystem, oldPath, newPath, uName string) (*FileDesc, error) {
	metaFilePath := filepath.Join(filepath.Dir(oldPath), metaDirConstName, filepath.Base(oldPath)+metaTokenConstSuffix)
	buf, err := readMetaFile(metaFilePath, fs)
	if err != nil {
		return nil, err
	}

	fileMeta := &FileMeta{}
	err = json.Unmarshal(buf, fileMeta)
	if err != nil {
		return nil, err
	}
	fileMeta.Name = filepath.Base(newPath)
	fileMeta.UpdatedBy = uName
	fileMeta.UpdatedAt = strconv.FormatInt(time.Now().Unix(), 10)
	buf, err = json.Marshal(fileMeta)
	if err != nil {
		return nil, err
	}

	metaFilePath = filepath.Join(filepath.Dir(newPath), metaDirConstName, filepath.Base(newPath)+metaTokenConstSuffix)
	err = writeMetaFile(metaFilePath, fs, buf)
	if err != nil {
		return nil, err
	}

	jsonFile := fileMeta.ToFileJson()
	jsonFile.Url = newPath

	return jsonFile, nil
}

func deleteMeta(fs FileSystem, fPath string) error {
	metaFilePath := filepath.Join(filepath.Dir(fPath), metaDirConstName, filepath.Base(fPath)+metaTokenConstSuffix)
	return fs.Delete(metaFilePath)
}
