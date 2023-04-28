package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// sql.DB 객체 생성
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default()
	r.GET("/namespaced_data/:nid", func(c *gin.Context) {
		nid := c.Param("nid") // fcb1a75aeaed7065

		// 복수 Row를 갖는 SQL 쿼리
		var heightCore int64
		var blob string
		var blobs []string
		rows, err := db.Query("SELECT blob_base64, height_core FROM blobs WHERE nid = ? order by height_core desc limit 1", nid)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close() //반드시 닫는다 (지연하여 닫기)

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
		nid := c.Param("nid")       // fcb1a75aeaed7065
		height := c.Param("height") // 200000

		// 복수 Row를 갖는 SQL 쿼리
		var heightCore int
		var blob string
		var blobs []string
		heightCore, err = strconv.Atoi(height)
		if err != nil {
			c.String(http.StatusOK, "Wrong height %s", height)
			return
		}
		rows, err := db.Query("SELECT blob_base64, height_core FROM blobs WHERE nid = ? and height_core = ?", nid, heightCore)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close() //반드시 닫는다 (지연하여 닫기)

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
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
