package licenseplate

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var mutex = &sync.Mutex{}

type Controller struct {
	blockchain     *Blockchain
	currentNodeURL string
}

type ResponseToSend struct {
	Note string
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(c.blockchain)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func (c *Controller) RegisterMulta(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body) // read the body of the request
	if err != nil {
		log.Fatalln("Error RegisterMulta", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error RegisterMulta", err)
	}
	var multa Multa
	if err := json.Unmarshal(body, &multa); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error RegisterMulta unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	success := c.blockchain.RegisterMulta(multa) // registers the multa into the blockchain
	if !success {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	var resp ResponseToSend
	resp.Note = "Multa created and broadcast successfully."
	data, _ := json.Marshal(resp)
	w.Write(data)
	return
}

func (c *Controller) RegisterAndBroadcastMulta(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	body, err := ioutil.ReadAll(r.Body) // read the body of the request
	if err != nil {
		log.Fatalln("Error RegisterMulta", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error RegisterMulta", err)
	}
	var multa Multa
	if err := json.Unmarshal(body, &multa); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error RegisterMulta unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	success := c.blockchain.RegisterMulta(multa) // registers the multa into the blockchain
	if !success {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// broadcast
	for _, node := range c.blockchain.NetworkNodes {
		if node != c.currentNodeURL {
			// call /register-node in node
			MakePostCall(node+"/multa", body)
		}
	}
	var empty []byte
	MakeGetCall(c.currentNodeURL+"/mine", empty)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	var resp ResponseToSend
	resp.Note = "Multa created and broadcast successfully."
	data, _ := json.Marshal(resp)
	w.Write(data)
	mutex.Unlock()
}

func (c *Controller) Mine(w http.ResponseWriter, r *http.Request) {
	lastBlock := c.blockchain.GetLastBlock()
	previousBlockHash := lastBlock.Hash
	currentBlockData := BlockData{Index: strconv.Itoa(lastBlock.Index - 1), Multas: c.blockchain.PendingMultas}
	currentBlockDataAsByteArray, _ := json.Marshal(currentBlockData)
	currentBlockDataAsStr := base64.URLEncoding.EncodeToString(currentBlockDataAsByteArray)

	nonce := c.blockchain.ProofOfWork(previousBlockHash, currentBlockDataAsStr)
	blockHash := c.blockchain.HashBlock(previousBlockHash, currentBlockDataAsStr, nonce)
	newBlock := c.blockchain.CreateNewBlock(nonce, previousBlockHash, blockHash)
	blockToBroadcast, _ := json.Marshal(newBlock)

	for _, node := range c.blockchain.NetworkNodes {
		if node != c.currentNodeURL {
			// call /receive-new-block in node
			MakePostCall(node+"/receive-new-block", blockToBroadcast)
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	var resp ResponseToSend
	resp.Note = "New block mined and broadcast successfully."
	data, _ := json.Marshal(resp)
	w.Write(data)
	return
}

func (c *Controller) RegisterNode(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body) // read the body of the request
	if err != nil {
		log.Fatalln("Error RegisterNode", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error RegisterNode", err)
	}
	var node struct {
		NewNodeURL string `json:"newNodeUrl"`
	}
	if err := json.Unmarshal(body, &node); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error RegisterNode unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	var resp ResponseToSend
	if node.NewNodeURL != c.currentNodeURL {
		success := c.blockchain.RegisterNode(node.NewNodeURL) // registers the node into the blockchain
		if !success {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	resp.Note = "Node registered successfully."
	data, _ := json.Marshal(resp)
	w.Write(data)
	return
}

func (c *Controller) RegisterNodesBulk(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body) // read the body of the request
	if err != nil {
		log.Fatalln("Error RegisterNodesBulk", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error RegisterNodesBulk", err)
	}
	var allNodes []string
	if err := json.Unmarshal(body, &allNodes); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error RegisterNodesBulk unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	for _, node := range allNodes {
		if node != c.currentNodeURL {
			success := c.blockchain.RegisterNode(node) // registers the node into the blockchain
			if !success {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	var resp ResponseToSend
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	resp.Note = "Bulk registration successful."
	data, _ := json.Marshal(resp)
	w.Write(data)
	return
}

func MakeCall(mode string, url string, jsonStr []byte) interface{} {
	// call url in node
	log.Println(mode)
	log.Println(url)
	req, err := http.NewRequest(mode, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in call " + url)
		log.Println(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	var returnValue interface{}
	if err := json.Unmarshal(respBody, &returnValue); err != nil { // unmarshal body contents as a type Candidate
		if err != nil {
			log.Fatalln("Error "+url+" unmarshalling data", err)
			return nil
		}
	}
	log.Println(returnValue)
	return returnValue
}

func MakePostCall(url string, jsonStr []byte) {
	// call url in POST
	MakeCall("POST", url, jsonStr)
}

func MakeGetCall(url string, jsonStr []byte) interface{} {
	// call url in GET
	return MakeCall("GET", url, jsonStr)
}

func BroadcastNode(newNode string, nodes []string) {
	for _, node := range nodes {
		if node != newNode {
			var registerNodesJSON = []byte(`{"newnodeurl":"` + newNode + `"}`)

			// call /register-node in node
			MakePostCall(node+"/register-node", registerNodesJSON)
		}
	}
}

func (c *Controller) RegisterAndBroadcastNode(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body) // read the body of the request
	if err != nil {
		log.Fatalln("Error RegisterNode", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error RegisterNode", err)
	}
	var node struct {
		NewNodeURL string `json:"newnodeurl"`
	}
	if err := json.Unmarshal(body, &node); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error RegisterNode unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	var resp ResponseToSend
	success := c.blockchain.RegisterNode(node.NewNodeURL) // registers the node into the blockchain
	if !success {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// broadcast
	BroadcastNode(node.NewNodeURL, c.blockchain.NetworkNodes)

	// register all nodes in new node
	allNodes := append(c.blockchain.NetworkNodes, c.currentNodeURL)
	payload, err := json.Marshal(allNodes)
	registerBulkJSON := []byte(payload)
	MakePostCall(node.NewNodeURL+"/register-nodes-bulk", registerBulkJSON)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	resp.Note = "Node registered successfully."
	data, _ := json.Marshal(resp)
	w.Write(data)
	return
}

func (c *Controller) ReceiveNewBlock(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body) // read the body of the request
	if err != nil {
		log.Fatalln("Error ReceiveNewBlock", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error ReceiveNewBlock", err)
	}

	var blockReceived Block
	if err := json.Unmarshal(body, &blockReceived); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error RegisterNode unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	var resp ResponseToSend
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// append block to blockchain
	if c.blockchain.CheckNewBlockHash(blockReceived) {
		resp.Note = "New Block received and accepted."
		c.blockchain.PendingMultas = Multas{}
		c.blockchain.Chain = append(c.blockchain.Chain, blockReceived)
	} else {
		resp.Note = "New Block rejected."
	}

	data, _ := json.Marshal(resp)
	w.Write(data)
	return
}

func (c *Controller) Consensus(w http.ResponseWriter, r *http.Request) {
	maxChainLength := 0
	var longestChain *Blockchain
	var resp ResponseToSend
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	for _, node := range c.blockchain.NetworkNodes {
		if node != c.currentNodeURL {
			// call /blockchain in node
			// call url in node
			req, err := http.NewRequest("GET", node+"/blockchain", nil)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Println("Error retrieving blockchain")
				log.Println(err)
			}
			defer resp.Body.Close()
			respBody, err := ioutil.ReadAll(resp.Body)
			var chain *Blockchain
			if err := json.Unmarshal(respBody, &chain); err != nil { // unmarshal body contents as a type Candidate
				if err != nil {
					log.Fatalln("Error unmarshalling data", err)
				}
			}

			if chain != nil {
				chainLength := len(chain.Chain)
				if maxChainLength < chainLength {
					maxChainLength = chainLength
					longestChain = chain
				}
			}
		}
	}

	log.Println(longestChain.ChainIsValid())

	if maxChainLength > len(c.blockchain.Chain) && longestChain.ChainIsValid() {
		c.blockchain.Chain = longestChain.Chain
		c.blockchain.PendingMultas = longestChain.PendingMultas

		resp.Note = "This chain has been replaced."
	} else {
		resp.Note = "This chain has not been replaced."
	}

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(resp)
	w.Write(data)
	return
}

func (c *Controller) GetMultasForPlate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	NumeroPlaca := vars["NumeroPlaca"]

	multas := c.blockchain.GetMultasForPlate(NumeroPlaca)
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(multas)
	w.Write(data)
	return
}