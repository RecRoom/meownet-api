package importer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"meow.net/utils"
)

const (
	importerOwnerAccountID   = 24
	importerRoomMetadataBlob = "RoomMetaDataf777cce2-40b5-4673-acf3-97d91416a4d0639135231157057262"
)

func dotnetTicks(t time.Time) int64 {
	return t.UTC().UnixNano()/100 + 621355968000000000
}

func makeRoomBlobName() string {
	return fmt.Sprintf("RoomData%s%d", uuid.New().String(), dotnetTicks(time.Now()))
}

func makeImageName(originalExt string) string {
	if originalExt == "" {
		originalExt = ".jpg"
	}
	return fmt.Sprintf("Image%s%d%s", uuid.New().String(), dotnetTicks(time.Now()), originalExt)
}

func storeRoomBlob(name string, data []byte) error {
	return storeBytes("room/"+name, data, "application/octet-stream", "data/room/"+name)
}

func storeImage(name string, data []byte, contentType string) error {
	return storeBytes(name, data, contentType, "data/images/"+name)
}

func storeHolotar(name string, data []byte) error {
	return storeBytes("data/"+name, data, "application/octet-stream", "data/data/"+name)
}

func storeBytes(r2Key string, data []byte, contentType, fallbackPath string) error {
	if utils.R2Enabled() {
		return utils.R2PutAt(r2Key, bytes.NewReader(data), int64(len(data)), contentType)
	}
	dir := fallbackPath[:strings.LastIndex(fallbackPath, "/")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(fallbackPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewReader(data))
	return err
}
