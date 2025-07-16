package manager

import (
	"fmt"
	"github.com/vldmir/zip-service/service"
	"github.com/vldmir/zip-service/util"
	"log"
	"net/url"
	"strconv"
	"sync"

	"github.com/vldmir/zip-service/models"
)

func Init() {
	fmt.Println(util.InitMotd)
}

func End() {
	fmt.Println(util.EndMotd)
}

func Run(urls []*url.URL, downloadDir string) {
	// Инициализация HTTP клиента один раз для всех загрузок
	client := service.NewHTTPClient()

	for _, urlPtr := range urls {
		urlStr := urlPtr.String()
		log.Printf("\n=== Processing URL: %s ===\n", urlStr)

		// make HEAD call
		method := "HEAD"
		headers := map[string]string{
			"User-Agent": "CFD Downloader",
		}
		resp, err := client.Do(method, urlStr, headers)
		if err != nil {
			log.Printf("HEAD request failed for %s: %v", urlStr, err)
			continue
		}

		// get Content-Length
		contentLength := resp.Header.Get(util.CONTENT_LENGTH_HEADER)
		contentLengthInBytes, err := strconv.Atoi(contentLength)
		if err != nil {
			log.Printf("Unsupported file download type for %s: %v", urlStr, err)
			continue
		}
		log.Println("Content-Length:", contentLengthInBytes)

		// get file name
		fname, err := util.ExtractFileName(urlStr)
		if err != nil {
			log.Printf("Error extracting filename from %s: %v", urlStr, err)
			continue
		}
		log.Println("Filename extracted: ", fname)

		// set concurrent workers
		chunks := util.WORKER_ROUTINES
		log.Printf("Set %v parallel workers/connections", chunks)

		// calculate chunk size
		chunksize := contentLengthInBytes / chunks
		log.Println("Each chunk size: ", chunksize)

		// create the downloadRequest object
		downReq := &models.DownloadRequest{
			Url:         urlStr,
			FileName:    fname,
			Chunks:      chunks,
			Chunksize:   chunksize,
			TotalSize:   contentLengthInBytes,
			HttpClient:  client,
			DownloadDir: downloadDir, // Устанавливаем директорию загрузки
		}

		// chunk it up
		byteRangeArray := make([][2]int, chunks)
		byteRangeArray = downReq.SplitIntoChunks()
		fmt.Println(byteRangeArray)

		// download each chunk concurrently
		var wg sync.WaitGroup
		for idx, byteChunk := range byteRangeArray {
			wg.Add(1)

			go func(idx int, byteChunk [2]int) {
				defer wg.Done()
				err := downReq.Download(idx, byteChunk)
				if err != nil {
					log.Printf("Failed to download chunk %v from %s: %v", idx, urlStr, err)
				}
			}(idx, byteChunk)
		}
		wg.Wait()

		// merge
		err = downReq.MergeDownloads()
		if err != nil {
			log.Printf("Failed merging tmp downloaded files for %s: %v", urlStr, err)
			continue
		}

		// cleanup
		err = downReq.CleanupTmpFiles()
		if err != nil {
			log.Printf("Failed cleaning up tmp downloaded files for %s: %v", urlStr, err)
		}

		// final file generated
		log.Printf("File successfully downloaded: %v\n", downReq.FileName)
	}
}
