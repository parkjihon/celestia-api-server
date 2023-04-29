package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	logging "github.com/ipfs/go-log/v2"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
)

var log = logging.Logger("gateway")

const (
	namespacedSharesEndpoint = "/namespaced_shares"
	namespacedDataEndpoint   = "/namespaced_data"
)

func writeError(w http.ResponseWriter, statusCode int, endpoint string, err error) {
	log.Errorw("serving request", "endpoint", endpoint, "err", err)

	w.WriteHeader(statusCode)
	errBody, jerr := json.Marshal(err.Error())
	if jerr != nil {
		log.Errorw("serializing error", "endpoint", endpoint, "err", jerr)
		return
	}
	_, werr := w.Write(errBody)
	if werr != nil {
		log.Errorw("writing error response", "endpoint", endpoint, "err", werr)
	}
}

func main() {
	r := gin.Default()
	r.GET("/namespaced_data/:nid", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		nid := c.Param("nid") // fcb1a75aeaed7065
		var heightCore int64
		var blob string
		var blobs []string
		rows, err := db.Query("SELECT blob_base64, height_core FROM blobs WHERE nid = ? order by height_core desc limit 1", nid)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&blob, &heightCore)
			if err != nil {
				log.Fatal(err)
			}
			blobs = append(blobs, blob)
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   blobs,
			"height": heightCore,
		})
	})
	r.GET("/namespaced_data/:nid/height/:height", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		nid := c.Param("nid")       // fcb1a75aeaed7065
		height := c.Param("height") // 200000
		var heightCore int
		var blob string
		var blobs []string
		heightCore, err = strconv.Atoi(height)
		if err != nil {
			c.String(http.StatusOK, "Wrong height %s", height)
			return
		}

		var storeHeight int
		// check whether heightCore is upper than latest height
		err = db.QueryRow("SELECT max(height) FROM core_blocks").Scan(&storeHeight)
		if err != nil {
			log.Fatal(err)
		}

		rows, err := db.Query("SELECT blob_base64, height_core FROM blobs WHERE nid = ? and height_core = ?", nid, heightCore)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&blob, &heightCore)
			if err != nil {
				log.Fatal(err)
			}
			blobs = append(blobs, blob)
		}
		if storeHeight < heightCore {
			c.Header("Content-Type", "application/json")
			c.Writer.Header().Set("Content-Type", "application/json")
			err = fmt.Errorf("current head local chain head: %d is lower than requested height: %d"+" give header sync some time and retry later", storeHeight, heightCore)
			writeError(c.Writer, http.StatusInternalServerError, namespacedDataEndpoint, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   blobs,
			"height": heightCore,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
