package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"go-fiber-api/config"
	"go-fiber-api/models"
	"go-fiber-api/routes"
	"go-fiber-api/seed"
)

// setupApp initializes Fiber app and Mongo connection for tests
func setupApp(t *testing.T) *fiber.App {
	t.Helper()
	os.Setenv("MONGO_URL", "mongodb://admin:cr969bp6x6@localhost:27017")
	os.Setenv("MONGO_NAME", "test")
	os.Setenv("JWT_SECRET", "test")
	os.Setenv("PORT", "4000")

	config.ConnectDB()

	// clean collections before each test
	collections := []string{"users", "products", "invoices", "store_settings", "counters"}
	for _, col := range collections {
		config.DB.Collection(col).Drop(context.TODO())
	}

	seed.SeedAdminUser()
	seed.SeedStoreSettings()

	app := fiber.New()
	routes.Setup(app, config.DB)
	return app
}

func login(t *testing.T, app *fiber.App, username, password string) string {
	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.Token
}

func TestUserAndAdminFlow(t *testing.T) {
	app := setupApp(t)

	adminToken := login(t, app, "admin", "admin123")
	userToken := login(t, app, "user", "user123")

	// === Create product as admin ===
	createBody := `{"name":"Test Product","price":1000}`
	req := httptest.NewRequest("POST", "/api/products/", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// === Get product list and capture ID ===
	req = httptest.NewRequest("GET", "/api/products/", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var list struct {
		Data struct {
			Products []models.Product `json:"products"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Data.Products) == 0 {
		t.Fatal("no products returned")
	}
	prodID := list.Data.Products[0].ID.Hex()

	// === Create invoice as user ===
	invoiceBody := fmt.Sprintf(`{"items":[{"productId":"%s","name":"Test Product","quantity":2,"price":1000}]}`, prodID)
	req = httptest.NewRequest("POST", "/api/invoices/", bytes.NewBufferString(invoiceBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var created struct {
		Data models.Invoice `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&created)
	invID := created.Data.ID.Hex()

	// === User cannot delete invoice ===
	req = httptest.NewRequest("DELETE", "/api/invoices?id="+invID, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	// === Admin deletes invoice ===
	req = httptest.NewRequest("DELETE", "/api/invoices?id="+invID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// === Admin deletes product (AdminOnly) ===
	req = httptest.NewRequest("DELETE", "/api/products?id="+prodID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// === User cannot delete product ===
	seed.SeedAdminUser() // reseed product for user delete test
	createBody = `{"name":"Temp","price":1}`
	req = httptest.NewRequest("POST", "/api/products/", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	app.Test(req, -1)

	// fetch ID
	req = httptest.NewRequest("GET", "/api/products/", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = app.Test(req, -1)
	json.NewDecoder(resp.Body).Decode(&list)
	pid := list.Data.Products[len(list.Data.Products)-1].ID.Hex()

	req = httptest.NewRequest("DELETE", "/api/products?id="+pid, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
func TestUserManagement(t *testing.T) {
	app := setupApp(t)

	adminToken := login(t, app, "admin", "admin123")

	// === Create new user as admin ===
	createBody := `{"username":"tempuser","password":"pass123","email":"temp@example.com","role":"user"}`
	req := httptest.NewRequest("POST", "/api/users/", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var created struct {
		Data models.User `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&created)
	userID := created.Data.ID

	// login with created user
	userToken := login(t, app, "tempuser", "pass123")

	// === Update user info ===
	updateBody := fmt.Sprintf(`{"id":"%s","username":"updateduser","email":"upd@example.com","role":"user"}`, userID)
	req = httptest.NewRequest("PUT", "/api/users/", bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// old username should no longer log in
	req = httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"username":"tempuser","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req, -1)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// new username login works
	userToken = login(t, app, "updateduser", "pass123")

	// === Change password ===
	changeBody := `{"old_password":"pass123","new_password":"newpass"}`
	req = httptest.NewRequest("PUT", "/api/users/password", bytes.NewBufferString(changeBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// login with old password fails
	req = httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"username":"updateduser","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req, -1)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// login with new password succeeds
	userToken = login(t, app, "updateduser", "newpass")

	// === user cannot delete other users ===
	req = httptest.NewRequest("DELETE", "/api/users?id="+userID, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	// === admin deletes user ===
	req = httptest.NewRequest("DELETE", "/api/users?id="+userID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// login after deletion fails
	req = httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"username":"updateduser","password":"newpass"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req, -1)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
