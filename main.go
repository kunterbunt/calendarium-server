package main

import (
	"encoding/csv"
	"fmt"
	"github.com/kunterbunt/calendarium-server/controller"
	"github.com/kunterbunt/calendarium-server/model"
	"log"
	"os"
	"strconv"
	"time"

	"strings"
)

func forwardMissingOrdersToBillbee(server *controller.Server) {
	orders, err := model.GetOrders(server.Db, &server.Mutex)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, order := range orders {
		if strings.Contains(order.BillbeeResponse, "Billbee returned HTTP status: 500") && order.FirstNameDelivery != "Test" {
			fmt.Println(order.FirstNameDelivery + " " + order.LastNameDelivery)
			billbeeResponse, err := server.BillbeeForwarder.ForwardOrder(&order)
			if err != nil {
				log.Println("\tBillbee forwarding error: " + err.Error())
				billbeeResponse = err.Error()
				continue
			}
			err = model.AddBillbeeResponseToOrder(order.ID, billbeeResponse, server.Db, &server.Mutex)
			if err != nil {
				log.Fatal("\terror while saving billbee response: " + err.Error())
			} else {
				log.Println("\tsaved billbee response.")
			}
		}
	}
}

func sendUzOrders(csvFilename string, server *controller.Server) error {
	file, err := os.Open(csvFilename)
	defer file.Close()
	if err != nil {
		return err
	}
	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}
	for i, line := range lines {
		if (i == 0) {
			continue
		}
		convivium := line[0]
		//typ := line[1]
		company := line[2]
		name2 := line[3]
		name3 := line[4]
		name4 := line[5]
		street := line[6]
		plz := line[7]
		city := line[8]
		country := line[9]
		memberNo := line[10]


		name := name2
		if (name3 != "") {
			name = name + ", " + name3
		}
		if (name4 != "") {
			name = name + ", " + name4
		}

		if (country == "") {
			country = "Deutschland"
		}

		if (i == 665 || i == 962) {
			fmt.Println(strconv.Itoa(i) + "/" + strconv.Itoa(len(lines)))
			order := model.Order{int64(i), 0, 1, time.Now().Format(time.RFC3339), company, "", name, company, "", name, "", street, "", plz, city, country, street, "", plz, city, country, "banktransfer", "", false, true, true, true, "Mitgliedsnummer " + memberNo, ""}
			_, err := server.BillbeeForwarder.ForwardUzOrder(&order, convivium)
			if (err != nil) {
				fmt.Println(err)
			}
		}

	}
	return nil
}

func main() {
	// Open the database.
	if len(os.Args) != 15 {
	//if len(os.Args) != 10 { // with Unterstützer-CSV file
		fmt.Println("Please provide: 1) sqlite file 2) billbee API key 3) billbee auth username 4) billbee auth password 5) billbee url 6) getOrder API BasicAuth username 7) getOrder API BasicAuth password 8) 1 to enable billbee forwarding 9) error email address 10) first error destination email address 11) second error destination email address 12) error email account password 13) error email SMTP host 14) error email SMTP port")
		os.Exit(1)
	}
	databaseFile := os.Args[1]
	db, err := model.OpenDb(databaseFile)
	if err != nil {
		panic(err)
	}
	billbeeAPIKey := os.Args[2]
	billbeeAuthUsername := os.Args[3]
	billbeeAuthPw := os.Args[4]
	billbeeUrl := os.Args[5]
	orderApiAuthUsername := os.Args[6]
	orderApiAuthPassword := os.Args[7]
	var withBillBeeForwarding bool
	if os.Args[8] == "1" {
		withBillBeeForwarding = true
	} else {
		withBillBeeForwarding = false
	}

	server := controller.NewServer(db, orderApiAuthUsername, orderApiAuthPassword)
	if withBillBeeForwarding {
		fmt.Println("Billbee forwarding enabled.")
		server.AttachBillbeeForwarder(billbeeAPIKey, billbeeAuthUsername, billbeeAuthPw, billbeeUrl)
		// Attach emailer to send emails upon error.
		errEmailAddr := os.Args[9]
		errDestEmailAddr := []string{os.Args[10], os.Args[11]}
		errEmailPw := os.Args[12]
		errEmailSmtpHost := os.Args[13]
		errEmailSmtpPort := os.Args[14]
		server.AttachEmailer(errEmailAddr, errEmailPw, errEmailSmtpHost, errEmailSmtpPort, errDestEmailAddr)
	}
	// Create tables if-need-be.
	err = model.Validate(db, &server.Mutex)
	if err != nil {
		panic(err)
	}
	// Create products if-need-be.
	products := model.GetProductsThatShouldExist()
	fmt.Print("Checking database:")
	var newOrOldMsg string
	for _, product := range products {
		productInDb, err := model.GetProduct(db, product.Name, &server.Mutex)
		if err != nil {
			panic(err)
		}
		if productInDb.ID == model.InvalidID {
			err = model.AddProduct(db, &product, &server.Mutex)
			if err != nil {
				panic(err)
			}
			newOrOldMsg = " instantiated new database."
		} else {
			newOrOldMsg = " using existing database."
		}
	}
	fmt.Println(newOrOldMsg)

	//err = sendUzOrders(os.Args[9], server)
	//if err != nil {
	//	panic(err)
	//}

	//forwardMissingOrdersToBillbee(server)

	// Start listening...
	fmt.Println("Listening on port 8000...")
	port := 8000
	server.ListenAndServe(port)
}
