package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kunterbunt/calendarium-server/model"
)

type billbeeInvoiceAddress struct {
	BillbeeID   int64  `json:"BillbeeId"`
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	Company     string `json:"Company"`
	Street      string `json:"Street"`
	HouseNumber string `json:"HouseNumber"`
	Zip         string `json:"Zip"`
	City        string `json:"City"`
	Country     string `json:"Country"`
	Email       string `json:"Email"`
}

type billbeeShippingAddress struct {
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	Company     string `json:"Company"`
	Street      string `json:"Street"`
	HouseNumber string `json:"HouseNumber"`
	Zip         string `json:"Zip"`
	City        string `json:"City"`
	Country     string `json:"Country"`
	Email       string `json:"Email"`
}

type billbeeProduct struct {
	Title     string `json:"Title"`
	BillbeeID int64  `json:"BillbeeId"`
}

type billbeeOrderItems struct {
	Product    billbeeProduct `json:"Product"`
	Quantity   int            `json:"Quantity"`
	TotalPrice float64        `json:"TotalPrice"`
}

type billbeeSeller struct {
	Platform        string `json:"Platform"`
	BillbeeShopName string `json:"BillbeeShopName"`
	BillbeeShopID   int64  `json:"BillbeeShopId"`
	Email           string `json:"Email"`
}

type billbeeCustomer struct {
	Name  string `json:"Name"`
	Email string `json:"Email"`
}

type billbeeBody struct {
	CreatedAt       string                 `json:"CreatedAt"`
	OrderNumber     string                 `json:"OrderNumber"`
	InvoiceAddress  billbeeInvoiceAddress  `json:"InvoiceAddress"`
	ShippingAddress billbeeShippingAddress `json:"ShippingAddress"`
	PaymentMethod   int                    `json:"PaymentMethod"`
	ShippingCost    float64                `json:"ShippingCost"`
	TotalCost       float64                `json:"TotalCost"`
	OrderItems      []billbeeOrderItems    `json:"OrderItems"`
	Currency        string                 `json:"Currency"`
	SellerComment   string                 `json:"SellerComment"`
	Tags            []string               `json:"Tags"`
}

func ToOrderId(id int64) string {
	return "CC-" + fmt.Sprintf("%06d", id)
}

func ToUzOrderId(id int64) string {
	return "UZ-" + fmt.Sprintf("%06d", id)
}

func newBillbeeOrderBody(order *model.Order) billbeeBody {
	// Payment type.
	var payment int
	if order.Payment == "banktransfer" {
		payment = 1
	} else {
		payment = 3 // PayPal
	}
	// Tags
	var tags []string
	if order.Reseller {
		tags = append(tags, "reseller")
	}
	if order.SlowFoodMember {
		tags = append(tags, "slow_food_mitglied")
	}
	// Instantiate JSON body
	body := billbeeBody{
		CreatedAt:   order.Date,
		OrderNumber: ToOrderId(order.ID),
		InvoiceAddress: billbeeInvoiceAddress{
			BillbeeID:   0,
			FirstName:   order.FirstNameInvoice,
			LastName:    order.LastNameInvoice,
			Company:     order.CompanyInvoice,
			Street:      order.AddressStreetInvoice,
			HouseNumber: order.AddressStreetNoInvoice,
			Zip:         order.AddressCodeInvoice,
			City:        order.AddressCityInvoice,
			Country:     order.AddressCountryInvoice,
			Email:       order.Email,
		},
		ShippingAddress: billbeeShippingAddress{
			FirstName:   order.FirstNameDelivery,
			LastName:    order.LastNameDelivery,
			Company:     order.CompanyDelivery,
			Street:      order.AddressStreetDelivery,
			HouseNumber: order.AddressStreetNoDelivery,
			Zip:         order.AddressCodeDelivery,
			City:        order.AddressCityDelivery,
			Country:     order.AddressCountryDelivery,
			Email:       order.Email,
		},
		PaymentMethod: payment,
		ShippingCost:  0,
		TotalCost:     model.ComputePrice(order),
		OrderItems: []billbeeOrderItems{{
			Product: billbeeProduct{
				Title:     "Calendarium Culinarium",
				BillbeeID: 200000000711626,
			},
			Quantity:   order.Amount,
			TotalPrice: model.ComputePrice(order),
		}},
		Currency: "EUR",
		//Seller: billbeeSeller{
		//	Platform:        "Manuell",
		//	BillbeeShopName: "Calendarium Culinarium",
		//	BillbeeShopID:   20000000000025044,
		//	Email:           "hallo@calendariumculinarium.de",
		//},
		SellerComment: order.Message,
		Tags:          tags,
	}
	return body
}

