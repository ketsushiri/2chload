package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Unit struct {
	board, thread string
}

func (u Unit) String() string {
	return u.board + "/" + u.thread
}

type File struct {
	Name string
	Path string
}

type Post struct {
	Files []File
}

type Thread struct {
	Posts []Post
}

type Api struct {
	Board   string
	Threads []Thread
}

const (
	baseUrl = "https://2ch.hk/"
	bitmask = 0750
	pics    = "pics"
	vids    = "vids"
	help    = "Usage: 2chload [board/thread]\n\te.g: 2chload b/23242553 pr/543323 math/235114"
	BUFSIZE = 2048
)

func getApiResponse(board string, thread string) *Api {
	var response Api
	url := baseUrl + board + "/res/" + thread + ".json"

	resp, err := http.Get(url)
	if err != nil {
		log.Println(Unit{board, thread}, err)
		return nil
	}
	defer resp.Body.Close()
	cont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(Unit{board, thread}, err)
		return nil
	}
	json.Unmarshal(cont, &response)
	if len(response.Threads) == 0 {
		log.Println(Unit{board, thread}, "не удалось получить тред.")
		return nil
	}
	return &response
}

func fetchFiles(board string, thread string) (files []string) {
	response := getApiResponse(board, thread)
	if response == nil {
		return files
	}
	for _, post := range response.Threads[0].Posts {
		for _, file := range post.Files {
			files = append(files, file.Path)
		}
	}
	return files
}

func download(file string, ch chan<- string) {
	start := time.Now()
	url := baseUrl + file[1:]
	meta := strings.Split(file[1:], "/")
	if meta[0] == "stickers" {
		ch <- fmt.Sprintf("%.2fs %10d найден стикер, игнорирую...\n", time.Since(start).Seconds(), 0)
		return
	}
	board, thread, name := meta[0], meta[2], meta[3] // [1] for /src/

	path := thread + "/" + pics
	if strings.HasSuffix(name, ".webm") || strings.HasSuffix(name, ".mp4") {
		path = thread + "/" + vids
	}

	_ = os.MkdirAll(path, bitmask) // ignoring
	filename := path + "/" + name

	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprintf("%s/%s/%s: %v\n", board, thread, name, err)
		return
	}
	defer resp.Body.Close()
	cont, err := os.Create(filename)
	if err != nil {
		ch <- fmt.Sprintf("%s/%s/%s: %v\n", board, thread, name, err)
		return
	}
	defer cont.Close()
	size, err := io.Copy(cont, resp.Body)
	secs := time.Since(start).Seconds()

	ch <- fmt.Sprintf("%.2fs %10d %s/%s/%s\n", secs, size, board, thread, name)
}

func getUnits(args []string) (units []Unit) {
	for _, arg := range args {
		result := strings.Split(arg, "/")
		if len(result) == 2 {
			units = append(units, Unit{result[0], result[1]})
		}
	}
	return units
}

func main() {
	log.SetFlags(log.Ltime)
	units := getUnits(os.Args[1:])
	if len(units) == 0 {
		fmt.Println(help)
		os.Exit(2)
	}
	var files []string
	for _, unit := range units {
		log.Println(unit, "ищем файлы...")
		local := fetchFiles(unit.board, unit.thread)
		log.Println(unit, "найдено", len(local))
		files = append(files, local...)
	}
	log.Println("всего будет скачано:", len(files))
	start := time.Now()
	ch := make(chan string, BUFSIZE)
	for _, file := range files {
		go download(file, ch)
	}
	for range files {
		fmt.Print(<-ch)
	}
	fmt.Printf("%.2fs завершено.\n", time.Since(start).Seconds())
}
