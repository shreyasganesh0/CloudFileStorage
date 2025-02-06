package main

import (
	"fmt"
	"net/http"
    "io"
    "os"
    "bytes"
    "mime"
    "path/filepath"
    "crypto/rand"
    "encoding/base64"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
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


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

    const MAX_MEMORY = 10 << 20;
    r.ParseMultipartForm(MAX_MEMORY);

    file, header, err_form := r.FormFile("thumbnail");
    if err_form  != nil{
        respondWithError(w, http.StatusBadRequest, "Couldnt fetch thumbnail", err);
        return;
    }

    media := header.Header.Get("Content-Type");
    media_type, _, err_mime := mime.ParseMediaType(media);
    fmt.Printf("%v\n", media_type);
    if err_mime != nil || (media_type != "image/png" &&  media_type != "image/jpg"){
        respondWithError(w, http.StatusInsufficientStorage, "Failed Creating the path", err_mime);
        return;
    }

    buf := make([]byte, 32);
    _, err_ran := rand.Read(buf)
    if err_ran != nil {
        respondWithError(w, http.StatusInsufficientStorage, "Failed Creating the path", err_ran);
        return;
    }

    rand_str := base64.RawURLEncoding.EncodeToString(buf);


    file_path := filepath.Join("assets", rand_str + "." + media_type[6:]);
    fmt.Printf("%s\n", file_path);

    fd, err_create := os.Create(file_path);
    if err_create != nil{
        respondWithError(w, http.StatusInsufficientStorage, "Failed Creating the path", err_create);
        return;
    }


    image_bytes, err_read := io.ReadAll(file);
    if err_read != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }

    image_bytes_reader := bytes.NewReader(image_bytes);
    _, err_cpy := io.Copy(fd, image_bytes_reader);
    if err_cpy != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }


    data_url := fmt.Sprintf("http://localhost:8091/assets/%s.%s", rand_str, media_type[6:]);


    video, err_query := cfg.db.GetVideo(videoID);
    if err_query != nil || video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
        return;
    }


    video.ThumbnailURL = &data_url;
    err_update := cfg.db.UpdateVideo(video);
    if err_update != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }

    respondWithJSON(w, http.StatusOK, video);
    return;

}