func newBillbeeUzOrderBody(order *model.Order, convivium string) billbeeBody {
	payment := 6 // Gutschein
	var tags []string
	if order.Reseller {
		tags = append(tags, "reseller")
	}
	if order.SlowFoodMember {
		tags = append(tags, "slow_food_mitglied")
	}
	tags = append(tags, convivium)
	tags = append(tags, "UZ")
	body := billbeeBody{
		CreatedAt:   order.Date,
		OrderNumber: ToUzOrderId(order.ID),
		InvoiceAddress: billbeeInvoiceAddress{
			BillbeeID:   0,
			FirstName:   order.FirstNameInvoice,
			LastName:    order.LastNameInvoice,
			Company:     order.CompanyInvoice,
			Street:      order.AddressStreetInvoice,
			HouseNumber: order.AddressStreetNoInvoice,
			Zip:         order.AddressCodeInvoice,
			City:        order.AddressCityInvoice,
			Country:     order.AddressCountryInvoice,
			Email:       order.Email,
		},
		ShippingAddress: billbeeShippingAddress{
			FirstName:   order.FirstNameDelivery,
			LastName:    order.LastNameDelivery,
			Company:     order.CompanyDelivery,
			Street:      order.AddressStreetDelivery,
			HouseNumber: order.AddressStreetNoDelivery,
			Zip:         order.AddressCodeDelivery,
			City:        order.AddressCityDelivery,
			Country:     order.AddressCountryDelivery,
			Email:       order.Email,
		},
		PaymentMethod: payment,
		ShippingCost:  0,
		TotalCost:     0,
		OrderItems: []billbeeOrderItems{{
			Product: billbeeProduct{
				Title:     "Calendarium Culinarium",
				BillbeeID: 200000000711626,
			},
			Quantity:   order.Amount,
			TotalPrice: 0,
		}},
		Currency: "EUR",
		//Seller: billbeeSeller{
		//	Platform:        "Manuell Shop",
		//	BillbeeShopName: "SFD UZ",
		//	//BillbeeShopID:   20000000000025044,
		//	BillbeeShopID:   20000000000030479,
		//	Email:           "hallo@calendariumculinarium.de",
		//},
		SellerComment: order.Message,
		Tags:          tags,
	}
	return body
}

// BillbeeHandler can forward orders to billbee.
type BillbeeHandler struct {
	apiKey          string
	authUsername    string
	authPassword    string
	url             string
	mutex           sync.Mutex
	lastRequestTime time.Time
	Emailer         *Emailer
	destEmails      []string
}

// NewBillbeeHandler instantiates a new forwarder.
func NewBillbeeHandler(apiKey string, authUsername string, authPassword string, url string) *BillbeeHandler {
	var handler BillbeeHandler
	handler.apiKey = apiKey
	handler.authUsername = authUsername
	handler.authPassword = authPassword
	handler.url = url
	handler.lastRequestTime = time.Now()
	handler.Emailer = nil
	return &handler
}

func (billbee *BillbeeHandler) AttachEmailer(emailAddr string, emailPassword string, smtpHost string, smtpPort string, destEmails []string) {
	billbee.Emailer = NewEmailer(emailAddr, emailPassword, smtpHost, smtpPort)
	billbee.destEmails = destEmails
}

