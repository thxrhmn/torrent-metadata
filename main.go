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
	Name         string         `json:"name"`
	CreatedBy    string         `json:"created_by"`
	CreationDate int64          `json:"creation_date"`
	Announce     string         `json:"announce"`
	AnnounceList []interface{}  `json:"announce_list"`
	Size         HumanReadable  `json:"size"`
	PieceLength  HumanReadable  `json:"piece_length"`
	TotalFiles   int            `json:"total_files"`
	Files        []TorrentFiles `json:"files"`
}

type TorrentFiles struct {
	Length HumanReadable `json:"length"`
	Path   []string      `json:"path"`
}

type HumanReadable struct {
	Size0 int64  `json:"0"`
	Size1 string `json:"1"`
}

func main() {
	var filePath string

	flag.StringVar(&filePath, "file", "", "Path to the torrent file")
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

	name := info["name"].(string)
	pieceLength := info["piece length"].(int64)
	pieceLengthString := humanize.Bytes(uint64(pieceLength))
	createdBy := decoderMap["created by"].(string)
	creationDate := decoderMap["creation date"].(int64)
	var announce string
	announceInterface := decoderMap["announce"]
	if announceInterface != nil {
		announce = decoderMap["announce"].(string)
	}
	announceList := extractAnnounceList(decoderMap["announce-list"])
	var torrentSize int64

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
		Size: HumanReadable{
			Size0: torrentSize,
			Size1: humanize.Bytes(uint64(torrentSize)),
		},
		PieceLength: HumanReadable{
			Size0: pieceLength,
			Size1: pieceLengthString,
		},
		TotalFiles: len(torrents),
		Files:      torrents,
	}

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	// Write the JSON data to the file
	jsonPath := filePath + ".json"
	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("JSON data successfully written to", jsonPath)
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

func extractFiles(files []interface{}) (int64, []TorrentFiles) {
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

		pathArray, ok := fileMap["path"].([]interface{})
		if !ok {
			continue
		}

		// Convert the array elements to strings
		var pathStrings []string
		for _, elem := range pathArray {
			if str, ok := elem.(string); ok {
				pathStrings = append(pathStrings, str)
			}
		}

		data := TorrentFiles{
			Length: HumanReadable{
				Size0: length,
				Size1: humanize.Bytes(uint64(length)),
			},
			Path: pathStrings,
		}
		torrentSize += length
		result = append(result, data)
	}

	return torrentSize, result
}
