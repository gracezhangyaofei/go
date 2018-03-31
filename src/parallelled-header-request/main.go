package main

import (
    "fmt"
    "net/http"
    "strings"
    "log"
    "time"
    "sync"
    "io/ioutil"
)

var cvalid = make(chan int);
var cinvalid = make(chan string);
var wg sync.WaitGroup

func urlCheck(url string, i int) {
    defer wg.Done()

    // resp, err := client.Get(url)
    // here got a problem, it's always return 403 but not 200, please double check
    resp, err := http.Get(url)
    log.Println(resp.StatusCode)
    log.Println(err)
    defer resp.Body.Close()

    if err != nil || resp.StatusCode != 200 {
        cinvalid <- url
    } else {
        _, err1 := ioutil.ReadAll(resp.Body)
        log.Println(err1)
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
        for i, v := range urls {
            fmt.Println("Start checking", v)
            go urlCheck(v, i)
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

func main() {
    http.HandleFunc("/", start) // set router
    err := http.ListenAndServe(":9090", nil) // set listen port
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
