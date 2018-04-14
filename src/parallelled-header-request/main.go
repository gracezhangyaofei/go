package main

import (
    "fmt"
    "net/http"
    "io"
    "bytes"
    "strings"
    "log"
    "time"
    "sync"
    "io/ioutil"
    "github.com/qiniu/api.v7/auth/qbox"
    "github.com/qiniu/api.v7/storage"

    "github.com/gin-gonic/gin"
)

var (
    accessKey = ""
    secretKey = ""
    mac = qbox.NewMac(accessKey, secretKey)
    bucket_domain = ""
    bucket = ""
)

var wg sync.WaitGroup

var client = &http.Client{
    Timeout: time.Second * 10,
}

func urlCheck(url string, i int, cvalid chan int, cinvalid chan string) {
    defer wg.Done()

    deadline := time.Now().Add(time.Second * 3600 * 2).Unix() //2小时有效期
    privateAccessURL := storage.MakePrivateURL(mac, bucket_domain, url, deadline)
    fmt.Print(".")
    resp, err := http.Get(privateAccessURL)
    defer resp.Body.Close()

    if err != nil || resp.StatusCode != 200 {
        cinvalid <- url
    } else {
        _, err1 := ioutil.ReadAll(resp.Body)
        if err1 != nil {
           cinvalid <- url
        } else {
            cvalid <- i
        }
    }
}

func start(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()  // parse arguments, you have to call this by yourself
    fmt.Println(r.Form)  // print form information in server side
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    start := time.Now()

    // this wg number should base on the urls count
    wg.Add(2)
    for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))

        urls := strings.Split(strings.Join(v, ""), ",")
        // for i, v := range urls {
        for _, v := range urls {
            fmt.Println("Start checking", v)
            // go urlCheck(v, i)
        }
    }

    for {
        iv, ok := <- cinvalid
        if !ok {
            return
        } else {
            fmt.Println("invalid value:", iv)
        }
    }

    wg.Wait()
    close(cvalid)
    close(cinvalid)

    elapsed := time.Since(start)
    log.Printf("Url check took %s", elapsed)
    fmt.Fprintf(w, "Finish checking") // send data to client side
}

func printSlice(x []string){
   fmt.Printf("len=%d cap=%d slice=%v\n",len(x),cap(x),x)
}

func putRequest(url string, data io.Reader)  {
    client := &http.Client{}
    req, err := http.NewRequest(http.MethodPut, url, data)
    if err != nil {
        // handle error
        log.Fatal(err)
    }
    _, err = client.Do(req)
    if err != nil {
        // handle error
        log.Fatal(err)
    }
}

func main() {
    r := gin.Default()
    r.POST("/task", func(c *gin.Context) {
        var cvalid = make(chan int);
        var cinvalid = make(chan string);
        id := c.Query("id")
        names := strings.Split(c.PostForm("name_keys"), ",")
        invalidList := make([]string, 4)
        for i, v := range names {
            wg.Add(1)
            go urlCheck(v, i, cvalid, cinvalid)
        }
        fmt.Printf("id: %s; urls: %s;", id, names)

        go func() {
            wg.Wait()
            close(cvalid)
            close(cinvalid)

            responseTo := "url_here"
            data := []byte(`{"urls":"` + strings.Join(invalidList, ",") + `"}`)
            request, _ := http.NewRequest(http.MethodPut, responseTo, bytes.NewBuffer(data))
            request.Header.Set("Content-Type", "application/json")
            _, err := client.Do(request)
            if err != nil {
                fmt.Println(err)
            }
        }()

        for {
            iv, ok := <- cinvalid
            if !ok {
                fmt.Println("no invalid values")
                return
            } else {
                fmt.Println("invalid value:", iv)
                invalidList = append(invalidList, iv)
            }
        }
    })
    r.Run()
}
