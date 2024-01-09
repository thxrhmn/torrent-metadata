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