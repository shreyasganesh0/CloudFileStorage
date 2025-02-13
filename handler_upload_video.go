package main

import (
	"fmt"
	"net/http"
    "io"
    "os"
    "bytes"
    "mime"
    "crypto/rand"
    "encoding/base64"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
    
    UPLOAD_LIMIT := int64(1 << 30); // 1GB

    r.Body = http.MaxBytesReader(w, r.Body, UPLOAD_LIMIT);
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return;
    }


	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	
    video, err_query := cfg.db.GetVideo(videoID);
    if err_query != nil || video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
        return;
    }


    r.ParseMultipartForm(UPLOAD_LIMIT);
    file, header, err_form := r.FormFile("video");
    if err_form  != nil{
        respondWithError(w, http.StatusBadRequest, "Couldnt fetch thumbnail", err);
        return;
    }
    defer file.Close();

    
    media := header.Header.Get("Content-Type");
    media_type, _, err_mime := mime.ParseMediaType(media);
    fmt.Printf("%v\n", media_type);
    if err_mime != nil || media_type != "video/mp4"{
        respondWithError(w, http.StatusInsufficientStorage, "Failed Creating the path", err_mime);
        return;
    }

    fd, err_f := os.CreateTemp("", "s3_temp_vid.mp4");
    if err_f != nil {
        respondWithError(w, http.StatusInsufficientStorage, "Failed Creating the path", err_f);
        return;
    }

    defer os.Remove(fd.Name());
    defer fd.Close();

    
    video_bytes, err_read := io.ReadAll(file);
    if err_read != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }

    video_bytes_reader := bytes.NewReader(video_bytes);
    _, err_cpy := io.Copy(fd, video_bytes_reader);
    if err_cpy != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }

    aspect_ratio, err_ratio := getVideoAspectRatio(fd.Name())
    if err_ratio != nil {
        fmt.Printf("Failed to get the aspect ratio %v\n", err_ratio)
    }

    var prefix string
    if aspect_ratio == "16:9" {
        prefix = "landscape/";
    } else if aspect_ratio == "9:16" {
        prefix = "portrait/";
    } else {
        prefix = "other/";
    }
    new_path, err_fast := processVideoForFastStart(fd.Name())
    if err_fast!= nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_fast);
        return;
    }
    new_fd, err_new := os.OpenFile(new_path, os.O_RDWR|os.O_CREATE, 0777);
    if err_new != nil {
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_new);
        return;
    }

    defer os.Remove(new_fd.Name());
    defer new_fd.Close();


    _, err_seek := new_fd.Seek(0, io.SeekStart);
    if err_seek != nil{
        respondWithError(w, http.StatusBadRequest, "Failed seeking bytes", err_seek);
        return;
    }

    buf := make([]byte, 32);
    _, err_ran := rand.Read(buf)
    if err_ran != nil {
        respondWithError(w, http.StatusInsufficientStorage, "Failed Creating the path", err_ran);
        return;
    }
    rand_str := base64.RawURLEncoding.EncodeToString(buf);
    url := prefix + rand_str + "." + media_type[6:];
    fmt.Printf("url is %v\n", url);

    fmt.Printf("%v\n", cfg);
    put_obj := s3.PutObjectInput{
        Bucket: &cfg.s3Bucket,
        Key: &url,
        Body: new_fd,
        ContentType: &media_type,
    };
    
    _, err_put := cfg.s3Client.PutObject(r.Context(), &put_obj);
    if err_put != nil{
        respondWithError(w, http.StatusInsufficientStorage, "Failed uploading file path", err_put);
        return;
    }

    VideoURL := fmt.Sprintf("https://%v.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, url);
    fmt.Printf("%v\n", VideoURL);
    video.VideoURL = &VideoURL;

    err_update := cfg.db.UpdateVideo(video);
    if err_update != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }

    respondWithJSON(w, http.StatusOK, video);
    return;
    
}
