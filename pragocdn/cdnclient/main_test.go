package cdnclient

import (
	"os"
	"testing"
)

func TestCDN(t *testing.T) {
	account := NewCDNAccount("http://localhost:8587", "test", "testpassword")

	resp, err := account.UploadFileFromPath(os.Getenv("HOME") + "/.pragocdn/testdata/img.jpg")
	if err != nil {
		t.Fatal(err)
	}

	err = account.DeleteFile(resp.UUID)
	if err != nil {
		t.Fatal(err)
	}

}
