package main

import(
    "fmt"
    "encoding/json"
)

func getVideoAspectRatio(filePath string) (string, error){

    type video_t struct{
        Streams []struct {
            Width int `json:"width"`
            Height int `json:"height"`
            AspectRatio string `json:"display_aspect_ratio"`
        } `json:"steams"`
    }

    
    cmd := exec.Command("ffprob", "-v", "error", "-print-format", "json", "-show_streams", filePath);
    
    var buf bytes.buffer;
    cmd.Stdout = &buf;
    
    err_run := cmd.Run();
    if err_run != nil {
        return "", err_run;
    }

    var json_ans video_t;
    decoder := json.NewDecoder(buf.Bytes());
    err_decode := decoder.Decode(&json_ans);
    if err_decoder != nil{
        return "", err_decoder;
    }

    fmt.Printf("%v\n", json_ans);
    return json_ans.Streams[0].display_aspect_ratio, nil;
}






