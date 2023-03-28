//
// Copyright 2018-2023 ARDUINO SA. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package devicenotification

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	win32 "github.com/arduino/go-win32-utils"
	"golang.org/x/sys/windows"
)

var osThreadID atomic.Uint32

// Start the device add/remove notification process, at every event a call to eventCB will be performed.
// This function will block until interrupted by the given context. Errors will be passed to errorCB.
// Returns error if sync process can't be started.
func Start(ctx context.Context, eventCB func(), errorCB func(msg string)) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	osThreadID.Store(windows.GetCurrentThreadId())

	eventsChan := make(chan bool, 1)
	var eventsChanLock sync.Mutex
	windowCallback := func(hwnd syscall.Handle, msg uint32, wParam uintptr, lParam uintptr) uintptr {
		// This mutex is required because the callback may be called
		// asynchronously by the OS threads, even after the channel has
		// been closed and the callback unregistered...
		eventsChanLock.Lock()
		if eventsChan != nil {
			select {
			case eventsChan <- true:
			default:
			}
		}
		eventsChanLock.Unlock()
		return win32.DefWindowProc(hwnd, msg, wParam, lParam)
	}
	defer func() {
		eventsChanLock.Lock()
		close(eventsChan)
		eventsChan = nil
		eventsChanLock.Unlock()
	}()

	go func() {
		for {
			if _, ok := <-eventsChan; !ok {
				return
			}
			eventCB()
		}
	}()

	// We must create the window used to receive notifications in the same
	// thread that destroys it otherwise it would fail
	windowHandle, className, err := createWindow(windowCallback)
	if err != nil {
		return err
	}
	defer func() {
		if err := destroyWindow(windowHandle, className); err != nil {
			errorCB(err.Error())
		}
	}()

	notificationsDevHandle, err := registerNotifications(windowHandle)
	if err != nil {
		return err
	}
	defer func() {
		if err := unregisterNotifications(notificationsDevHandle); err != nil {
			errorCB(err.Error())
		}
	}()

	go func() {
		<-ctx.Done()
		_ = win32.PostMessage(windowHandle, win32.WMQuit, 0, 0)
	}()

	for {
		// Verify running thread prerequisites
		if currThreadID := windows.GetCurrentThreadId(); currThreadID != osThreadID.Load() {
			panic(fmt.Sprintf("this function must run on the main OS Thread: currThread=%d, osThread=%d", currThreadID, osThreadID.Load()))
		}

		var m win32.TagMSG
		if res := win32.GetMessage(&m, windowHandle, win32.WMQuit, win32.WMQuit); res == 0 { // 0 means we got a WMQUIT
			break
		} else if res == -1 { // -1 means that an error occurred
			err := windows.GetLastError()
			errorCB("error consuming messages: " + err.Error())
			break
		} else {
			// we got a message != WMQuit, it should not happen but, just in case...
			win32.TranslateMessage(&m)
			win32.DispatchMessage(&m)
		}
	}

	return nil
}

func createWindow(windowCallback win32.WindowProcCallback) (syscall.Handle, *byte, error) {
	// Verify running thread prerequisites
	if currThreadID := windows.GetCurrentThreadId(); currThreadID != osThreadID.Load() {
		panic(fmt.Sprintf("this function must run on the main OS Thread: currThread=%d, osThread=%d", currThreadID, osThreadID.Load()))
	}

	moduleHandle, err := win32.GetModuleHandle(nil)
	if err != nil {
		return syscall.InvalidHandle, nil, err
	}

	className, err := syscall.BytePtrFromString("device-notification")
	if err != nil {
		return syscall.InvalidHandle, nil, err
	}
	windowClass := &win32.WndClass{
		Instance:  moduleHandle,
		ClassName: className,
		WndProc:   syscall.NewCallback(windowCallback),
	}
	if _, err := win32.RegisterClass(windowClass); err != nil {
		return syscall.InvalidHandle, nil, fmt.Errorf("registering new window: %s", err)
	}

	windowHandle, err := win32.CreateWindowEx(win32.WsExTopmost, className, className, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	if err != nil {
		return syscall.InvalidHandle, nil, fmt.Errorf("creating window: %s", err)
	}
	return windowHandle, className, nil
}

func destroyWindow(windowHandle syscall.Handle, className *byte) error {
	// Verify running thread prerequisites
	if currThreadID := windows.GetCurrentThreadId(); currThreadID != osThreadID.Load() {
		panic(fmt.Sprintf("this function must run on the main OS Thread: currThread=%d, osThread=%d", currThreadID, osThreadID.Load()))
	}

	if err := win32.DestroyWindowEx(windowHandle); err != nil {
		return fmt.Errorf("error destroying window: %s", err)
	}
	if err := win32.UnregisterClass(className); err != nil {
		return fmt.Errorf("error unregistering window class: %s", err)
	}
	return nil
}

func registerNotifications(windowHandle syscall.Handle) (syscall.Handle, error) {
	// Verify running thread prerequisites
	if currThreadID := windows.GetCurrentThreadId(); currThreadID != osThreadID.Load() {
		panic(fmt.Sprintf("this function must run on the main OS Thread: currThread=%d, osThread=%d", currThreadID, osThreadID.Load()))
	}

	notificationFilter := win32.DevBroadcastDeviceInterface{
		DwDeviceType: win32.DbtDevtypeDeviceInterface,
		ClassGUID:    win32.UsbEventGUID,
	}
	notificationFilter.DwSize = uint32(unsafe.Sizeof(notificationFilter))

	var flags uint32 = win32.DeviceNotifyWindowHandle | win32.DeviceNotifyAllInterfaceClasses
	notificationsDevHandle, err := win32.RegisterDeviceNotification(windowHandle, &notificationFilter, flags)
	if err != nil {
		return syscall.InvalidHandle, err
	}

	return notificationsDevHandle, nil
}

func unregisterNotifications(notificationsDevHandle syscall.Handle) error {
	// Verify running thread prerequisites
	if currThreadID := windows.GetCurrentThreadId(); currThreadID != osThreadID.Load() {
		panic(fmt.Sprintf("this function must run on the main OS Thread: currThread=%d, osThread=%d", currThreadID, osThreadID.Load()))
	}

	if err := win32.UnregisterDeviceNotification(notificationsDevHandle); err != nil {
		return fmt.Errorf("error unregistering device notifications: %s", err)
	}
	return nil
}
