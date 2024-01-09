package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	bencode "github.com/jackpal/bencode-go"
)

type Torrent struct {
	Name         interface{}    `json:"name"`
	CreatedBy    interface{}    `json:"created_by"`
	CreationDate interface{}    `json:"creation_date"`
	Announce     interface{}    `json:"announce"`
	AnnounceList []interface{}  `json:"announce_list"`
	Size         string         `json:"size"`
	TotalFiles   int            `json:"total_files"`
	Files        []TorrentFiles `json:"files"`
}

type TorrentFiles struct {
	Length int64         `json:"length"`
	Path   []interface{} `json:"path"`
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "file", "tor.torrent", "Path to the torrent file")
	flag.Parse()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	decoder, err := bencode.Decode(file)
	if err != nil {
		fmt.Println("Error decoding file:", err)
		return
	}

	decoderMap, ok := decoder.(map[string]interface{})
	if !ok {
		fmt.Println("Error: decoder is not a map[string]interface{} type")
		return
	}

	var torrents []TorrentFiles

	info, ok := decoderMap["info"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: info is not a map[string]interface{} type")
		return
	}

	name := info["name"]
	createdBy := decoderMap["created by"]
	creationDate := decoderMap["creation date"]
	announce := decoderMap["announce"]
	announceList := extractAnnounceList(decoderMap["announce-list"])
	var torrentSize string

	files, ok := info["files"].([]interface{})
	if ok {
		size, tor := extractFiles(files)
		torrents = tor
		torrentSize = size
	}

	data := Torrent{
		Name:         name,
		CreatedBy:    createdBy,
		CreationDate: creationDate,
		Announce:     announce,
		AnnounceList: announceList,
		Size:         torrentSize,
		TotalFiles:   len(torrents),
		Files:        torrents,
	}

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}

func extractAnnounceList(announceList interface{}) []interface{} {
	var result []interface{}

	list, ok := announceList.([]interface{})
	if !ok {
		return result
	}

	for _, val := range list {
		innerSlice, ok := val.([]interface{})
		if !ok || len(innerSlice) == 0 {
			continue
		}

		result = append(result, innerSlice[0])
	}

	return result
}

func extractFiles(files []interface{}) (string, []TorrentFiles) {
	var result []TorrentFiles

	var torrentSize int64

	for _, file := range files {
		fileMap, ok := file.(map[string]interface{})
		if !ok {
			continue
		}

		length, ok := fileMap["length"].(int64)
		if !ok {
			continue
		}

		path, ok := fileMap["path"].([]interface{})
		if !ok {
			continue
		}

		data := TorrentFiles{
			Length: length,
			Path:   path,
		}
		torrentSize += length
		result = append(result, data)
	}

	size := humanize.Bytes(uint64(torrentSize))
	return size, result
}
