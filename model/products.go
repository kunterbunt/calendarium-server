package model

// GetProductsThatShouldExist returns an array of products that should exist in the database.
func GetProductsThatShouldExist() [1]Product {
	var products [1]Product
	products[0] = Product{InvalidID, "Calendarium Culinarium", "Der Slow Food Youth Saisonkalender", 20.0, 0.0}
	return products
}
