package account

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"time"

	"github.com/google/uuid"
	xdraw "golang.org/x/image/draw"
)

const profileImageSize = 512

func makeProfileImageName(accountID uint) string {
	return fmt.Sprintf("profile_%d_%d_%s.png", accountID, time.Now().UTC().UnixNano(), uuid.New().String())
}

func squareProfilePNG(raw []byte, size int) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	b := src.Bounds()
	side := b.Dx()
	if b.Dy() < side {
		side = b.Dy()
	}
	ox := b.Min.X + (b.Dx()-side)/2
	oy := b.Min.Y + (b.Dy()-side)/2
	square := image.Rect(ox, oy, ox+side, oy+side)

	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, square, xdraw.Src, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
