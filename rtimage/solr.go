package rtimage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// SolrClient connessione a solr
type SolrClient struct {
	host string
}

type SolrDoc struct {
	ID             string `json:"id"`
	Sha256         string `json:"sha256_s"`
	SourceFilename string `json:"sourceFilename_s"`
}

type ResponseHeader struct {
	Status int
	QTime  int
	Params interface{}
}

type Response struct {
	NumFound int
	Start    int
	Docs     []SolrDoc
}
type SolrResp struct {
	ResponseHeader ResponseHeader
	Response       Response
}

func (c SolrClient) getByID(id string) (SearchDoc, error) {

	url := fmt.Sprintf("%s/photos/select?q=id:%s", c.host, id)
	// fmt.Println("URL:>", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//log.Printf(string(body))
	respobj := SolrResp{}
	err = json.Unmarshal(body, &respobj)
	if err != nil {
		panic(err)
	}
	//log.Printf("Search Response : %v", respobj)
	if respobj.Response.NumFound > 0 {
		solrDoc := respobj.Response.Docs[0]
		return SearchDoc{sha256: solrDoc.Sha256}, nil
	} else {
		return SearchDoc{}, fmt.Errorf("Non trovato %s", id)
	}
}

func (c SolrClient) save(doc SearchDoc) {

	url := fmt.Sprintf("%s/photos/update?commit=true", c.host)
	// fmt.Println("URL:>", url)

	solrDoc := SolrDoc{ID: doc.sha256, Sha256: doc.sha256, SourceFilename: doc.filename}
	docs := []SolrDoc{solrDoc}
	jsonStr, err := json.Marshal(docs)
	//log.Printf("update document :> %s", jsonStr)

	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}
