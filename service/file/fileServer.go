package file

import (
	"io/ioutil"
	"os"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"strings"
)

type FileServer struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

// GetFilesAndDirs 依据指定路径遍历文件
func (f *FileServer) GetFilesAndDirs(dirPath, filePattern string, depth int) (files, dirs []string, err error) {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, nil, err
	}
	pathSep := string(os.PathSeparator)

	for _, file := range dir {
		if file.IsDir() {
			dirs = append(dirs, dirPath+pathSep+file.Name())
			if depth > 0 {
				nextFiles, nextDirs, nextErr := f.GetFilesAndDirs(dirPath+pathSep+file.Name(), filePattern, depth-1)
				if nextErr != nil {
					return files, dirs, nextErr
				}
				files = append(files, nextFiles...)
				dirs = append(dirs, nextDirs...)
			}
		} else {
			if filePattern != "" && strings.HasSuffix(file.Name(), filePattern) {
				continue
			}
			files = append(files, dirPath+pathSep+file.Name())
		}
	}

	return
}
