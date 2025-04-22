package media

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

var MimeTypes = map[string]string{
	"image/jpeg": ".jpeg",
	"image/png":  ".png",
}

func ConvertoTimestampToUTC(timestamp int) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(int64(timestamp), 0).UTC()
}

func GetFileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func GetExtensionByMimeType(mimeType string) (string, error) {

	ext, ok := MimeTypes[mimeType]
	if !ok {
		return "", fmt.Errorf("unable to find extension for mime type %s", mimeType)
	}

	return ext, nil
}
