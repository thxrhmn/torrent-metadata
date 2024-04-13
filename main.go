package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	maxGoroutines   = 1
	torrentPath     = "/root/etc/torrents"
	torrentDonePath = "/root/etc/torrents-done"
)

type TorrentFiles struct {
	Path   string            `json:"path" bson:"path"`
	Length HumanReadableSize `json:"length" bson:"length"`
}

type Torrent struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UUID         string             `json:"uuid" bson:"uuid"`
	Name         string             `json:"name" bson:"name"`
	CreatedBy    string             `json:"created_by" bson:"created_by"`
	CreationDate int64              `json:"creation_date" bson:"creation_date"`
	TotalFiles   int                `json:"total_files" bson:"total_files"`
	Hash         string             `json:"hash" bson:"hash"`
	Magnet       Magnet             `json:"magnet" bson:"magnet"`
	PieceLength  int64              `json:"piece_length" bson:"piece_length"`
	Length       HumanReadableSize  `json:"length" bson:"length"`
	Files        []TorrentFiles     `json:"files" bson:"files"`
	CreatedAt    int64              `json:"created_at" bson:"created_at"`
	UpdatedAt    int64              `json:"updated_at" bson:"updated_at"`
}

type Magnet struct {
	DisplayName string   `json:"display_name" bson:"display_name"`
	InfoHash    string   `json:"info_hash" bson:"info_hash"`
	Trackers    []string `json:"trackers" bson:"trackers"`
}

type HumanReadableSize struct {
	Length1 int64  `json:"0" bson:"0"`
	Length2 string `json:"1" bson:"1"`
}

func main() {
	// Waktu sebelum eksekusi operasi database
	start := time.Now()

	var dirPath string

	flag.StringVar(&dirPath, "dir", "", "Path to the directory containing .torrent files")
	flag.Parse()

	if dirPath == "" {
		log.Fatal("Please provide a directory path using the -dir flag")
	}

	// MongoDB connection URI
	uri := "mongodb://thxrhmn:thxrhmn@95.111.231.236:27017/"
	// uri := "mongodb://thxrhmn:thxrhmn@localhost:27017/"
	dbName := "torsnif"
	collectionName := "torrents"

	// Create a MongoDB session
	ctx := context.Background()
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: uri})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close(ctx)

	// Collection reference
	collection := client.Database(dbName).Collection(collectionName)

	// Find all .torrent files in the directory
	torrentFiles, err := findTorrentFiles(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create a wait group for Goroutines
	var wg sync.WaitGroup

	// Buat buffered channel untuk mengontrol jumlah goroutine
	semaphore := make(chan struct{}, maxGoroutines)

	for _, filePath := range torrentFiles {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		go func(filePath string) {
			defer func() {
				wg.Done()
				<-semaphore // Release semaphore
			}()

			err := processTorrentFile(ctx, collection, filePath)
			if err != nil {
				log.Printf("Error processing %s: %v\n", filePath, err)
			}
		}(filePath)
	}

	// Wait for all Goroutines to finish
	wg.Wait()

	fmt.Println("All torrents processed successfully")

	// Waktu setelah eksekusi operasi database
	end := time.Now()
	// Menghitung selisih waktu
	duration := end.Sub(start)
	fmt.Println("Waktu eksekusi:", duration)
}

func findTorrentFiles(dirPath string) ([]string, error) {
	var torrentFiles []string

	// Walk through the directory and find .torrent files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".torrent" {
			torrentFiles = append(torrentFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return torrentFiles, nil
}

func processTorrentFile(ctx context.Context, collection *qmgo.Collection, torrentPath string) error {
	torrentInfo, err := metainfo.LoadFromFile(torrentPath)
	if err != nil {
		log.Fatal(err)
	}

	infoHash := torrentInfo.HashInfoBytes()
	info, err := torrentInfo.UnmarshalInfo()
	if err != nil {
		fmt.Println(err)
	}

	magnet := torrentInfo.Magnet(&infoHash, &info)

	var files []TorrentFiles
	for _, file := range info.Files {
		path := file.DisplayPath(&info)
		length := file.Length

		data := TorrentFiles{
			Path: path,
			Length: HumanReadableSize{
				Length1: length,
				Length2: humanize.Bytes(uint64(length)),
			},
		}
		files = append(files, data)
	}

	// Generate a UUID
	uuid := uuid.New().String()

	torrent := Torrent{
		ID:           primitive.NewObjectID(),
		UUID:         uuid,
		Name:         magnet.DisplayName,
		CreatedBy:    torrentInfo.CreatedBy,
		CreationDate: torrentInfo.CreationDate,
		PieceLength:  info.PieceLength,
		Length: HumanReadableSize{
			Length1: info.TotalLength(),
			Length2: humanize.Bytes(uint64(info.TotalLength())),
		},
		TotalFiles: len(info.Files),
		Hash:       infoHash.HexString(),
		Magnet: Magnet{
			DisplayName: magnet.DisplayName,
			InfoHash:    magnet.InfoHash.HexString(),
			Trackers:    magnet.Trackers,
		},
		Files:     files,
		CreatedAt: time.Now().Unix(), // Set createdAt field with current Unix time
		UpdatedAt: time.Now().Unix(), // Set updatedAt field with current Unix time
	}

	// Lakukan query ke database untuk mencari hash yang sudah ada
	existingTorrent := &Torrent{}
	err = collection.Find(ctx, bson.M{"hash": torrent.Hash}).One(existingTorrent)
	if err != nil {
		if err != qmgo.ErrNoSuchDocuments {
			return fmt.Errorf("error checking existing torrent: %w", err)
		}
	} else {
		fmt.Printf("❌  %s already in database\n", torrent.Magnet.InfoHash)
		// fmt.Printf("❌  %s Torrent dengan hash yang sama sudah ada dalam database\n", torrent.Magnet.InfoHash)
		// Move the .torrent file
		err = deleteTorrentFile(torrentPath)
		if err != nil {
			return fmt.Errorf("error delete torrent file: %w", err)
		}
		return nil
	}

	// Insert the data into the collection
	_, err = collection.InsertOne(ctx, &torrent)
	if err != nil {
		return fmt.Errorf("❌ error inserting into collection: %w", err)
	}

	fmt.Printf("✅  %s\n", torrent.Hash)
	// fmt.Printf("✅  %s\n", torrent.Magnet.DisplayName)

	// Move the .torrent file
	err = moveTorrentFile(torrentPath)
	if err != nil {
		return fmt.Errorf("error moving torrent file: %w", err)
	}

	return nil
}

func moveTorrentFile(torrentPath string) error {
	// Define the destination directory
	destDir := torrentDonePath
	// Extract the filename
	_, fileName := filepath.Split(torrentPath)
	destPath := filepath.Join(destDir, fileName)
	// Move the file
	err := os.Rename(torrentPath, destPath)
	if err != nil {
		return err
	}
	return nil
}

func deleteTorrentFile(filePath string) error {
    err := os.Remove(filePath)
    if err != nil {
        return fmt.Errorf("error deleting file: %w", err)
    }
    return nil
}