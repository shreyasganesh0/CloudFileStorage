package main

import (
	"fmt"
	"net/http"
    "io"
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

    media_type := header.Header.Get("Content-Type");
    
    image_bytes, err_read := io.ReadAll(file);
    if err_read != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }


    video, err_query := cfg.db.GetVideo(videoID);
    if err_query != nil || video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
        return;
    }

    tbn := thumbnail{
        data: image_bytes,
        mediaType: media_type,
    };

    videoThumbnails[videoID] = tbn;

    thumbnailURL := ("http://localhost:8090/api/thumbnails/" + videoIDString);
    video.ThumbnailURL = &thumbnailURL;
    err_update := cfg.db.UpdateVideo(video);
    if err_update != nil{
        respondWithError(w, http.StatusBadRequest, "Failed reading bytes", err_read);
        return;
    }

    respondWithJSON(w, http.StatusOK, video);
    return;

}
