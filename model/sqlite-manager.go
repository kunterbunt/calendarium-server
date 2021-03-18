package model

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // init driver
	"sync"
)

// OpenDb opens the database with the given filename and returns a pointer to it.
func OpenDb(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Validate the content of the provided database and creates tables if-need-be.
func Validate(db *sql.DB, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()
	// PRODUCTS table.
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE, description TEXT, price REAL, shipping REAL)")
	if err != nil {
		return err
	}
	statement.Exec()
	// ORDERS table.
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS orders (id INTEGER PRIMARY KEY, product_id INTEGER NOT NULL, amount INTEGER NOT NULL, date INTEGER NOT NULL, first_name_invoice TEXT NOT NULL, last_name_invoice TEXT NOT NULL, first_name_delivery TEXT NOT NULL, last_name_delivery TEXT NOT NULL, email TEXT NOT NULL, address_street_invoice TEXT NOT NULL, address_street_no_invoice TEXT NOT NULL, address_code_invoice TEXT NOT NULL, address_country_invoice TEXT NOT NULL, address_city_invoice TEXT NOT NULL, address_street_delivery TEXT NOT NULL, address_street_no_delivery TEXT NOT NULL, address_code_delivery TEXT NOT NULL, address_city_delivery TEXT NOT NULL, address_country_delivery TEXT NOT NULL, payment TEXT, premium TEXT, is_reseller BOOLEAN, slow_food_member BOOLEAN, agrees_agbs BOOLEAN, agrees_data_privacy BOOLEAN, message TEXT, billbee_api_response TEXT, FOREIGN KEY (product_id) REFERENCES products (id))")
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

// AddProduct adds a product to the database.
func AddProduct(db *sql.DB, product *Product, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()
	statement, err := db.Prepare("INSERT INTO products (name, description, price, shipping) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	statement.Exec(product.Name, product.Description, product.Price, product.Shipping)
	return nil
}

// GetProducts returns all products in the database.
func GetProducts(db *sql.DB, mutex *sync.Mutex) ([]Product, error) {
	mutex.Lock()
	defer mutex.Unlock()
	query := "SELECT id, name, description, price, shipping FROM products"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	products := make([]Product, 0)
	for rows.Next() {
		var product Product
		err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Shipping)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

// GetProduct queries the database for a product with the given name.
// If the returned product's ID is identical to model.InvalidID, then the corresponding product was not found.
func GetProduct(db *sql.DB, name string, mutex *sync.Mutex) (*Product, error) {
	mutex.Lock()
	defer mutex.Unlock()
	statement := "SELECT id, description, price, shipping FROM products WHERE name=?"
	row := db.QueryRow(statement, name)
	product := Product{InvalidID, name, "", 0.0, 0.0}
	err := row.Scan(&product.ID, &product.Description, &product.Price, &product.Shipping)
	switch err {
	case sql.ErrNoRows:
		return &product, nil
	case nil:
		return &product, nil
	default:
		return nil, err
	}
}

// GetProductByID queries the database for a product with the given ID.
// If the returned product's ID is identical to model.InvalidID, then the corresponding product was not found.
func GetProductByID(db *sql.DB, id int, mutex *sync.Mutex) (*Product, error) {
	mutex.Lock()
	defer mutex.Unlock()
	statement := "SELECT id, name, description, price, shipping FROM products WHERE id=?"
	row := db.QueryRow(statement, id)
	product := Product{InvalidID, "", "", 0.0, 0.0}
	err := row.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Shipping)
	switch err {
	case sql.ErrNoRows:
		return &product, nil
	case nil:
		return &product, nil
	default:
		return nil, err
	}
}

// AddOrder adds an order to the database and sets the ID in the order.
func AddOrder(db *sql.DB, order *Order, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	statement, err := db.Prepare("INSERT INTO orders (product_id, amount, date, first_name_invoice, last_name_invoice, first_name_delivery, last_name_delivery, email, address_street_invoice, address_street_no_invoice, address_code_invoice, address_city_invoice, address_country_invoice, address_street_delivery, address_street_no_delivery, address_code_delivery, address_city_delivery, address_country_delivery, payment , premium, is_reseller, slow_food_member, agrees_agbs, agrees_data_privacy, message, billbee_api_response) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	message := order.Message
	if order.CompanyInvoice != "" {
		message = message + " company_invoice='" + order.CompanyInvoice + "'"
	}
	if order.CompanyDelivery != "" {
		message = message + " company_delivery='" + order.CompanyDelivery + "'"
	}
	result, err := statement.Exec(order.ProductID, order.Amount, order.Date, order.FirstNameInvoice, order.LastNameInvoice, order.FirstNameDelivery, order.LastNameDelivery, order.Email, order.AddressStreetInvoice, order.AddressStreetNoInvoice, order.AddressCodeInvoice, order.AddressCityInvoice, order.AddressCountryInvoice, order.AddressStreetDelivery, order.AddressStreetNoDelivery, order.AddressCodeDelivery, order.AddressCityDelivery, order.AddressCountryDelivery, order.Payment, order.Premium, order.Reseller, order.SlowFoodMember, order.AgreesAGB, order.AgreesPrivacy, message, order.BillbeeResponse)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	order.ID = id
	return nil
}

// GetNumOrders returns the number of orders currently saved in the database.
func GetNumOrders(db *sql.DB, mutex *sync.Mutex) (int, error) {
	mutex.Lock()
	defer mutex.Unlock()
	statement, err := db.Prepare("SELECT COUNT(*) FROM orders")
	if err != nil {
		return 0, err
	}
	var n int
	err = statement.QueryRow().Scan(&n)
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return n, nil
	default:
		return 0, err
	}
}

// AddBillbeeResponseToOrder saves the API response from forwarding an order to billbee for later debugging purposes.
func AddBillbeeResponseToOrder(id int64, billbeeResponse string, db *sql.DB, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()
	statement, err := db.Prepare("UPDATE orders SET billbee_api_response = ? where id = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(billbeeResponse, id)
	if err != nil {
		return err
	}
	return nil
}

// GetOrders returns all orders.
func GetOrders(db *sql.DB, mutex *sync.Mutex) ([]Order, error) {
	mutex.Lock()
	defer mutex.Unlock()
	query := "SELECT * FROM orders"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	orders := make([]Order, 0)
	for rows.Next() {
		var order Order
		err = rows.Scan(&order.ID, &order.ProductID, &order.Amount, &order.Date, &order.FirstNameInvoice, &order.LastNameInvoice, &order.FirstNameDelivery, &order.LastNameDelivery, &order.Email, &order.AddressStreetInvoice, &order.AddressStreetNoInvoice, &order.AddressCodeInvoice, &order.AddressCityInvoice, &order.AddressCountryInvoice, &order.AddressStreetDelivery, &order.AddressStreetNoDelivery, &order.AddressCodeDelivery, &order.AddressCityDelivery, &order.AddressCountryDelivery, &order.Payment, &order.Premium, &order.Reseller, &order.SlowFoodMember, &order.AgreesAGB, &order.AgreesPrivacy, &order.Message, &order.BillbeeResponse)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
