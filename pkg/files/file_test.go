package files

import (
	"testing"
)

func TestGetMimeType(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		expected string
	}{
		{"JPGImage", "test.jpg", "image/jpeg"},
		{"JPEGImage", "test.jpeg", "image/jpeg"},
		{"PNGImage", "test.png", "image/png"},
		{"GIFImage", "test.gif", "image/gif"},
		{"PDFFile", "file.pdf", "application/pdf"},
		{"MP4Video", "video.mp4", "video/mp4"},
		{"NoExtension", "file", "application/octet-stream"},
		{"UnknownExtension", "unknown.somerndext", "application/octet-stream"},
		{"UpperCaseExtension", "image.PNG", "image/png"},
		{"MixedCaseExtension", "photo.JpG", "image/jpeg"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetMimeType(tc.fileName)
			if result != tc.expected {
				t.Errorf("GetExtension(%v) = %s; want %s", tc.fileName, result, tc.expected)
			}
		})
	}
}

func TestGetExtension(t *testing.T) {
	testCases := []struct {
		name           string
		fileName       string
		withLeadingDot bool
		expected       string
	}{
		{"LowerCaseWithLeadingDot", "test.jpg", true, ".jpg"},
		{"LowerCaseWithoutLeadingDot", "test.jpg", false, "jpg"},
		{"UpperCaseWithLeadingDot", "test.JPG", true, ".jpg"},
		{"UpperCaseWithoutLeadingDot", "test.JPG", false, "jpg"},
		{"MixedCaseWithLeadingDot", "test.JpG", true, ".jpg"},
		{"MixedCaseWithoutLeadingDot", "test.JpG", false, "jpg"},
		{"MultipleDotsWithLeadingDot", "test.WtV.JpG", true, ".jpg"},
		{"MultipleDotsWithoutLeadingDot", "test.WtV.JpG", false, "jpg"},
		{"NoExtensionWithLeadingDot", "test", true, ""},
		{"NoExtensionWithoutLeadingDot", "test", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetExtension(tc.fileName, tc.withLeadingDot)
			if result != tc.expected {
				t.Errorf("GetExtension(%v,%v) = %s; want %s", tc.fileName, tc.withLeadingDot, result, tc.expected)
			}
		})
	}
}

func TestGenerateNewFilepath(t *testing.T) {
	const a = "test/images/deals/0195d774-6fb5-ffff-73b7-ca8d51f870fe.jpg"
	var newpath = GenerateNewPathBasedOn(a, "jpg")
	t.Log(newpath)
}
