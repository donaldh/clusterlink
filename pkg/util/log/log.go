// Copyright 2023 The ClusterLink Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logutils

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

const (
	logrusFieldStack = 6
)

// SetLog sets logrus logger (format, file, level).
func SetLog(logLevel string, logFileName string) (*os.File, error) {
	var f *os.File
	if logFileName != "" {
		usr, _ := user.Current()
		logFileFullPath := path.Join(usr.HomeDir, logFileName)
		createLogFolder(logFileFullPath)

		f, err := os.OpenFile(logFileFullPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
		fmt.Printf("Creating log file: %v\n", logFileFullPath)
		if err != nil {
			return nil, fmt.Errorf("error opening log file: %w", err)
		}
		// assign it to the standard logger
		logrus.SetOutput(f)
	}

	// Set logrus.
	ll, err := logrus.ParseLevel(logLevel)
	if err != nil {
		ll = logrus.ErrorLevel
	}
	logrus.SetLevel(ll)
	logrus.SetFormatter(&formatter{
		TextFormatter: &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			PadLevelText:    true,
			DisableQuote:    true,
		},
	})
	return f, nil
}

type formatter struct {
	*logrus.TextFormatter
}

// Format sets the line number and file for errors and fatal
func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level <= logrus.ErrorLevel {
		_, file, line, _ := runtime.Caller(logrusFieldStack)
		entry.Data["file"] = file
		entry.Data["line"] = fmt.Sprintf("%d", line)
	}

	return f.TextFormatter.Format(entry)
}

// createLogFolder create the log folder if not exists.
func createLogFolder(filePath string) {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("Failed to create directory:", err)
		return
	}
}
