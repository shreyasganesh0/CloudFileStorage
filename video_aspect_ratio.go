package main

import(
    "fmt"
    "encoding/json"
    "os/exec"
    "bytes"
)

func getVideoAspectRatio(filePath string) (string, error){

    type video_t struct{
        Streams []struct {
            Width int `json:"width"`
            Height int `json:"height"`
            AspectRatio string `json:"display_aspect_ratio"`
        } `json:"streams"`
    }

    
    fmt.Printf("Path is: %v\n", filePath);
    cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath);
    
    var buf bytes.Buffer;
    cmd.Stdout = &buf;
    
    err_run := cmd.Run();
    if err_run != nil {
        return "", err_run;
    }

    var json_ans video_t;
    err_decode := json.Unmarshal(buf.Bytes(), &json_ans);
    if err_decode != nil{
        return "", err_decode;
    }

    fmt.Printf("%v\n", json_ans);
    return json_ans.Streams[0].AspectRatio, nil;
}

func processVideoForFastStart(filePath string) (string, error) {
     
    cmd := exec.Command("ffmpeg", "-i" , filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", filePath + ".processing");
    
    err_run := cmd.Run();
    if err_run != nil {
        return "", err_run;
    }

    return filePath + ".processing", nil;
}





