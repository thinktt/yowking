import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	// get the token
	resp, err := http.Get("http://example.com")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", body)
}