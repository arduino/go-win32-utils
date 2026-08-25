// This file is part of go-win32-utils.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: BSD-3-Clause

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
