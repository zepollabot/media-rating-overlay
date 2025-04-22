// Media Rating Overlay - A tool for adding rating information to media posters
// Copyright (C) 2025 Pietro Pollarolo
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// Contact: [Your email address]

package utils

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

func ConvertoTimestampToUTC(timestamp int) time.Time {

	return time.Unix(int64(timestamp), 0).UTC()
}
