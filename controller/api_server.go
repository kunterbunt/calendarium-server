package controller

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/kunterbunt/calendarium-server/model"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Server implements a REST API server.
type Server struct {
	handler          http.Handler
	router           *mux.Router
	Db               *sql.DB
	Mutex            sync.Mutex
	BillbeeForwarder *BillbeeHandler
	BasicAuthUsername string
	BasicAuthPassword string
}

// getProducts gets all products.
func (server *Server) getProducts(writer http.ResponseWriter, request *http.Request) {
	log.Print("getProducts API call...")
	products, err := model.GetProducts(server.Db, &server.Mutex)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(products)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("\tsent reply.")
}

// getOrders gets all orders.
func (server *Server) getOrders(writer http.ResponseWriter, request *http.Request) {
	log.Print("getOrders API call...")
	orders, err := model.GetOrders(server.Db, &server.Mutex)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(orders)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("\tsent reply.")
}

// getProduct gets a single product.
func (server *Server) getProduct(writer http.ResponseWriter, request *http.Request) {
	log.Print("getProduct API call...")
	params := mux.Vars(request)
	product, err := model.GetProduct(server.Db, params["name"], &server.Mutex)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(product)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	log.Print("\tQuery for product with name '" + params["name"] + "'.")
}

// createOrder places an order.
func (server *Server) createOrder(writer http.ResponseWriter, request *http.Request) {
	log.Print("createOrder API call...")
	var order model.Order
	err := json.NewDecoder(request.Body).Decode(&order)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	order.Date = time.Now().Format(time.RFC3339)

	// Check that product exists.
	product, err := model.GetProductByID(server.Db, order.ProductID, &server.Mutex)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if product.ID == model.InvalidID {
		log.Println("\tinvalid product ID.")
		http.Error(writer, "Bitte wählen Sie ein existierendes Produkt.", http.StatusBadRequest)
		return
	}
	err = model.VerifyOrder(&order)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	err = model.AddOrder(server.Db, &order, &server.Mutex)
	if err != nil {
		log.Println("\tError: " + err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Placed order: %+v", order)

	if server.BillbeeForwarder != nil {
		billbeeResponse, err := server.BillbeeForwarder.ForwardOrder(&order)
		additionalErr := ""
		if err != nil {
			log.Println("\tBillbee forwarding error: " + err.Error())
			billbeeResponse = err.Error()
			additionalErr = "Deine Bestellung ist eingegangen, aber es gab einen Fehler beim Weiterleiten an unsere Bestellverwaltung. Bitte teil uns diesen Fehler mit, damit wir ihn beseitigen können. Der Fehler: " + err.Error()
		}
		err = model.AddBillbeeResponseToOrder(order.ID, billbeeResponse, server.Db, &server.Mutex)
		if err != nil {
			log.Fatal("\terror while saving billbee response: " + additionalErr + err.Error())
		} else {
			log.Println("\tsaved billbee response.")
		}
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write([]byte("Vielen Dank für Deine Bestellung mit Bestellnr. '" + ToOrderId(order.ID) + "'."))
	//_, err = writer.Write([]byte("Vielen Dank für Deine Bestellung. " + additionalErr))
	if err != nil {
		panic(err)
	}
}

// BasicAuth from https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
func BasicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

// NewServer instantiates a new API server.
func NewServer(db *sql.DB, BasicAuthUsername string, BasicAuthPassword string) *Server {
	var server Server
	server.router = mux.NewRouter()
	server.Db = db
	server.BasicAuthUsername = BasicAuthUsername
	server.BasicAuthPassword = BasicAuthPassword
	// Init handlers.
	server.router.HandleFunc("/api/products", server.getProducts).Methods("GET")
	server.router.HandleFunc("/api/products/{id}", server.getProduct).Methods("GET")
	server.router.HandleFunc("/api/orders", server.createOrder).Methods("POST")
	server.router.HandleFunc("/api/orders", BasicAuth(server.getOrders, server.BasicAuthUsername, server.BasicAuthPassword, "Please enter your username and password for this site"))

	server.handler = cors.Default().Handler(server.router)
	return &server
}

// ListenAndServe starts the HTTP server.
func (server *Server) ListenAndServe(port int) {
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), server.handler))
}

func (server *Server) AttachBillbeeForwarder(billbeeAPIKey string, billbeeAuthUsername string, billbeeAuthPw string, billbeeUrl string) {
	server.BillbeeForwarder = NewBillbeeHandler(billbeeAPIKey, billbeeAuthUsername, billbeeAuthPw, billbeeUrl)
}

func (server *Server) AttachEmailer(emailAddr string, emailPassword string, smtpHost string, smtpPort string, destEmails []string) {
	if server.BillbeeForwarder != nil {
		server.BillbeeForwarder.AttachEmailer(emailAddr, emailPassword, smtpHost, smtpPort, destEmails)
	} else {
		panic("Called AttachEmailer before AttachBillbeeForwarder!")
	}
}