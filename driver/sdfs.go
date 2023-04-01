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
	//fmt.Println(url)
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

func (f *File) flushBuffer() (int64, error) {
	fmt.Printf("flushing buffer: %d\n", f.f_offset)
	req, err := http.NewRequest("GET", f.location, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", f.f_offset, f.f_offset+f.bufSize-1))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get file from server: %s", resp.Status)
	}
	n, err := io.ReadFull(resp.Body, f.buf)
	if err != nil {
		return 0, err
	}
	f.f_pos = 0
	return int64(n), nil
}
