package utils

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
	if depth == 0 {
		dirs = []string{dirPath}
	}
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, nil, err
	}
	pathSep := string(os.PathSeparator)

	for _, file := range dir {
		if file.IsDir() {
			if depth == 0 {
				continue
			}
			dirs = append(dirs, dirPath+pathSep+file.Name())
			if depth > 1 {
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

// GetLimitFilesAndDirs 获取指定文件夹下所有文件和文件夹路径
func (f *FileServer) GetLimitFilesAndDirs(dirPath, filePattern string) (files, dirs []string, err error) {
	return f.GetFilesAndDirs(dirPath, filePattern, 1)
}

// GetAllFiles 层级遍历指定文件夹下所有文件，获取其中所有文件路径
func (f *FileServer) GetAllFiles(dirPath, filePattern string) (files []string, err error) {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	pathSep := string(os.PathSeparator)

	for _, file := range dir {
		if file.IsDir() {
			nextFiles, nextErr := f.GetAllFiles(dirPath+pathSep+file.Name(), filePattern)
			if nextErr != nil {
				return files, nextErr
			}
			files = append(files, nextFiles...)
		} else {
			files = append(files, dirPath+pathSep+file.Name())
		}
	}

	return
}

// GetAllDirs 层级遍历指定文件夹下所有文件夹路径
func (f *FileServer) GetAllDirs(dirPath string, depth int) (dirs []string, err error) {
	if depth == 0 {
		return []string{dirPath}, nil
	}
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	pathSep := string(os.PathSeparator)

	for _, file := range dir {
		if file.IsDir() {
			dirs = append(dirs, dirPath+pathSep+file.Name())
			if depth > 1 {
				nextDirs, nextErr := f.GetAllDirs(dirPath+pathSep+file.Name(), depth-1)
				if nextErr != nil {
					return dirs, nextErr
				}
				dirs = append(dirs, nextDirs...)
			}
		}
	}

	return
}
