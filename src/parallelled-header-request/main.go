package main

import (
    "fmt"
    "net/http"
    // "io"
    "bytes"
    "strings"
    // "log"
    "time"
    "sync"
    "io/ioutil"
    "github.com/qiniu/api.v7/auth/qbox"
    "github.com/qiniu/api.v7/storage"

    "github.com/gin-gonic/gin"
    // "encoding/json"
    // "strconv"
)

var (
    accessKey = ""
    secretKey = ""
    mac = qbox.NewMac(accessKey, secretKey)
    bucket_domain = ""
    bucket = ""
)

var wg sync.WaitGroup
var mutex sync.Mutex

var client = &http.Client{
    Timeout: time.Second * 10,
}

func urlCheck(url string, i int, cinvalid chan string) {
    defer wg.Done()

    deadline := time.Now().Add(time.Second * 3600 * 2).Unix() //2小时有效期
    privateAccessURL := storage.MakePrivateURL(mac, bucket_domain, url, deadline)

    resp, err := http.Get(privateAccessURL)
    defer resp.Body.Close()

    if err != nil || resp.StatusCode != 200 {
        fmt.Println("X:",resp.StatusCode)
        cinvalid <- url
    } else {
        _, err1 := ioutil.ReadAll(resp.Body)
        if err1 != nil {
            fmt.Print("X")
            cinvalid <- url
        } else {
            fmt.Print("√")
        }
    }
}

func main() {
    r := gin.Default()
    r.POST("/task", func(c *gin.Context) {
        var cinvalid = make(chan string);

        id := c.Query("id")
        name_keys := strings.Split(c.PostForm("name_keys"), ",")
        fmt.Printf("id: %s; urls: %s;\n", id, name_keys)
        length_of_name_keys := len(name_keys)

        var invalidList = make([]string, length_of_name_keys)
        for i, v := range name_keys {
            wg.Add(1)
            go urlCheck(v, i, cinvalid)
        }
        
        go func() {
            for {
                iv, ok := <- cinvalid
                if !ok {
                    fmt.Println("no invalid values")
                    close(cinvalid)
                    wg.Done()
                    break
                } else {
                    fmt.Println("invalid value:", iv)
                    invalidList = append(invalidList, iv)
                }
            }
        }()

        wg.Wait()
        
        fmt.Println("Finished here!")
        fmt.Println(invalidList)

        responseTo := ""
        data := []byte(`{"urls":"` + strings.Join(invalidList, ",") + `"}`)
        request, _ := http.NewRequest(http.MethodPut, responseTo, bytes.NewBuffer(data))
        request.Header.Set("Content-Type", "application/json")
        _, err := client.Do(request)
        if err != nil {
            fmt.Println(err)
        }
    })
    r.Run()
}
