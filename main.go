package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	url = ""
)

const (
	defaultBufferSize = int64(10000000)
)

func download(url string, from, to int64) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "identity;q=1, *;q=0")
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", from, to))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func printDownloadProgress(downloaded, total int64) {
	fmt.Printf("File downloaded: %.2f %s\n", float64(downloaded)/float64(total)*100, "%.")
}

// this help in resume file download.
func getFileStat(fileName string) (int64, error) {
	if f, err := os.Stat(fileName); err == nil {
		return f.Size(), nil
	} else if os.IsNotExist(err) {
		return 0, err
	} else {
		// #ToDo https://stackoverflow.com/a/12518877/4408364
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return -1, err
	}
}

func downloadAndSaveFile(url string, fileSize int64, fileType string) error {
	fileName := path.Base(url)
	var fromLen int64
	var toLen int64
	var err error
	fromLen, err = getFileStat(fileName)
	if err != nil && fromLen == -1 {
		return err
	}
	toLen = fromLen + defaultBufferSize
	if toLen > fileSize {
		toLen = fileSize
	}
	for toLen < fileSize {
		content, err := download(url, fromLen, toLen)
		if err != nil {
			return err
		}
		if err := write(fileName, content); err != nil {
			return fmt.Errorf("Can not save file: %s\nStopping download", err)
		}
		// go write(fileName, content)
		printDownloadProgress(toLen, fileSize)
		fromLen = toLen + 1
		toLen = fromLen + defaultBufferSize
		if toLen > fileSize {
			toLen = fileSize
		}
	}
	fmt.Printf("Download completed. File saved at: %s\n", fileName)
	return nil
}

func write(fileName string, content []byte) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	return err
}

func getFileInfo(url string) (int64, string, error) {
	qParam := "mime=true"
	url = fmt.Sprintf("%s?%s", url, qParam)
	resp, err := http.Head(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	len := resp.Header.Get("content-length")
	lenInt, err := strconv.Atoi(len)
	if err != nil {
		return 0, "", err
	}
	fileType := resp.Header.Get("content-type")
	// #ToDo handle array index out of bound error.
	fileType = strings.Split(fileType, "/")[1]
	return int64(lenInt), fileType, nil
}

func main() {
	fmt.Println("Enter url")
	var err error
	if url, err = bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		fmt.Printf("Error reading url: %s", err)
		return
	}
	fileLen, fileType, err := getFileInfo(url)
	if err != nil {
		fmt.Printf("Error getting file information: %s\n", err)
		return
	}
	fmt.Printf("File info Length: %d Type: %s\n", fileLen, fileType)
	log.Fatal(downloadAndSaveFile(url, fileLen, fileType))
}
