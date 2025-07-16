package models

import (
	"fmt"
	"github.com/vldmir/zip-service/service"
	"io"
	"log"
	"os"

	"github.com/vldmir/zip-service/util"
)

type DownloadRequest struct {
	Url         string // eg: https://cdn.videvo.net/videvo_files/video/premium/video0042/large_watermarked/900-2_900-6334-PD2_preview.mp4
	FileName    string
	Chunks      int
	Chunksize   int
	TotalSize   int
	HttpClient  *service.HTTPClient
	DownloadDir string // Добавляем поле для директории загрузки
}

func (d *DownloadRequest) SplitIntoChunks() [][2]int {
	arr := make([][2]int, d.Chunks)
	for i := 0; i < d.Chunks; i++ {
		if i == 0 {
			arr[i][0] = 0
			arr[i][1] = d.Chunksize
		} else if i == d.Chunks-1 {
			arr[i][0] = arr[i-1][1] + 1
			arr[i][1] = d.TotalSize - 1
		} else {
			arr[i][0] = arr[i-1][1] + 1
			arr[i][1] = arr[i][0] + d.Chunksize
		}
	}

	return arr
}

func (d *DownloadRequest) Download(idx int, byteChunk [2]int) error {
	log.Println(fmt.Sprintf("Downloading chunk %v", idx))
	// make GET request with range
	method := "GET"
	headers := map[string]string{
		"User-Agent": "CFD Downloader",
		"Range":      fmt.Sprintf("bytes=%v-%v", byteChunk[0], byteChunk[1]),
	}
	resp, err := d.HttpClient.Do(method, d.Url, headers)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Chunk fail %v", resp.StatusCode))
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(d.DownloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %v", err)
	}

	// Создаем временный файл в указанной директории
	tmpFilePath := fmt.Sprintf("%s/%s-%v.tmp", d.DownloadDir, util.TMP_FILE_PREFIX, idx)
	file, err := os.Create(tmpFilePath)
	if err != nil {
		return fmt.Errorf("Can't create a file %v: %v", tmpFilePath, err)
	}
	defer file.Close()

	// write to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to write to file: %v", err)
	}
	log.Println(fmt.Sprintf("Wrote chunk %v to file %s", idx, tmpFilePath))

	return nil
}

func (d *DownloadRequest) MergeDownloads() error {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(d.DownloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %v", err)
	}

	// Создаем итоговый файл в указанной директории
	outputFilePath := fmt.Sprintf("%s/%s", d.DownloadDir, d.FileName)
	out, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %v", outputFilePath, err)
	}
	defer out.Close()

	// Объединяем все чанки
	for idx := 0; idx < d.Chunks; idx++ {
		tmpFilePath := fmt.Sprintf("%s/%s-%v.tmp", d.DownloadDir, util.TMP_FILE_PREFIX, idx)
		in, err := os.Open(tmpFilePath)
		if err != nil {
			return fmt.Errorf("Failed to open chunk file %s: %v", tmpFilePath, err)
		}

		_, err = io.Copy(out, in)
		in.Close() // Закрываем файл сразу после использования
		if err != nil {
			return fmt.Errorf("Failed to merge chunk file %s: %v", tmpFilePath, err)
		}
	}

	log.Printf("File chunks merged successfully to %s", outputFilePath)
	return nil
}

func (d *DownloadRequest) CleanupTmpFiles() error {
	log.Println("Starting to clean tmp downloaded files...")

	// Удаляем все временные файлы
	for idx := 0; idx < d.Chunks; idx++ {
		tmpFilePath := fmt.Sprintf("%s/%s-%v.tmp", d.DownloadDir, util.TMP_FILE_PREFIX, idx)
		err := os.Remove(tmpFilePath)
		if err != nil {
			// Продолжаем удалять другие файлы даже если один не удалился
			log.Printf("Failed to remove chunk file %s: %v", tmpFilePath, err)
		}
	}

	return nil
}
