package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/andygrunwald/vdf"
)

const manifestURL = "https://steamcdn-a.akamaihd.net/client/steam_client_win32"
const downloadURL = "https://steamcdn-a.akamaihd.net/client"

func getSteamManifest() (*SteamClientWin32, error) {
	url, err := url.Parse(manifestURL)
	if err != nil {
		return nil, err
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	url.RawQuery = ts

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parser := vdf.NewParser(resp.Body)
	root, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}

	clientManifest := SteamClientWin32{}
	err = json.Unmarshal(buf, &clientManifest)
	if err != nil {
		return nil, err
	}

	return &clientManifest, nil
}

func getSteamUI(zipFiles []*zip.File) ([]byte, error) {
	for _, zipFile := range zipFiles {
		if zipFile.Name != "SteamUI.dll" {
			continue
		}

		f, err := zipFile.Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return ioutil.ReadAll(f)
	}

	return nil, fmt.Errorf("cloud not find 'SteamUI.dll' in zip")
}

func getFromServer() ([]byte, error) {
	clientManifest, err := getSteamManifest()
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(downloadURL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, clientManifest.Win32.BinsWin32.File)

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	zipSize, err := strconv.Atoi(clientManifest.Win32.BinsWin32.Size)
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) != zipSize {
		return nil, fmt.Errorf("sizes do not match")
	}

	if fmt.Sprintf("%x", sha256.Sum256(bodyBytes)) != clientManifest.Win32.BinsWin32.Sha2 {
		return nil, fmt.Errorf("sha256 sums do not match")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		return nil, err
	}

	dllBytes, err := getSteamUI(zipReader.File)
	if err != nil {
		return nil, err
	}

	return dllBytes, nil
}
