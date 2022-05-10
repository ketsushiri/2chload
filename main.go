package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type File struct {
	Displayname string
	Fullname    string
	Height      int32
	Md5         string
	Name        string
	Nsfw        int32
	Path        string
	Thumbnail   string
	Tn_height   int32
	Tn_width    int32
	Type        int32
	Width       int32
}

type Post struct {
	Banned    int32
	Closed    int32
	Comment   string
	Date      string
	Email     string
	Endless   int32
	Files     []File
	Lasthit   int64
	Name      string
	Num       int64
	Number    int32
	Op        int32
	Parent    string
	Sticky    int32
	Subject   string
	Timestamp int64
	Trip      string
}

type Thread struct {
	Posts []Post
}

type News struct {
	Date    string
	Num     int64
	Subject string
	Views   int64
}

type Meta struct {
	Board string
	Info  string
	Name  string
}

type Api struct {
	Board               string
	BoardInfo           string
	BoardInfoOuter      string
	BoardName           string
	Advert_bottom_image string
	Advert_bottom_link  string
	Advert_mobile_image string
	Advert_mobile_link  string
	Advert_top_image    string
	Advert_top_link     string
	Board_banner_image  string
	Board_banner_link   string
	Bump_limit          int32
	Current_thread      string
	Default_name        string
	Enable_dices        int8
	Enable_flags        int8
	Enable_icons        int8
	Enable_images       int8
	Enable_likes        int8
	Enable_names        int8
	Enable_oekaki       int8
	Enable_posting      int8
	Enable_sage         int8
	Enable_shield       int8
	Enable_subject      int8
	Enable_thread_tags  int8
	Enable_trips        int8
	Enable_video        int8
	Files_count         int8
	Is_board            int8
	Is_closed           int8
	Is_index            int8
	Max_comment         int32
	Max_files_size      int32
	Max_num             int64
	News_abu            []News
	Posts_count         int32
	Thread_first_image  string
	Threads             []Thread
	Tiltle              string
	Top                 []Meta
	Unique_posters      string
}

const (
	baseUrl = "https://2ch.hk/"
	bitmask = 0750
	pics    = "pics"
	vids    = "vids"
)

var counter int32 // Im retarded a little bit.

func createDir(name, subdir string) (path string, err error) {
	path = fmt.Sprintf("%s/%s", name, subdir)
	err = os.MkdirAll(path, bitmask)
	if err != nil {
		return "", err
	}
	return path, nil
}

func saveContent(thread string, response *Api, ch chan<- string) error {
	start := time.Now()
	picsPath, err := createDir(thread, pics)
	if err != nil {
		return err
	}
	vidsPath, err := createDir(thread, vids)
	if err != nil {
		return err
	}
	for _, post := range response.Threads[0].Posts {
		for _, file := range post.Files {
			path := baseUrl + file.Path
			var finalPath string
			if strings.HasSuffix(file.Name, ".webm") || strings.HasSuffix(file.Name, ".mp4") {
				finalPath = vidsPath
			} else {
				finalPath = picsPath
			}
			filename := finalPath + "/" + file.Name
			content, err := http.Get(path)
			if err != nil {
				return err
			}
			data, err := os.Create(filename)
			if err != nil {
				return err
			}
			size, err := io.Copy(data, content.Body)
			content.Body.Close()
			data.Close()

			secs := time.Since(start).Seconds()
			ch <- fmt.Sprintf("%.2fs %7d\t%s/%s/%s\n", secs, size, response.Board, thread, file.Name)
		}
	}
	return nil
}

func fetchContent(board, thread string, ch chan<- string) {
	var response Api
	apiUrl := baseUrl + board + "/res/" + thread + ".json"
	defer func() { counter++ }()

	resp, err := http.Get(apiUrl)
	if err != nil || resp.StatusCode != 200 {
		ch <- fmt.Sprintf("Request to the 2ch's api has failed.\n")
		return
	}
	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- fmt.Sprintf("%v\n", err)
		return
	}
	resp.Body.Close()
	json.Unmarshal(respB, &response)

	err = saveContent(thread, &response, ch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "2chload: %v\n", err)
	}
}

func main() {
	start := time.Now()
	ch := make(chan string)
	for _, arg := range os.Args[1:] {
		pair := strings.Split(arg, "/")
		if len(pair) < 2 {
			counter++
			continue
		}
		go fetchContent(pair[0], pair[1], ch)
	}
	for counter < int32(len(os.Args[1:])) {
		fmt.Printf(<-ch)
	}
	fmt.Printf("%.2fs finished.\n", time.Since(start).Seconds())
}
