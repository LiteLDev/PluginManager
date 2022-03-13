package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
)

type DownloadProgressPrinter struct {
	Count    uint64
	Total    uint64
	FileName string
}

func (w *DownloadProgressPrinter) Write(p []byte) (int, error) {
	n := len(p)
	w.Count += uint64(n)
	w.RefreshProgress()
	return n, nil
}

func (w DownloadProgressPrinter) RefreshProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 40+len(w.FileName)))
	fmt.Printf("\rDownloading %s\t[%s/%s]", w.FileName, humanize.Bytes(w.Count), humanize.Bytes(w.Total))
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func DownloadFile(filepath string, url string) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &DownloadProgressPrinter{
		FileName: filepath,
		Total:    uint64(resp.ContentLength),
	}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	fmt.Print("\tDone\n")

	out.Close()

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}
func UnzipFiles(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	hasManifest := false
	for _, f := range r.File {
		_, filename := filepath.Split(f.Name)
		if !f.FileInfo().IsDir() && filename == "manifest.json" {
			hasManifest = true
			break
		}
	}
	if !hasManifest {
		return fmt.Errorf("no manifest.json found in zip file")
	}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			var dir string
			if lastIndex := strings.LastIndex(path, string(os.PathSeparator)); lastIndex > -1 {
				dir = path[:lastIndex]
			}

			err = os.MkdirAll(dir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return err
			}
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
