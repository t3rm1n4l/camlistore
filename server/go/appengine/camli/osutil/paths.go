// +build appengine

/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package osutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func HomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("HOMEPATH")
	}
	return os.Getenv("HOME")
}

func CacheDir() string {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		panic("CacheDir not implemented on OS == " + runtime.GOOS)
	}
	return filepath.Join(HomeDir(), ".cache")
}

func CamliConfigDir() string {
	if p := os.Getenv("CAMLI_CONFIG_DIR"); p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "camli")
	}
	return filepath.Join(HomeDir(), ".camli")
}

func UserServerConfigPath() string {
	return filepath.Join(CamliConfigDir(), "serverconfig")
}

// Find the correct absolute path corresponding to a relative path, 
// searching the following sequence of directories:
// 1. Working Directory
// 2. CAMLI_CONFIG_DIR (deprecated, will complain if this is on env)
// 3. (windows only) APPDATA/camli
// 4. All directories in CAMLI_INCLUDE_PATH (standard PATH form for OS)
func FindCamliInclude(configFile string) (absPath string, err os.Error) {
	// Try to open as absolute / relative to CWD
	_, err = os.Stat(configFile)
	if err == nil {
		return configFile, nil
	}
	if filepath.IsAbs(configFile) {
		// End of the line for absolute path
		return "", err
	}

	// Try the config dir
	configDir := CamliConfigDir()
	if _, err = os.Stat(filepath.Join(configDir, configFile)); err == nil {
		return filepath.Join(configDir, configFile), nil
	}

	// Finally, search CAMLI_INCLUDE_PATH
	p := os.Getenv("CAMLI_INCLUDE_PATH")
	for _, d := range strings.Split(p, string(filepath.ListSeparator)) {
		if _, err = os.Stat(filepath.Join(d, configFile)); err == nil {
			return filepath.Join(d, configFile), nil
		}
	}

	return "", os.ErrNotExist
}
