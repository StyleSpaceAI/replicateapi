package replicateapi

import (
	"encoding/base64"
	"net/http"
)

// EncodeImage into the format accepted by replicate APIs
func EncodeImage(image []byte) (string, error) {
	encoded := base64.StdEncoding.EncodeToString(image)
	contentType := http.DetectContentType(image)

	encoded = "data:" + contentType + ";base64," + encoded
	return encoded, nil
}
