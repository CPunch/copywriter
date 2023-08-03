package replicate

import (
	"fmt"
	"os"
	"testing"
)

func TestGenerateImage(t *testing.T) {
	client := NewClient(os.Getenv("REPLICATE_API_KEY"))
	url, err := client.MakePrediction("a vision of paradise. unreal engine")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(url)
}
