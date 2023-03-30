package driver

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Lyianu/sdfs/log"
	"github.com/Lyianu/sdfs/pkg/settings"
)

func getFileDownloadLink(path, svr string) string {
	url := settings.URLSDFSScheme + svr + settings.URLSDFSDownload + "?path=" + path
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("failed to get download link: %s", err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read master response: %s", err)
		return ""
	}
	return string(b)
}
