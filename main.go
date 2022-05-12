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
    "log"
)

const (
    BUFSIZE = 2048
)

type Unit struct {
    board, thread string
}

func (u Unit) String() string {
    return u.board + "/" + u.thread
}

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
    help    = "Usage: 2chload [board/thread]\n\te.g: 2chload b/23242553 pr/543323 math/235114"
)

func createDir(name, subdir string) (path string, err error) {
    path = fmt.Sprintf("%s/%s", name, subdir)
    err = os.MkdirAll(path, bitmask)
    if err != nil {
        return "", err
    }
    return path, nil
}

func getApiResponse(board string, thread string) *Api {
    var response Api
    url := baseUrl + board + "/res/" + thread + ".json"

    resp, err := http.Get(url)
    if err != nil {
        log.Println(board, thread, err)
        return nil
    }
    defer resp.Body.Close()
    cont, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println(board, thread, err)
        return nil
    }
    json.Unmarshal(cont, &response)
    return &response
}

func fetchFiles(board string, thread string) []string {
    var files []string
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

func getUnits(args []string) []Unit {
    var units []Unit
    for _, arg := range args {
        result := strings.Split(arg, "/")
        if len(result) == 2 {
            units = append(units, Unit{result[0], result[1]})
        }
    }
    return units
}

func main() {
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
