package site

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/CAFxX/httpcompression"
	"github.com/delaneyj/datastar"
	"github.com/go-chi/chi/v5"
	"github.com/klauspost/compress/zstd"
)

type AsciiAnimation struct {
	Frames []string
}

func setupExamplesBadApple(examplesRouter chi.Router) error {
	anim := &AsciiAnimation{}
	gob.Register(anim)

	// Load the compressed animation
	compressedBadApple, err := staticFS.ReadFile("static/images/badapple.zst")
	if err != nil {
		return fmt.Errorf("error reading compressed bad apple: %w", err)
	}

	zstdDecoder, err := zstd.NewReader(nil)
	if err != nil {
		return fmt.Errorf("error creating zstd decoder: %w", err)
	}
	badAppleGob, err := zstdDecoder.DecodeAll(compressedBadApple, nil)
	if err != nil {
		return fmt.Errorf("error decoding compressed bad apple: %w", err)
	}
	if err := gob.NewDecoder(bytes.NewReader(badAppleGob)).Decode(anim); err != nil {
		return fmt.Errorf("error decoding bad apple: %w", err)
	}

	compress, err := httpcompression.DefaultAdapter()
	if err != nil {
		return err
	}

	examplesRouter.Route("/bad_apple/updates", func(r chi.Router) {
		r.Use(compress)

		frameCount := len(anim.Frames)

		type BadAppleStore struct {
			Contents   string  `json:"_contents"`
			Percentage float64 `json:"percentage"`
		}

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {

			t := time.NewTicker(time.Second / 30)
			sse := datastar.NewSSE(w, r)
			currentFrameIdx := 0

			store := &BadAppleStore{
				Contents:   "",
				Percentage: 0,
			}

			datastar.PatchStore(sse, store)

			for {
				select {
				case <-r.Context().Done():
					return
				case <-t.C:
					frame := anim.Frames[currentFrameIdx]
					nextFrame := (currentFrameIdx + 1) % frameCount
					store.Contents = frame
					store.Percentage = 100 * float64(nextFrame) / float64(frameCount)
					datastar.PatchStore(sse, store)
					currentFrameIdx = nextFrame
				}
			}
		})

	})

	return nil
}