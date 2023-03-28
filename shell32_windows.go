package win32

import (
	"fmt"
	"syscall"
	"unsafe"
)

func init() {
	if err1, err2 := procSHGetKnownFolderPath.Find(), procCoTaskMemFree.Find(); err1 != nil || err2 != nil {
		procSHGetKnownFolderPath = nil
	}
	if err := procSHGetFolderPathW.Find(); err != nil {
		procSHGetFolderPathW = nil
	}
}

func getFolder(id *folderIdentifier) (string, error) {
	if procSHGetKnownFolderPath != nil {
		var pathptr *uint16
		if err := getKnownFolderPath(id.FOLDERID, 0, 0, &pathptr); err != nil {
			return "", err
		}
		defer taskMemFree(uintptr(unsafe.Pointer(pathptr)))
		return syscall.UTF16ToString((*[65535]uint16)(unsafe.Pointer(pathptr))[:]), nil
	}
	if procSHGetFolderPathW != nil {
		path := make([]uint16, 1024) // MAX_PATH in win32 API is defined as 260, so 1024 should be fine
		if err := getFolderPath(0, id.CSIDL, 0, 0, &path[0]); err != nil {
			return "", err
		}
		return syscall.UTF16ToString(path), nil
	}
	return "", fmt.Errorf("could not call shell32 API to retrieve folder")
}

// GetDocumentsFolder returns the Document folder
func GetDocumentsFolder() (string, error) {
	return getFolder(documentsFolder)
}

// GetLocalAppDataFolder returns the LocalAppData folder
func GetLocalAppDataFolder() (string, error) {
	return getFolder(localAppDataFolder)
}

// GetRoamingAppDataFolder returns the AppData folder
func GetRoamingAppDataFolder() (string, error) {
	return getFolder(roamingAppDataFolder)
}

var documentsFolder = &folderIdentifier{FOLDERID: folderIDDocuments, CSIDL: csidlMyDocuments}
var roamingAppDataFolder = &folderIdentifier{FOLDERID: folderIDRoamingAppData, CSIDL: csidlAppData}
var localAppDataFolder = &folderIdentifier{FOLDERID: folderIDLocalAppData, CSIDL: csidlAppData}
