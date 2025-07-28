package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	if v, errGetVideo := cfg.db.GetVideo(videoID); errGetVideo == nil {
		if v.CreateVideoParams.UserID != userID {
			respondWithError(w,
				http.StatusUnauthorized,
				"the video is owened by other user",
				errors.New("the video is owened by other user"))
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

		mediaType, _, errParseMediaType := mime.ParseMediaType(fileHeader.Header.Get("Content-Type"))
		if errParseMediaType != nil {
			respondWithError(w,
				http.StatusInternalServerError,
				errParseMediaType.Error(),
				errParseMediaType)
		}

		if mediaType == "image/png" || mediaType == "image/jpeg" {
			//saveAsset
			fileExtension, _ := strings.CutPrefix(mediaType, "image/")
			randomData := make([]byte, 32)
			rand.Read(randomData)
			thumbnailFilePath := filepath.Join(cfg.assetsRoot, base64.RawURLEncoding.EncodeToString(randomData)) + "." + fileExtension
			fileOnDisk, saveFileToDisk := os.Create(thumbnailFilePath)
			io.Copy(fileOnDisk, file)
			if saveFileToDisk != nil {
				respondWithError(w,
					http.StatusInternalServerError,
					saveFileToDisk.Error(),
					saveFileToDisk)
			}
			thumbnailURL := r.Header.Get("Origin") + "/" + thumbnailFilePath
			v.ThumbnailURL = &thumbnailURL

			if errUpdateVideo := cfg.db.UpdateVideo(v); errUpdateVideo != nil {
				respondWithError(w,
					http.StatusInternalServerError,
					errUpdateVideo.Error(),
					errUpdateVideo)
			}
			respondWithJSON(w, http.StatusCreated, v)
		} else {
			respondWithError(w,
				http.StatusBadRequest,
				"not an image",
				errors.New("not an image"))
		}

	} else if errGetVideo != nil {
		respondWithError(w,
			http.StatusInternalServerError,
			errGetVideo.Error(),
			errGetVideo)
	}

}
