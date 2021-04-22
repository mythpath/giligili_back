package utils

import (
	"github.com/sirupsen/logrus"
	"math"
	"regexp"
	"strconv"
)

type FileSize struct {
}

func (f *FileSize) GenerateFileSize(line string) int64 {
	re := "^([0-9]+)(B|K|k|M|m|G|g|KB|kb|kB|MB|mb|mB|GB|gb|gB{0,1})$"
	fileSizeRegexp := regexp.MustCompile(re)
	paras := fileSizeRegexp.FindStringSubmatch(line)

	size, errAtoi := strconv.Atoi(paras[1])
	if errAtoi != nil {
		logrus.WithFields(logrus.Fields{
			"file size": paras[1],
			"err msg":   errAtoi,
		}).Errorln("convert string to int failed.")
		return 0
	}

	// coefficient: convert unit to Bytes
	coefficient := int64(1)
	switch paras[2] {
	case "B":
		break
	case "KB", "K", "k", "kB", "kb":
		coefficient *= int64(math.Pow(1024, 1))
	case "MB", "M", "m", "mb", "mB":
		coefficient *= int64(math.Pow(1024, 2))
	case "GB", "G", "g", "gb", "gB":
		coefficient *= int64(math.Pow(1024, 3))
	}

	return int64(size) * coefficient
}
