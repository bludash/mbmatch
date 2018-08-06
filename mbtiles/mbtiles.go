package mbtiles

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	// sqlite3 because mbtiles is sqlite3 only
	_ "github.com/mattn/go-sqlite3"
)

// DB an MBTile reader
type DB struct {
	*sql.DB
	Debug bool
}

// NewDB open an MBTile for reading
func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path+"?mode=ro")
	if err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

// ReadTile read a tile returns empty []byte if not existing
func (db *DB) ReadTile(z uint8, x uint64, y uint64) ([]byte, error) {
	var data []byte
	err := db.QueryRow("select tile_data from tiles where zoom_level = ? and tile_column = ? and tile_row = ?", z, x, y).Scan(&data)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ServeHTTP serve the mbtiles at /tiles/11/618/722.pbf
func (db *DB) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	s := strings.Split(req.URL.Path, "/")

	if len(s) != 5 {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}
	z, _ := strconv.Atoi(s[2])
	x, _ := strconv.Atoi(s[3])
	y, _ := strconv.Atoi(strings.Trim(s[4], ".pbf"))

	data, err := db.ReadTile(uint8(z), uint64(x), uint64(1<<uint(z)-y-1))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(data) == 0 {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Header().Set("Content-Encoding", "gzip")
	_, _ = w.Write(data)
}
