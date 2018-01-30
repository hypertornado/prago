package cdnclient

import (
	"fmt"
	"os"
	"testing"
)

func TestCDN(t *testing.T) {
	account := NewCDNAccount("http://localhost:8585", "test", "testpassword")

	resp, err := account.UploadFileFromPath(os.Getenv("HOME") + "/.pragocdn/testdata/img.jpg")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp)
}
