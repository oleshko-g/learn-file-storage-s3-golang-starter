package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

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

	v, errGetVideo := cfg.db.GetVideo(videoID)
	if errGetVideo == nil {
		if v.CreateVideoParams.UserID != userID {
			respondWithError(w,
				http.StatusUnauthorized,
				"error: the video is owened by other user",
				errors.New("error: the video is owened by other user"))
		}

		fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

		errParseMultipartForm := r.ParseMultipartForm(10 << 20 /* 10 Megabytes */)
		if errParseMultipartForm != nil {
			respondWithError(w,
				http.StatusInternalServerError,
				errParseMultipartForm.Error(),
				errParseMultipartForm)
		}

		file, fileHeader, errFormFile := r.FormFile("thumbnail")
		if errParseMultipartForm != nil {
			respondWithError(w,
				http.StatusInternalServerError,
				errFormFile.Error(),
				errFormFile)
		}

		fileData, errReadAllFile := io.ReadAll(file)
		if errReadAllFile != nil {
			respondWithError(w,
				http.StatusInternalServerError,
				errReadAllFile.Error(),
				errReadAllFile)
		}

		videoThumbnails[videoID] = thumbnail{
			data:      fileData,
			mediaType: fileHeader.Header.Get("Content-Type"),
		}
		log.Printf("%+v", r)
		thumbnailURL := r.Header["Origin"][0] + "/api/thumbnails/" + videoIDString
		v.ThumbnailURL = &thumbnailURL

		errUpdateVideo := cfg.db.UpdateVideo(v)
		if errUpdateVideo != nil {
			respondWithError(w,
				http.StatusInternalServerError,
				errUpdateVideo.Error(),
				errUpdateVideo)
		}

		respondWithJSON(w, http.StatusCreated, v)

	} else {
		respondWithError(w,
			http.StatusInternalServerError,
			errGetVideo.Error(),
			errGetVideo)
	}

}
