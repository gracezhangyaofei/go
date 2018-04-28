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