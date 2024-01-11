# TORRENT METADATA EXTRACTOR

This project is a simple tool for extracting metadata from torrent files. With this script, users can easily access detailed information about torrent files without the need to open a torrent client.

## Key Features:
- Quickly and efficiently extract metadata from torrent files.
- Display information such as file name, size, created by, and more.

## Usage:
1. Make sure you have Golang installed on your system.
2. Clone this repository and install its package  by running
```bash
git clone https://github.com/thxrhmn/torrent-metadata.git
cd torrent-metadata
go mod tidy
```
3. Run the script by following the command
```bash
go run main.go -file <your .torrent file path>
```
example:
```bash
go run main.go -file my.torrent
```
example result:
```json
// my.torrent.json
{
   "name": "My Torrent File",
   "created_by": "@thxrhmn",
   "creation_date": 1704957362,
   "announce": "udp://example.com:6969/announce",
   "announce_list": [
      "udp://example.com:6969/announce",
      "udp://example.com:6970/announce",
   ],
   "size": {
      "0": 1000000000,
      "1": "1.0 GB"
   },
   "piece_length": {
      "0": 4194304,
      "1": "4.2 MB"
   },
   "total_files": 2,
   "files": [
      {
         "length": {
            "0": 122,
            "1": "122 B"
         },
         "path": "0. website.url"
      },
      {
         "length": {
            "0": 1000000000,
            "1": "1.0 GB"
         },
         "path": "1. video.mp4"
      }
   ]
}
```