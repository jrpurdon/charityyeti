package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// brainTreeData is information we receive from the front end after a transaction
// gets started, plus the options array we add on. We pass this data back to
// Brain Tree opaquely.
type brainTreeData struct {
	DonorDocumentID string            `json:"donorDocumentId"`
	OriginalTweetID string            `json:"originalTweetId"`
	InvokerTweetID  string            `json:"invokerTweetId"`
	inReplyToUser   int64             `json:"inReplyToUser"`
	Nonce           string            `json:"paymentMethodNonce"`
	Amount          string            `json:"amount"`
	DeviceData      string            `json:"deviceData"`
	Options         []map[string]bool `json:"options"`
}

// brainTreeTransaction is the return data the middleware sends us after a
// successful transaction
type brainTreeTransaction struct {
	Amount         string         `json:"amount"`
	BillingDetails billingDetails `json:"billingDetails"`
	ID             string         `json:"id"`
}

// billingDetails is the data Brain Tree stored about this user and returned to us
type billingDetails struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	StreetAddress   string `json:"streetAddress"`
	ExtendedAddress string `json:"extendedAddress"`
	Locality        string `json:"locality"`
	Region          string `json:"region"`
	PostalCode      string `json:"postalCode`
	Country         string `json:"countryName"`
}

// receivePaymentRequest receives a brainTreeData struct as an POST body,
// and then relays that request to the middleware that forwards directly
// to Brain Tree. We then send the middleware's response back on the request.
func receivePaymentRequest(w http.ResponseWriter, r *http.Request) {
	log.Info("Received request for BrainTree token")

	// unmarshal the request body
	var btData brainTreeData
	if err := json.NewDecoder(r.Body).Decode(&btData); err != nil {
		// if we
		log.Errorf("could not decode incoming request: %v", err)
		http.Error(w, "bad request", 400)
		return
	}

	log.Info("checking for nonce...")
	if btData.Nonce == "" {
		http.Error(w, "bad request", 400)
		return
	}

	// oh good, things are kosher, at least so far
	log.Debugf("Nonce found: %v", btData.Nonce)

	// we need to send this nonce to BrainTree so they can process this transaction
	// TODO: we need to actually do the needful, kindly
	transaction, statusCode, err := doBrainTreeRequest(btData)
	if err != nil {
		// uh oh
		log.Errorf("Brain Tree request failed: %v", err)
		http.Error(w, err.Error(), statusCode)
		return
	}

	// brain tree request was fine, so now we need to update Mongo
	d := donation{
		TransactionID:   transaction.ID,
		OriginalTweetID: btData.OriginalTweetID,
		InvokerTweetID:  btData.InvokerTweetID,
		Honorary:        honorary,
		DonationValue:   transaction.Amount,
	}
	_, err = addDonation(btData.DonorDocumentID, d)

	// if everything is cool then we done
	w.WriteHeader(200)

	// not really necessary but I like closure
	return
}

// doBrainTreeRequest sends the request to the Brain Tree middleware
func doBrainTreeRequest(data brainTreeData) (*brainTreeTransaction, int, error) {
	log.Info("Building request to send Nonce to BrainTree")

	// marshal the incoming data to send on the wire
	btBytes, err := json.Marshal(data)
	if err != nil {
		// not sure how this can even happen
		log.Errorf("Could not marshal incoming brain tree struct: %v", err)
		return nil, 500, err
	}

	// setup the request
	req, err := http.NewRequest(http.MethodGet, cfg.MiddlewareEndpoint, bytes.NewReader(btBytes))
	if err != nil {
		log.Errorf("Could not create request to middleware: %v", err)
		return nil, 500, err
	}

	// set the headers
	req.Header.Set("Content-Type", "application/json")

	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Could not make the request to the middleware: %v", err)
		return nil, 500, err
	}
	defer resp.Body.Close()

	// read the response back
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Could not read response: %v", err)
		return nil, 500, err
	}

	log.Infof("Response: %v", string(respBytes))

	// unmarshal to transaction type
	var t brainTreeTransaction
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		// something weird happend
		log.Errorf("Could not decode transaction data: %v", err)
		return nil, 500, err
	}

	// everything is great!
	return &t, 200, nil
}

// checkMiddlewareHealth hits the health endpoint in the middleware
// and returns the status code
func checkMiddlewareHealth(w http.ResponseWriter, r *http.Request) {
	log.Info("Checking middleware health")
	resp, err := http.Get(cfg.MiddlewareHealth)
	if err != nil {
		log.Errorf("Could not make request to middleware health: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	respBytes, _ := ioutil.ReadAll(resp.Body)

	log.Infof("Response from middleware: %v", string(respBytes))
	w.WriteHeader(resp.StatusCode)
	w.Write(respBytes)
	return
}
