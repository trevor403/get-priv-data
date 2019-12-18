package main

// SteamPackage describes the CDN a zip package
type SteamPackage struct {
	File   string `json:"file"`
	Size   string `json:"size"`
	Sha2   string `json:"sha2"`
	Zipvz  string `json:"zipvz"`
	Sha2Vz string `json:"sha2vz"`
}

// SteamClientWin32 minimal Win32 manifest entry
type SteamClientWin32 struct {
	Win32 struct {
		Version   string       `json:"version"`
		BinsWin32 SteamPackage `json:"bins_win32"`
	} `json:"win32"`
}
