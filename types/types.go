package types

type Top10TXedNID struct {
	Nid     string `json:"nid" db:"nid"`
	ChainId string `json:"chainId" db:"chain_id"`
	Cnt     int    `json:"cnt" db:"cnt"`
}

type NewCreatedNID struct {
	Nid     string `json:"nid" db:"nid"`
	ChainId string `json:"chainId" db:"chain_id"`
	Time    string `json:"time" db:"time"`
}

type CoreBlocksRow struct {
	//Id int `db:"id"`
	BlockHash       string `json:"blockHash" db:"block_hash"`
	ChainId         string `json:"chainId" db:"chain_id"`
	Height          string `json:"height" db:"height"`
	Time            string `json:"time" db:"time"`
	ProposerAddress string `json:"proposerAddress" db:"proposer_address"`
	TxCnt           int    `json:"txCnt" db:"tx_cnt"`
	BlobCnt         int    `json:"blobCnt" db:"blob_cnt"`
	SquareSize      string `json:"squareSize" db:"square_size"`
}

type ChainsRow struct {
	NidBase64  string `json:"nidBase64" db:"nid_base64"`
	Nid        string `json:"nid" db:"nid"`
	ChainId    string `json:"chainId" db:"chain_id"`
	Height     int64  `json:"height" db:"height"`
	HeightCore string `json:"heightCore" db:"height_core"`
	CntBlob    int    `json:"cntBlob" db:"cnt_blob"`
	CntTx      int    `json:"cntTx" db:"cnt_tx"`
}

type BlobsRow struct {
	//Id int `db:"id"`
	NidBase64    string `json:"nidBase64" db:"nid_base64"`
	Nid          string `json:"nid" db:"nid"`
	ShareVersion int64  `json:"shareVersion" db:"share_version"`
	BlockHash    string `json:"blockHash" db:"block_hash"`
	ChainId      string `json:"chainId" db:"chain_id"`
	Height       int64  `json:"height" db:"height"`
	HeightCore   string `json:"heightCore" db:"height_core"`
	Time         string `json:"time" db:"time"`
	VersionApp   uint64 `json:"versionApp" db:"version_app"`
	VersionBlock uint64 `json:"versionBlock" db:"version_block"`
	ValidatorCnt int    `json:"validatorCnt" db:"validator_cnt"`
	TxCnt        int    `json:"txCnt" db:"tx_cnt"`
	BlobBase64   string `json:"blobBase64" db:"blob_base64"`
}

type RollupTxRow struct {
	//Id int `db:"id"`
	NidBase64   string `json:"nidBase64" db:"nid_base64"`
	Nid         string `json:"nid" db:"nid"`
	ChainId     string `json:"chainId" db:"chain_id"`
	Height      int64  `json:"height" db:"height"`
	TxHash      string `json:"txhash" db:"txhash"`
	Time        string `json:"time" db:"time"`
	Memo        string `json:"memo" db:"memo"`
	TypeUrl     string `json:"typeUrl" db:"type_url"`
	FeeAmount   string `json:"feeAmount" db:"fee_amount"`
	FeeGasLimit uint64 `json:"feeGasLimit" db:"fee_gas_limit"`
}

type ChainSummary struct {
	NidBase64 string `json:"nidBase64" db:"nid_base64"`
	Nid       string `json:"nid" db:"nid"`
	ChainId   string `json:"chainId" db:"chain_id"`
	Height    int64  `json:"height" db:"height"`
	CntTx     int    `json:"cntTx" db:"cnt_tx"`
	Time      string `json:"time" db:"time"`
}
