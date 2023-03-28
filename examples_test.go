//
// Copyright 2018-2023 ARDUINO SA. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package win32_test

import (
	"fmt"

	win32 "github.com/arduino/go-win32-utils"
)

func Example() {
	d, err := win32.GetDocumentsFolder()
	fmt.Printf("Documents       folder: [err=%v] %s\n", err, d)
	d, err = win32.GetLocalAppDataFolder()
	fmt.Printf("Local AppData   folder: [err=%v] %s\n", err, d)
	d, err = win32.GetRoamingAppDataFolder()
	fmt.Printf("Roaming AppData folder: [err=%v] %s\n", err, d)
}
