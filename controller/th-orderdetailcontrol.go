package controller

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"web-project/models"

	"github.com/gin-gonic/gin"
)

// CreateOrderDetail_th handles the creation of a new order detail in Thai
func CreateOrderDetail_th(c *gin.Context, db *sql.DB) {
	var orderDetail_th models.Order_detail_th

	if db == nil {
		log.Fatalf("DB connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if err := c.ShouldBindJSON(&orderDetail_th); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert array of topping to a comma-separated string
	toppings := strings.Join(orderDetail_th.Topping_name_th, ",")

	// Insert order detail into the database
	insertQuery := "INSERT INTO order_detail (Order_id, Size_name_th, Flavor_name_th, Topping_name_th, Sauce_name_th) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(insertQuery, orderDetail_th.Order_id, orderDetail_th.Size_name_th, orderDetail_th.Flavor_name_th, toppings, orderDetail_th.Sauce_name_th)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting data"})
		return
	}

	// Decrease the stock of size
	_, err = db.Exec("UPDATE size SET Size_Stock = Size_Stock - 1 WHERE Size_name_th = ?", orderDetail_th.Size_name_th)
	if err != nil {
		log.Printf("Error updating size stock: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating size stock"})
		return
	}

	// Decrease the stock of flavor
	_, err = db.Exec("UPDATE flavor SET Flavor_Stock = Flavor_Stock - 1 WHERE Flavor_name_th = ?", orderDetail_th.Flavor_name_th)
	if err != nil {
		log.Printf("Error updating flavor stock: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating flavor stock"})
		return
	}

	// Decrease the stock of toppings
	for _, t := range orderDetail_th.Topping_name_th {
		_, err = db.Exec("UPDATE topping SET Topping_Stock = Topping_Stock - 1 WHERE Topping_name_th = ?", t)
		if err != nil {
			log.Printf("Error updating topping stock: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating topping stock"})
			return
		}
	}

	// Decrease the stock of sauce
	_, err = db.Exec("UPDATE sauce SET Sauce_Stock = Sauce_Stock - 1 WHERE Sauce_name_th = ?", orderDetail_th.Sauce_name_th)
	if err != nil {
		log.Printf("Error updating sauce stock: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating sauce stock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order detail created successfully"})
}

// GetOrderDetail_th retrieves an order detail in Thai by its ID
func GetOrderDetail_th(c *gin.Context, db *sql.DB) {
	detailID := c.Param("id")

	if db == nil {
		log.Fatalf("DB connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	var orderDetail models.Order_detail_th
	var toppings string
	err := db.QueryRow("SELECT Size_name_th, Flavor_name_th, Topping_name_th, Sauce_name_th FROM order_detail WHERE Order_id = ?", detailID).Scan(&orderDetail.Size_name_th, &orderDetail.Flavor_name_th, &toppings, &orderDetail.Sauce_name_th)
	if err != nil {
		log.Printf("Error querying data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying data"})
		return
	}

	toppingSlice := strings.Split(toppings, ",")

	// Calculate the total price
	totalPrice, err := calculateTotalPrice_th(db, orderDetail.Size_name_th, orderDetail.Flavor_name_th, toppingSlice, orderDetail.Sauce_name_th)
	if err != nil {
		log.Printf("Error calculating total price: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating total price"})
		return
	}

	orderDetail.Sum_Price = totalPrice

	// Update the Sum_Price in the table
	updateQuery := "UPDATE order_detail SET Sum_Price = ? WHERE Order_id = ?"
	_, err = db.Exec(updateQuery, orderDetail.Sum_Price, detailID)
	if err != nil {
		log.Printf("Error updating Sum_Price: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating Sum_Price"})
		return
	}

	c.JSON(http.StatusOK, orderDetail)
}

// calculateTotalPrice_th calculates the total price of an order in Thai
func calculateTotalPrice_th(db *sql.DB, size, flavor string, toppings []string, sauce string) (int, error) {
	var sizePrice, flavorPrice, saucePrice int
	var toppingPrice int = 0

	// Retrieve the price of each component
	err := db.QueryRow("SELECT Size_price FROM size WHERE Size_name_th = ?", size).Scan(&sizePrice)
	if err != nil {
		return 0, err
	}

	err = db.QueryRow("SELECT Flavor_price FROM flavor WHERE Flavor_name_th = ?", flavor).Scan(&flavorPrice)
	if err != nil {
		return 0, err
	}

	err = db.QueryRow("SELECT Sauce_price FROM sauce WHERE Sauce_name_th = ?", sauce).Scan(&saucePrice)
	if err != nil {
		return 0, err
	}

	// Calculate the price of toppings
	for _, t := range toppings {
		var price int
		err = db.QueryRow("SELECT Topping_price FROM topping WHERE Topping_name_th = ?", t).Scan(&price)
		if err != nil {
			return 0, err
		}
		toppingPrice += price
	}

	// Calculate the total price
	totalPrice := sizePrice + flavorPrice + toppingPrice + saucePrice

	return totalPrice, nil
}

// GetOrderDetails_th retrieves all order details in Thai
func GetOrderDetails_th(c *gin.Context, db *sql.DB) {
	if db == nil {
		log.Fatalf("DB connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	rows, err := db.Query("SELECT * FROM order_detail")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying data"})
		return
	}
	defer rows.Close()

	var orderDetails []models.Order_detail_th
	for rows.Next() {
		var orderDetail models.Order_detail_th
		err := rows.Scan(&orderDetail.Order_id, &orderDetail.Size_name_th, &orderDetail.Flavor_name_th, &orderDetail.Topping_name_th, &orderDetail.Sauce_name_th, &orderDetail.Sum_Price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning data"})
			return
		}
		orderDetails = append(orderDetails, orderDetail)
	}

	c.JSON(http.StatusOK, orderDetails)
}

// UpdateOrderDetail_th updates an existing order detail in Thai
func UpdateOrderDetail_th(c *gin.Context, db *sql.DB) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var orderDetail models.Order_detail_th

	if err := c.ShouldBindJSON(&orderDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the database
	updateQuery := "UPDATE order_detail SET Size_name_th=?, Flavor_name_th=?, Topping_name_th=?, Sauce_name_th=?, Sum_Price=? WHERE Order_id=?"
	_, err := db.Exec(updateQuery, orderDetail.Size_name_th, orderDetail.Flavor_name_th, strings.Join(orderDetail.Topping_name_th, ","), orderDetail.Sauce_name_th, orderDetail.Sum_Price, id)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order detail updated successfully"})
}

// DeleteOrderDetail_th deletes an existing order detail in Thai
func DeleteOrderDetail_th(c *gin.Context, db *sql.DB) {
	detailID := c.Param("id")

	if db == nil {
		log.Fatalf("DB connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	deleteQuery := "DELETE FROM order_detail WHERE Order_id = ?"
	_, err := db.Exec(deleteQuery, detailID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order detail deleted successfully"})
}
