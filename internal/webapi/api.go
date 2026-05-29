package webapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/veirt/vivere/internal/masterdata"
)

const defaultLanguageBook = "TextResourceJaJpMB"

type API struct {
	master      masterdata.Loader
	exportsRoot string
}

type apiError struct {
	Error string `json:"error"`
}

func NewAPI(master masterdata.Loader, exportsRoot ...string) API {
	api := API{master: master}
	if len(exportsRoot) > 0 {
		api.exportsRoot = exportsRoot[0]
	}
	return api
}

func (a API) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/characters", a.handleCharacters)
	mux.HandleFunc("GET /api/characters/{id}", a.handleCharacter)
	mux.HandleFunc("GET /api/master-books", a.handleBooks)
	mux.HandleFunc("GET /api/master-books/{book}", a.handleBookRows)
}

func (a API) handleCharacters(w http.ResponseWriter, r *http.Request) {
	lang := strings.TrimSpace(r.URL.Query().Get("lang"))
	if lang == "" {
		lang = defaultLanguageBook
	}
	if err := masterdata.ValidateBookName(lang); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	characters, err := a.master.LoadCharacters(lang)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.enrichCharacterAudio(characters)
	writeJSON(w, http.StatusOK, characters)
}

func (a API) handleCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("character id must be a positive integer"))
		return
	}
	lang := strings.TrimSpace(r.URL.Query().Get("lang"))
	if lang == "" {
		lang = defaultLanguageBook
	}
	if err := masterdata.ValidateBookName(lang); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	characters, err := a.master.LoadCharacters(lang)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.enrichCharacterAudio(characters)
	for _, character := range characters {
		if character.ID == id {
			writeJSON(w, http.StatusOK, character)
			return
		}
	}
	writeError(w, http.StatusNotFound, errors.New("character not found"))
}

func (a API) handleBooks(w http.ResponseWriter, _ *http.Request) {
	books, err := a.master.ListBooks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (a API) handleBookRows(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("book")
	if err := masterdata.ValidateBookName(name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	limit, err := parseLimit(r.URL.Query().Get("limit"), 100)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	rows, err := a.master.LoadRows(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if limit < len(rows) {
		rows = rows[:limit]
	}
	writeJSON(w, http.StatusOK, rows)
}

func parseLimit(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 0 {
		return 0, errors.New("limit must be a non-negative integer")
	}
	if limit > 1000 {
		return 0, errors.New("limit must be less than or equal to 1000")
	}
	return limit, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload = append(payload, '\n')

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, apiError{Error: err.Error()})
}
