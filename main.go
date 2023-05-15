package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	logging "github.com/ipfs/go-log/v2"

	"github.com/gin-gonic/gin"

	"celestia-api-server/types"

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
			fmt.Println(err)
		}
		defer db.Close()

		nid := c.Param("nid") // fcb1a75aeaed7065
		var heightCore int64
		var blob string
		var blobs []string
		rows, err := db.Query("SELECT blob_base64, height_core FROM blobs WHERE nid = ? order by height_core desc limit 1", nid)
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&blob, &heightCore)
			if err != nil {
				fmt.Println(err)
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
			fmt.Println(err)
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
			fmt.Println(err)
		}

		rows, err := db.Query("SELECT blob_base64, height_core FROM blobs WHERE nid = ? and height_core = ?", nid, heightCore)
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&blob, &heightCore)
			if err != nil {
				fmt.Println(err)
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
	r.GET("/explorer/core/info", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		// 	- 총 nid 수
		//   select count(*) as nid_cnt from (select nid,count(*) as cnt from blobs group by nid) as A;
		// - 총 blob 수
		//   select count(*) from blobs;
		// - 총 tran 수
		//   select count(*) from rollup_txs;

		var cntNIDs int
		var cntBlobs int
		var cntTXs int

		err = db.QueryRow("select count(*) as nid_cnt from (select nid,count(*) as cnt from blobs group by nid) as A").Scan(&cntNIDs)
		if err != nil {
			fmt.Println(err)
		}
		err = db.QueryRow("select count(*) from blobs").Scan(&cntBlobs)
		if err != nil {
			fmt.Println(err)
		}
		err = db.QueryRow("select count(*) from rollup_txs").Scan(&cntTXs)
		if err != nil {
			fmt.Println(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"cntNIDs":  cntNIDs,
			"cntBlobs": cntBlobs,
			"cntTXs":   cntTXs,
		})
	})
	r.GET("/explorer/core/summary", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		var top10TXedNID types.Top10TXedNID
		var top10TXedNIDs []types.Top10TXedNID

		rows, err := db.Query("select nid, chain_id, count(*) as cnt from rollup_txs group by nid, chain_id order by cnt desc limit 5")
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&top10TXedNID.Nid, &top10TXedNID.ChainId, &top10TXedNID.Cnt)
			if err != nil {
				fmt.Println(err)
			}
			top10TXedNIDs = append(top10TXedNIDs, top10TXedNID)
		}

		// var topHeightNID types.Top10TXedNID
		// var topHeightNIDs []types.Top10TXedNID
		// rows, err = db.Query("select nid, max(height) as height from blobs where height <> 0 group by nid order by height desc limit 5")
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// defer rows.Close()
		// for rows.Next() {
		// 	err := rows.Scan(&topHeightNID.Nid, &topHeightNID.Cnt)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 	}
		// 	err = db.QueryRow("select chain_id from blobs where nid = ? limit 1", topHeightNID.Nid).Scan(&topHeightNID.ChainId)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 	}

		// 	topHeightNIDs = append(topHeightNIDs, topHeightNID)
		// }

		//select nid, FROM_UNIXTIME(`time`) as time, chain_id from blobs where height = 1 order by id desc limit 10;
		var newCreatedNID types.NewCreatedNID
		var newCreatedNIDs []types.NewCreatedNID

		rows, err = db.Query("select nid, FROM_UNIXTIME(`time`) as time, chain_id from blobs where height = 1 order by id desc limit 5")
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&newCreatedNID.Nid, &newCreatedNID.Time, &newCreatedNID.ChainId)
			if err != nil {
				fmt.Println(err)
			}
			newCreatedNIDs = append(newCreatedNIDs, newCreatedNID)
		}

		c.JSON(http.StatusOK, gin.H{
			"top10TXedNIDs": top10TXedNIDs,
			// "topHeightNIDs":  topHeightNIDs,
			"newCreatedNIDs": newCreatedNIDs,
		})
	})
	r.GET("/explorer/blocks", func(c *gin.Context) {
		//select * from core_blocks order by height desc limit 10;
		//block_hash, height, time, proposer_address, tx_cnt, blob_cnt, square_size
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		var coreBlocksRow types.CoreBlocksRow
		var coreBlocksRows []types.CoreBlocksRow

		rows, err := db.Query("select block_hash, height, time, proposer_address, tx_cnt, blob_cnt, square_size from core_blocks order by height desc limit 30")
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&coreBlocksRow.BlockHash,
				&coreBlocksRow.Height,
				&coreBlocksRow.Time,
				&coreBlocksRow.ProposerAddress,
				&coreBlocksRow.TxCnt,
				&coreBlocksRow.BlobCnt,
				&coreBlocksRow.SquareSize,
			)
			if err != nil {
				fmt.Println(err)
			}
			coreBlocksRows = append(coreBlocksRows, coreBlocksRow)
		}

		c.JSON(http.StatusOK, gin.H{
			"coreBlocksRows": coreBlocksRows,
		})
	})
	r.GET("/explorer/chains", func(c *gin.Context) {
		//select nid, max(height_core) as max_height_core, count(*) cnt_blob from blobs group by nid order by max_height_core desc limit 30;
		// nid, nid_base64, chain_id, height, height_core, cntBlob
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		var chainsRow types.ChainsRow
		var chainsRows []types.ChainsRow

		rows, err := db.Query("select nid, max(height_core) as max_height_core from blobs group by nid order by max_height_core desc limit 30")
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&chainsRow.Nid,
				&chainsRow.HeightCore,
				//&chainsRow.CntBlob,
			)
			if err != nil {
				fmt.Println(err)
			}
			err = db.QueryRow("select nid_base64, chain_id, height from blobs where nid = ? and height_core = ?", chainsRow.Nid, chainsRow.HeightCore).Scan(
				&chainsRow.NidBase64,
				&chainsRow.ChainId,
				&chainsRow.Height,
			)
			if err != nil {
				fmt.Println(err)
			}

			chainsRows = append(chainsRows, chainsRow)
		}

		c.JSON(http.StatusOK, gin.H{
			"chainsRows": chainsRows,
		})
	})
	r.GET("/explorer/blobs", func(c *gin.Context) {
		//select * from blobs order by height_core desc limit 30;
		//nid, share_version, block_hash, chain_id, height, height_core, time, version_app, version_block, validator_cnt, tx_cnt
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		var blobsRow types.BlobsRow
		var blobsRows []types.BlobsRow

		rows, err := db.Query("select block_hash, nid, share_version, chain_id, height, height_core, tx_cnt, FROM_UNIXTIME(`time`) as time from blobs order by height_core desc limit 30")
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&blobsRow.BlockHash,
				&blobsRow.Nid,
				&blobsRow.ShareVersion,
				&blobsRow.ChainId,
				&blobsRow.Height,
				&blobsRow.HeightCore,
				&blobsRow.TxCnt,
				&blobsRow.Time,
			)
			if err != nil {
				fmt.Println(err)
			}
			blobsRows = append(blobsRows, blobsRow)
		}

		c.JSON(http.StatusOK, gin.H{
			"blobsRows": blobsRows,
		})
	})
	r.GET("/explorer/blocks/:height", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		height := c.Param("height")
		var blockRow types.CoreBlocksRow

		fmt.Println(height)
		err = db.QueryRow("select block_hash, chain_id, height, time, proposer_address, tx_cnt, blob_cnt, square_size from core_blocks where height = ?", height).Scan(
			&blockRow.BlockHash,
			&blockRow.ChainId,
			&blockRow.Height,
			&blockRow.Time,
			&blockRow.ProposerAddress,
			&blockRow.TxCnt,
			&blockRow.BlobCnt,
			&blockRow.SquareSize,
		)
		if err != nil {
			fmt.Println(err)
		}

		var blobsRow types.BlobsRow
		var blobsRows []types.BlobsRow

		rows, err := db.Query("select nid, share_version, chain_id, height, height_core, tx_cnt, FROM_UNIXTIME(`time`) as time, block_hash from blobs where height_core = ?", height)
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&blobsRow.Nid,
				&blobsRow.ShareVersion,
				&blobsRow.ChainId,
				&blobsRow.Height,
				&blobsRow.HeightCore,
				&blobsRow.TxCnt,
				&blobsRow.Time,
				&blobsRow.BlockHash,
			)
			if err != nil {
				fmt.Println(err)
			}
			blobsRows = append(blobsRows, blobsRow)
		}

		c.JSON(http.StatusOK, gin.H{
			"blockRow":  blockRow,
			"blobsRows": blobsRows,
		})
	})
	r.GET("/explorer/rollups/:nid", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		nid := c.Param("nid")

		var chainSummary types.ChainSummary
		err = db.QueryRow("select nid, nid_base64, chain_id, FROM_UNIXTIME(`time`) as time, height from blobs where nid = ? order by height desc limit 1", nid).Scan(
			&chainSummary.Nid,
			&chainSummary.NidBase64,
			&chainSummary.ChainId,
			&chainSummary.Time,
			&chainSummary.Height,
		)
		if err != nil {
			fmt.Println(err)
		}
		err = db.QueryRow("select count(*) as cnt_tx from rollup_txs where nid = ?", nid).Scan(
			&chainSummary.CntTx,
		)
		if err != nil {
			fmt.Println(err)
		}

		var blobsRow types.BlobsRow
		var blobsRows []types.BlobsRow

		rows, err := db.Query("select nid, share_version, chain_id, height, height_core, tx_cnt, FROM_UNIXTIME(`time`) as time, block_hash from blobs where nid = ? order by height desc limit 10", nid)
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&blobsRow.Nid,
				&blobsRow.ShareVersion,
				&blobsRow.ChainId,
				&blobsRow.Height,
				&blobsRow.HeightCore,
				&blobsRow.TxCnt,
				&blobsRow.Time,
				&blobsRow.BlockHash,
			)
			if err != nil {
				fmt.Println(err)
			}
			blobsRows = append(blobsRows, blobsRow)
		}

		var rollupTxRow types.RollupTxRow
		var rollupTxRows []types.RollupTxRow

		rows, err = db.Query("select nid_base64, nid, chain_id, height, txhash, FROM_UNIXTIME(`time`) as time, memo, type_url, fee_amount, fee_gas_limit from rollup_txs where nid = ? order by height desc limit 10", nid)
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&rollupTxRow.NidBase64,
				&rollupTxRow.Nid,
				&rollupTxRow.ChainId,
				&rollupTxRow.Height,
				&rollupTxRow.TxHash,
				&rollupTxRow.Time,
				&rollupTxRow.Memo,
				&rollupTxRow.TypeUrl,
				&rollupTxRow.FeeAmount,
				&rollupTxRow.FeeGasLimit,
			)
			if err != nil {
				fmt.Println(err)
			}
			rollupTxRows = append(rollupTxRows, rollupTxRow)
		}

		c.JSON(http.StatusOK, gin.H{
			"chainSummary": chainSummary,
			"blobsRows":    blobsRows,
			"rollupTxRows": rollupTxRows,
		})
	})
	r.GET("/explorer/rollups/:nid/blocks/:height", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		height := c.Param("height")
		nid := c.Param("nid")
		var blobRow types.BlobsRow

		fmt.Println(height)
		err = db.QueryRow("select nid_base64, nid, share_version, block_hash, chain_id, height, height_core, time, version_app, version_block, validator_cnt, tx_cnt, blob_base64 from blobs where nid = ? and height = ?", nid, height).Scan(
			&blobRow.NidBase64,
			&blobRow.Nid,
			&blobRow.ShareVersion,
			&blobRow.BlockHash,
			&blobRow.ChainId,
			&blobRow.Height,
			&blobRow.HeightCore,
			&blobRow.Time,
			&blobRow.VersionApp,
			&blobRow.VersionBlock,
			&blobRow.ValidatorCnt,
			&blobRow.TxCnt,
			&blobRow.BlobBase64,
		)
		if err != nil {
			fmt.Println(err)
		}

		var rollupTxRow types.RollupTxRow
		var rollupTxRows []types.RollupTxRow

		rows, err := db.Query("select nid_base64, nid, chain_id, height, txhash, FROM_UNIXTIME(`time`) as time, memo, type_url, fee_amount, fee_gas_limit from rollup_txs where nid = ? and height = ?", nid, height)
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&rollupTxRow.NidBase64,
				&rollupTxRow.Nid,
				&rollupTxRow.ChainId,
				&rollupTxRow.Height,
				&rollupTxRow.TxHash,
				&rollupTxRow.Time,
				&rollupTxRow.Memo,
				&rollupTxRow.TypeUrl,
				&rollupTxRow.FeeAmount,
				&rollupTxRow.FeeGasLimit,
			)
			if err != nil {
				fmt.Println(err)
			}
			rollupTxRows = append(rollupTxRows, rollupTxRow)
		}

		c.JSON(http.StatusOK, gin.H{
			"blobRow":      blobRow,
			"rollupTxRows": rollupTxRows,
			//"blobsRows": blobsRows,
		})
	})
	r.GET("/explorer/rollups/:nid/txs/:txhash", func(c *gin.Context) {
		db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/celestia-rollup-explorer")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		nid := c.Param("nid")
		txhash := c.Param("txhash")

		var txRow types.RollupTxRow
		err = db.QueryRow("select nid_base64, nid, chain_id, height, txhash, FROM_UNIXTIME(`time`) as time, memo, type_url, fee_amount, fee_gas_limit from rollup_txs where nid = ? and txhash = ?", nid, txhash).Scan(
			&txRow.NidBase64,
			&txRow.Nid,
			&txRow.ChainId,
			&txRow.Height,
			&txRow.TxHash,
			&txRow.Time,
			&txRow.Memo,
			&txRow.TypeUrl,
			&txRow.FeeAmount,
			&txRow.FeeGasLimit,
		)
		if err != nil {
			fmt.Println(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"txRow": txRow,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
