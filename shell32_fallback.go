//go:build !windows

//
// Copyright 2018-2023 ARDUINO SA. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package win32

import (
	"fmt"
	"runtime"
)

// The functions defined below allow compile on non-Windows OS. The caller
// may choose to not call those functions based on runtime.GOOS value.

// GetDocumentsFolder returns the Document folder
func GetDocumentsFolder() (string, error) {
	return "", fmt.Errorf("operating system not supported: %s", runtime.GOOS)
}

// GetLocalAppDataFolder returns the LocalAppData folder
func GetLocalAppDataFolder() (string, error) {
	return "", fmt.Errorf("operating system not supported: %s", runtime.GOOS)
}

// GetRoamingAppDataFolder returns the AppData folder
func GetRoamingAppDataFolder() (string, error) {
	return "", fmt.Errorf("operating system not supported: %s", runtime.GOOS)
}
