package files

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/uid"
	"github.com/pixie-sh/core-go/pkg/utils"
)

func FilenameFromFileExternalURL(externalURL string, desireFilename string) (string, string, error) {
	fileUrl, err := url.Parse(externalURL)
	if err != nil {
		return "", "", err
	}

	filePath := fileUrl.Path
	segments := strings.Split(filePath, "/")

	extension := path.Ext(segments[len(segments)-1])
	imageName := desireFilename + extension
	return imageName, extension, nil
}

func DownloadFile(URL string, handler func(body io.Reader) error) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	if response.StatusCode > 299 {
		content, _ := utils.ReadCloserToString(response.Body)
		return errors.New("error downloading file: %s", content)
	}

	return handler(response.Body)
}

func FormatImageName(basePath string, imageName string, extension string, imageType ImageType) (string, string) {
	if basePath[len(basePath)-1] != '/' {
		basePath += "/"
	}

	switch imageType {
	case ImageTypeEvents:
		return basePath + imageName, strings.ToLower(extension)
	case ImageTypeLocations:
		return basePath + imageName, strings.ToLower(extension)
	case ImageTypeProfile:
		return basePath + imageName, strings.ToLower(extension)
	default:
		return "lost/" + imageName, strings.ToLower(extension)
	}
}

func ImageBlobFromReader(imageContent io.Reader) ([]byte, string, error) {
	decode, formatString, err := image.Decode(imageContent)
	if err != nil {
		return nil, "", err
	}

	buff := new(bytes.Buffer)
	if formatString == "jpeg" {
		err = jpeg.Encode(buff, decode, &jpeg.Options{Quality: 75})
		if err != nil {
			return nil, "", err
		}
	} else if formatString == "png" {
		err = png.Encode(buff, decode)
		if err != nil {
			return nil, "", err
		}
	}

	return buff.Bytes(), formatString, nil
}

func ResizeImage(imageFile io.Reader, width, height int) (*bytes.Reader, string, error) {
	img, _, err := image.Decode(imageFile)
	if err != nil {
		return nil, "", err
	}

	resize := imaging.Resize(img, width, height, imaging.Lanczos)
	buff := new(bytes.Buffer)
	err = png.Encode(buff, resize)
	if err != nil {
		return nil, "", err
	}

	readers := bytes.NewReader(buff.Bytes())
	return readers, "png", nil
}

func ResizeImageFromBytes(input []byte, width, height int) ([]byte, error) {
	src, formatDetected, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, errors.NewWithError(err, "failed to decode image")
	}

	resized := imaging.Resize(src, width, height, imaging.Lanczos)

	var buffer bytes.Buffer
	var encodeErr error

	switch formatDetected {
	case "jpeg":
		encodeErr = jpeg.Encode(&buffer, resized, nil)
		break
	case "jpg":
		encodeErr = jpeg.Encode(&buffer, resized, nil)
		break
	case "png":
		encodeErr = png.Encode(&buffer, resized)
		break
	case "gif":
		encodeErr = gif.Encode(&buffer, resized, nil)
		break
	default:
		return nil, errors.New("unsupported image format: %s", formatDetected)
	}

	if encodeErr != nil {
		return nil, errors.NewWithError(encodeErr, "failed to encode resized image")
	}

	return buffer.Bytes(), nil
}

func GenerateNewPathBasedOn(path string, extension string) string {
	newFilename := fmt.Sprintf("%s.%s", uid.NewUUID(), extension)
	dir := filepath.Dir(path)
	return filepath.Join(dir, newFilename)
}
