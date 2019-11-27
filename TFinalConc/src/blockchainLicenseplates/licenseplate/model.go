package licenseplate

type Multa struct {
	NumeroPlaca string `json:"numeroplaca"`
	Codigo   string  `json:"codigo"`
	Fecha    string   `json:"fecha"`
	Gravedad string  `json:"gravedad"`
	Monto    float64 `json:"monto"`
}

type Multas []Multa

type Block struct {
	Index             int    `json:"index"`
	Timestamp         int64  `json:"timestamp"`
	Multas            Multas `json:"multas"`
	Nonce             int    `json:"nonce"`
	Hash              string `json:"hash"`
	PreviousBlockHash string `json:"previousblockhash"`
}

type Blocks []Block

type Blockchain struct {
	Chain         Blocks   `json:"chain"`
	PendingMultas Multas   `json:"pending_multas"`
	NetworkNodes  []string `json:"network_nodes"`
}

type BlockData struct {
	Index  string
	Multas Multas
}
