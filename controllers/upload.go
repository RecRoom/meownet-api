package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"meow.net/utils"
)

type uploadFileType int

const (
	uploadFileUnknown      uploadFileType = 0
	uploadFileRoomSave     uploadFileType = 1
	uploadFileHolotar      uploadFileType = 2
	uploadFileImage        uploadFileType = 3
	uploadFileVideo        uploadFileType = 4
	uploadFileInvention    uploadFileType = 5
	uploadFileRoomMetadata uploadFileType = 6
)

func dotnetTicks(t time.Time) int64 {
	return t.UTC().UnixNano()/100 + 621355968000000000
}

func namePrefixFor(ft uploadFileType) string {
	switch ft {
	case uploadFileRoomSave:
		return "Room"
	case uploadFileRoomMetadata:
		return "RoomMetaData"
	case uploadFileHolotar:
		return "Holotar"
	case uploadFileImage:
		return "Image"
	case uploadFileVideo:
		return "Video"
	case uploadFileInvention:
		return "Invention"
	default:
		return ""
	}
}

func storageFolderFor(ft uploadFileType) string {
	switch ft {
	case uploadFileRoomSave, uploadFileRoomMetadata:
		return "room/"
	case uploadFileImage:
		return ""
	case uploadFileHolotar:
		return "data/"
	case uploadFileVideo:
		return "video/"
	case uploadFileInvention:
		return "invention/"
	default:
		return ""
	}
}

func makeUploadName(ft uploadFileType) string {
	prefix := namePrefixFor(ft)
	if prefix == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%d", prefix, uuid.New().String(), dotnetTicks(time.Now()))
}

func saveFileBytes(name string, r io.Reader, size int64, contentType string) error {
	if utils.R2Enabled() {
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		return utils.R2PutObject(name, r, size, contentType)
	}
	if err := os.MkdirAll("data/images", 0o755); err != nil {
		return err
	}
	f, err := os.Create("data/images/" + name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func LoadStoredImage(name string) ([]byte, error) {
	if utils.R2Enabled() {
		return utils.R2GetObject(name)
	}
	if data, err := os.ReadFile("data/images/" + name); err == nil {
		return data, nil
	}
	return os.ReadFile("data/" + name)
}

func SaveStoredImage(name string, data []byte, contentType string) error {
	return saveFileBytes(name, bytes.NewReader(data), int64(len(data)), contentType)
}

func DeleteStoredImage(name string) error {
	return deleteStoredImage(name)
}

func deleteStoredImage(name string) error {
	if name == "" {
		return nil
	}
	if utils.R2Enabled() {
		return utils.R2Remove(name)
	}
	err := os.Remove("data/images/" + name)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func saveUploadBytes(ft uploadFileType, name string, r io.Reader, size int64, contentType string) error {
	if ft == uploadFileUnknown {
		return fmt.Errorf("no storage folder for FileType=%d", ft)
	}
	folder := storageFolderFor(ft)
	if utils.R2Enabled() {
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		return utils.R2PutAt(folder+name, r, size, contentType)
	}
	dir := strings.TrimRight(folder, "/")
	if !strings.HasPrefix(dir, "data/") && dir != "data" {
		dir = "data/" + dir
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(dir + "/" + name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// POST /upload
func Upload(w http.ResponseWriter, r *http.Request) {
	log.Printf("[UPLOAD] ct=%q len=%d", r.Header.Get("Content-Type"), r.ContentLength)
	w.Header().Set("Content-Type", "application/json")

	rlKey := "upload_ip:" + utils.ClientIP(r)
	if accountID, ok := AccountIDFromRequest(r); ok {
		rlKey = "upload:" + strconv.FormatUint(uint64(accountID), 10)
	}
	if !utils.ActionAllowBurst(rlKey, 3*time.Second, 10) {
		log.Printf("[UPLOAD] 429 key=%s", rlKey)
		http.Error(w, "too many uploads", http.StatusTooManyRequests)
		return
	}

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		http.Error(w, "Requires multipart/form-data", http.StatusBadRequest)
		return
	}

	_, params, err := mime.ParseMediaType(ct)
	if err != nil {
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}
	mr := multipart.NewReader(r.Body, params["boundary"])

	var (
		binaryData   []byte
		binaryCT     string
		hasBinary    bool
		fileType     uploadFileType
		explicitName string
	)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[UPLOAD] multipart read error: %v", err)
			break
		}

		partName := part.FormName()
		partFile := part.FileName()
		partCT := part.Header.Get("Content-Type")
		log.Printf("[UPLOAD] part name=%q filename=%q content-type=%q", partName, partFile, partCT)

		isBinary := partFile != "" || strings.HasPrefix(partCT, "image/") || partCT == "application/octet-stream"

		if isBinary {
			data, readErr := io.ReadAll(part)
			part.Close()
			if readErr != nil {
				log.Printf("[UPLOAD] read binary error: %v", readErr)
				continue
			}
			binaryData = data
			binaryCT = partCT
			hasBinary = true
			continue
		}

		buf, _ := io.ReadAll(io.LimitReader(part, 64*1024))
		part.Close()
		text := strings.TrimSpace(string(buf))

		switch strings.ToLower(partName) {
		case "filetype":
			if v, err := strconv.Atoi(text); err == nil {
				fileType = uploadFileType(v)
			}
		case "imagename", "filename", "name":
			if explicitName == "" {
				explicitName = text
			}
		}
	}

	var fileName string
	if hasBinary {
		fileName = makeUploadName(fileType)
		if fileName == "" {
			log.Printf("[UPLOAD] missing or unknown FileType=%d", fileType)
			http.Error(w, "missing or unknown FileType", http.StatusBadRequest)
			return
		}
		if err := saveUploadBytes(fileType, fileName, bytes.NewReader(binaryData), int64(len(binaryData)), binaryCT); err != nil {
			log.Printf("[UPLOAD] save error: %v", err)
			http.Error(w, "save failed", http.StatusInternalServerError)
			return
		}
		log.Printf("[UPLOAD] saved %d bytes as %q (FileType=%d)", len(binaryData), fileName, fileType)
	} else if explicitName != "" {
		fileName = explicitName
	}

	if fileName == "" {
		http.Error(w, "missing filename or valid upload data", http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"filename": fileName,
	})
}
