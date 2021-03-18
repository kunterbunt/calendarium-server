package model

import "errors"

// InvalidID corresponds to the ID that is returned when something is *not* found in the database.
var InvalidID = -1

// Product database entry.
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Shipping    float64 `json:"shipping"`
}

// Order database entry.
type Order struct {
	ID                      int64  `json:"id"`
	ProductID               int    `json:"product_id"`
	Amount                  int    `json:"amount"`
	Date                    string `json:"date"`
	CompanyInvoice	        string `json:"company_invoice"`
	FirstNameInvoice        string `json:"first_name_invoice"`
	LastNameInvoice         string `json:"last_name_invoice"`
	CompanyDelivery	        string `json:"company_delivery"`
	FirstNameDelivery       string `json:"first_name_delivery"`
	LastNameDelivery        string `json:"last_name_delivery"`
	Email                   string `json:"email"`
	AddressStreetInvoice    string `json:"address_street_invoice"`
	AddressStreetNoInvoice  string `json:"address_street_no_invoice"`
	AddressCodeInvoice      string `json:"address_code_invoice"`
	AddressCityInvoice      string `json:"address_city_invoice"`
	AddressCountryInvoice   string `json:"address_country_invoice"`
	AddressStreetDelivery   string `json:"address_street_delivery"`
	AddressStreetNoDelivery string `json:"address_street_no_delivery"`
	AddressCodeDelivery     string `json:"address_code_delivery"`
	AddressCityDelivery     string `json:"address_city_delivery"`
	AddressCountryDelivery  string `json:"address_country_delivery"`
	Payment                 string `json:"payment"`
	Premium                 string `json:"premium"`
	Reseller                bool   `json:"is_reseller"`
	SlowFoodMember          bool   `json:"slow_food_member"`
	AgreesAGB               bool   `json:"agrees_agb"`
	AgreesPrivacy           bool   `json:"agrees_data_privacy"`
	Message                 string `json:"message"`
	BillbeeResponse         string `json:"billbee_api_response"`
}

// ComputePrice applies our discount model.
func ComputePrice(order *Order) float64 {
	discount := 0.0
	if order.Amount < 3 {
		discount = 0.0
	} else if order.Amount < 5 {
		discount = .1
	} else if order.Amount < 50 {
		discount = 0.15
	} else {
		discount = 0.2
	}
	priceBeforeDiscount := float64(order.Amount) * 20.0
	priceAfterDiscount := priceBeforeDiscount * (1.0 - discount)
	return priceAfterDiscount
}

// VerifyOrder verifies that an order is valid.
func VerifyOrder(order *Order) error {
	if order.Amount <= 0 {
		return errors.New("Bitte bestellen Sie mindestens ein Produkt!")
	}
	if order.Email == "" { // @todo verify email integrity
		return errors.New("Bitte geben Sie gültige Emailadresse an!")
	}
	// Invoice.
	if order.FirstNameInvoice == "" || order.LastNameInvoice == "" {
		return errors.New("Bitte geben Sie einen Namen an (Rechnungsanschrift)!")
	}
	if order.AddressStreetInvoice == "" {
		return errors.New("Bitte geben Sie eine gültigen Straßennamen an (Rechnungsanschrift)!")
	}
	if order.AddressStreetNoInvoice == "" {
		return errors.New("Bitte geben Sie eine gültige Straßennummer an (Rechnungsanschrift)!")
	}
	if order.AddressCodeInvoice == "" {
		return errors.New("Bitte geben Sie eine gültige Postleitzahl an (Rechnungsanschrift)!")
	}
	if order.AddressCityInvoice == "" {
		return errors.New("Bitte geben Sie eine gültige Stadt an (Rechnungsanschrift)!")
	}
	if order.AddressCountryInvoice == "" {
		return errors.New("Bitte geben Sie ein gültiges Land an (Rechnungsanschrift)!")
	}
	// Delivery.
	if order.FirstNameDelivery == "" || order.LastNameDelivery == "" {
		return errors.New("Bitte geben Sie einen Namen an (Versandanschrift)!")
	}
	if order.AddressStreetDelivery == "" {
		return errors.New("Bitte geben Sie eine gültigen Straßennamen an (Versandanschrift)!")
	}
	if order.AddressStreetNoDelivery == "" {
		return errors.New("Bitte geben Sie eine gültige Straßennummer an (Versandanschrift)!")
	}
	if order.AddressCodeDelivery == "" {
		return errors.New("Bitte geben Sie eine gültige Postleitzahl an (Versandanschrift)!")
	}
	if order.AddressCityDelivery == "" {
		return errors.New("Bitte geben Sie eine gültige Stadt an (Versandanschrift)!")
	}
	if order.AddressCountryDelivery == "" {
		return errors.New("Bitte geben Sie ein gültiges Land an (Versandanschrift)!")
	}
	// Misc.
	if order.Payment == "" || (order.Payment != "banktransfer" && order.Payment != "paypal") {
		return errors.New("Bitte wählen Sie eine gültige Zahlart aus (banktransfer oder paypal)!")
	}
	if order.AgreesAGB == false {
		return errors.New("Sie müssen für eine Bestellung die AGBs unter https://calendariumculinarium.de/agb akzeptieren!")
	}
	if order.AgreesPrivacy == false {
		return errors.New("Sie müssen für eine Bestellung die Datenschutzerklärung unter https://calendariumculinarium.de/datenschutz akzeptieren!")
	}
	return nil
}
