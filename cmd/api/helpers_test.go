package main

import (
	"github.com/minhnghia2k3/greenlight/internal/data"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkWriteJSON(b *testing.B) {
	app := &application{}
	movie := data.Movie{
		ID:        int64(1),
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()

		err := app.writeJSON(rr, http.StatusOK, envelop{"movie": movie}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

//func BenchmarkWriteJSONIndent(b *testing.B) {
//	app := &application{}
//	movie := data.Movie{
//		ID:        1,
//		CreatedAt: time.Now(),
//		Title:     "Casablanca",
//		Runtime:   102,
//		Genres:    []string{"drama", "romance", "war"},
//		Version:   1,
//	}
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		rr := httptest.NewRecorder()
//
//		err := app.writeJSONIndent(rr, http.StatusOK, movie, nil)
//		if err != nil {
//			b.Fatal(err)
//		}
//	}
//}