// ForwardOrder forwards an order to billbee.
func (billbee *BillbeeHandler) ForwardOrder(order *model.Order) (string, error) {
	billbee.mutex.Lock()
	defer billbee.mutex.Unlock()
	timeDifference := time.Now().Sub(billbee.lastRequestTime)
	if timeDifference.Milliseconds() < 500 {
		wantedDifference := time.Duration(500 - timeDifference.Milliseconds())
		timeToSleep := wantedDifference * time.Millisecond
		log.Print("billbee handler waiting for " + strconv.Itoa(int(timeToSleep.Milliseconds())) + " ms... ")
		time.Sleep(timeToSleep)
	}
	billbee.lastRequestTime = time.Now()
	jsonContent, err := json.Marshal(newBillbeeOrderBody(order))
	if err != nil {
		if billbee.Emailer != nil {
			err2 := billbee.Emailer.SendEmail(billbee.destEmails, "Fehler beim Bestellung erstellen", err.Error()+"\r\n\r\nBei Bestellung mit ID "+strconv.Itoa(int(order.ID))+"\r\nHier Bestelldetails einsehen: https://calendariumculinarium.de/api/orders")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		return "", err
	}
	request, err := http.NewRequest("POST", billbee.url, bytes.NewBuffer(jsonContent))
	if err != nil {
		if billbee.Emailer != nil {
			err2 := billbee.Emailer.SendEmail(billbee.destEmails, "Fehler beim Bestellung erstellen", err.Error()+"\r\n\r\nBei Bestellung mit ID "+strconv.Itoa(int(order.ID))+"\r\nHier Bestelldetails einsehen: https://calendariumculinarium.de/api/orders")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		return "", err
	}

	request.SetBasicAuth(billbee.authUsername, billbee.authPassword)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Billbee-Api-Key", billbee.apiKey)

	client := &http.Client{}
	response, err := client.Do(request)
	fmt.Println("printing request")
	fmt.Println(request)
	fmt.Println("printing json")
	fmt.Println(newBillbeeOrderBody(order))
	fmt.Println("printing response")
	fmt.Println(response)
	if err != nil {
		if billbee.Emailer != nil {
			err2 := billbee.Emailer.SendEmail(billbee.destEmails, "Fehler beim Bestellung weiterleiten", err.Error()+"\r\n\r\nBei Bestellung mit ID "+strconv.Itoa(int(order.ID))+"\r\nHier Bestelldetails einsehen: https://calendariumculinarium.de/api/orders")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		return "", err
	}

	if response.StatusCode != 201 {
		errorString := "Billbee returned HTTP status: " + response.Status
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			errorString = errorString + " and error reading response body: " + err.Error()
		} else {
			errorString = errorString + " with error message: " + string(body)
		}
		if billbee.Emailer != nil {
			err2 := billbee.Emailer.SendEmail(billbee.destEmails, "Fehler bei billbee", errorString+"\r\n\r\nBei Bestellung mit ID "+strconv.Itoa(int(order.ID))+"\r\nHier Bestelldetails einsehen: https://calendariumculinarium.de/api/orders")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		return "", errors.New(errorString)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		if billbee.Emailer != nil {
			err2 := billbee.Emailer.SendEmail(billbee.destEmails, "Fehler beim Response lesen", err.Error()+"\r\n\r\nBei Bestellung mit ID "+strconv.Itoa(int(order.ID))+"\r\nHier Bestelldetails einsehen: https://calendariumculinarium.de/api/orders")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		return "", err
	}
	return string(body), nil

	//return "", nil
}

// ForwardOrder forwards an Unterstuetzer order to billbee.
func (billbee *BillbeeHandler) ForwardUzOrder(order *model.Order, convivium string) (string, error) {
	billbee.mutex.Lock()
	defer billbee.mutex.Unlock()
	timeDifference := time.Now().Sub(billbee.lastRequestTime)
	if timeDifference.Milliseconds() < 500 {
		wantedDifference := time.Duration(500 - timeDifference.Milliseconds())
		timeToSleep := wantedDifference * time.Millisecond
		log.Print("billbee handler waiting for " + strconv.Itoa(int(timeToSleep.Milliseconds())) + " ms... ")
		time.Sleep(timeToSleep)
	}
	billbee.lastRequestTime = time.Now()
	jsonContent, err := json.Marshal(newBillbeeUzOrderBody(order, convivium))
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest("POST", billbee.url, bytes.NewBuffer(jsonContent))
	if err != nil {
		return "", err
	}

	request.SetBasicAuth(billbee.authUsername, billbee.authPassword)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Billbee-Api-Key", billbee.apiKey)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 201 {
		errorString := "Billbee returned HTTP status: " + response.Status
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			errorString = errorString + " and error reading response body: " + err.Error()
		} else {
			errorString = errorString + " with error message: " + string(body)
		}
		return "", errors.New(errorString)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil

	//return "", nil
}
