package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"smat/iot/simulation/iot-inventory-management/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"smat/iot/simulation/iot-inventory-management/internal/service"
)

type UIHandler struct {
	deviceService service.DeviceService
	templates     *template.Template
}

func NewUIHandler(deviceService service.DeviceService) *UIHandler {
	// Create function map
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"printf": fmt.Sprintf,
	}

	// Initialize templates
	templates := template.New("").Funcs(funcMap)

	// Try to find the template directory
	templateDir := "web/templates"

	// Check if we're running from a different directory
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// Try alternative paths
		possiblePaths := []string{
			"./web/templates",
			"../web/templates",
			"/src/web/templates", // Docker path
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				templateDir = path
				break
			}
		}
	}

	log.Printf("Using template directory: %s", templateDir)

	// Parse all template files
	templateFiles := []string{}

	// Walk through the template directory and collect all .html files
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			templateFiles = append(templateFiles, path)
			log.Printf("Found template file: %s", path)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking template directory: %v", err)
		// Try using glob patterns as fallback
		patterns := []string{
			filepath.Join(templateDir, "pages", "*.html"),
			filepath.Join(templateDir, "components", "*.html"),
			filepath.Join(templateDir, "layouts", "*.html"),
		}

		for _, pattern := range patterns {
			files, err := filepath.Glob(pattern)
			if err != nil {
				log.Printf("Error with glob pattern %s: %v", pattern, err)
				continue
			}
			templateFiles = append(templateFiles, files...)
		}
	}

	// Parse the collected template files
	if len(templateFiles) > 0 {
		templates, err = templates.ParseFiles(templateFiles...)
		if err != nil {
			log.Fatalf("Failed to parse template files: %v", err)
		}
	} else {
		log.Fatal("No template files found!")
	}

	// Log loaded templates for debugging
	log.Println("Successfully loaded templates:")
	for _, tmpl := range templates.Templates() {
		log.Printf("  - %s", tmpl.Name())
	}

	return &UIHandler{
		deviceService: deviceService,
		templates:     templates,
	}
}

func (h *UIHandler) LoginPage(c *gin.Context) {
	// Check if templates are loaded
	if h.templates == nil {
		log.Printf("Templates not loaded!")
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	// Check if user already has a valid client_id cookie
	clientID, err := c.Cookie("client_id")
	if err == nil && clientID != "" {
		// If client_id exists, redirect to dashboard
		if _, err := uuid.Parse(clientID); err == nil {
			c.Redirect(http.StatusFound, "/ui/dashboard")
			return
		}
	}

	data := gin.H{
		"Title": "Login",
	}

	var buf bytes.Buffer
	err = h.templates.ExecuteTemplate(&buf, "login", data)
	if err != nil {
		log.Printf("Error rendering login template: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func (h *UIHandler) HandleLogin(c *gin.Context) {
	clientID := c.PostForm("client_id")

	// Validate UUID format
	if _, err := uuid.Parse(clientID); err != nil {
		// Return the form with error message
		errorHTML := `
        <form id="login-form" hx-post="/ui/login" hx-target="this" hx-swap="outerHTML" class="space-y-6">
            <div>
                <label for="client_id" class="block text-sm font-medium text-gray-700 mb-2">
                    Client ID
                </label>
                <input type="text"
                       id="client_id"
                       name="client_id"
                       required
                       value="` + clientID + `"
                       placeholder="Enter your client ID (UUID format)"
                       class="w-full px-4 py-3 border border-red-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors">
            </div>
            <div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
                Invalid Client ID format. Please enter a valid UUID.
            </div>
            <button type="submit"
                    class="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors duration-200 flex items-center justify-center">
                <span>Access Dashboard</span>
                <svg class="ml-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6"/>
                </svg>
            </button>
        </form>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(errorHTML))
		return
	}

	// Check if client exists (optional validation)
	clientUUID, _ := uuid.Parse(clientID)
	devices, err := h.deviceService.GetDevicesByClient(c.Request.Context(), clientUUID)
	if err != nil {
		log.Printf("Error fetching devices for client %s: %v", clientID, err)
	}

	if len(devices) == 0 {
		log.Printf("No devices found for client %s, but allowing login", clientID)
	}

	// Set session cookie
	c.SetCookie("client_id", clientID, 3600, "/", "", false, true)

	// Use HX-Redirect header for HTMX
	c.Header("HX-Redirect", "/ui/dashboard")
	c.Status(http.StatusOK)
}

func (h *UIHandler) Dashboard(c *gin.Context) {
	// Check if templates are loaded
	if h.templates == nil {
		log.Printf("Templates not loaded!")
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	clientID, err := c.Cookie("client_id")
	var clientUUID uuid.UUID

	if err != nil || clientID == "" {
		// Redirect to login if no client_id
		c.Redirect(http.StatusFound, "/ui/login")
		return
	}

	clientUUID, err = uuid.Parse(clientID)
	if err != nil {
		// Invalid UUID, redirect to login
		c.Redirect(http.StatusFound, "/ui/login")
		return
	}

	devices, err := h.deviceService.GetDevicesByClient(c.Request.Context(), clientUUID)
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		devices = []*domain.Device{}
	}

	totalSold := 0
	for _, device := range devices {
		totalSold += device.TotalItemSoldCount
	}

	data := gin.H{
		"Title":       "Dashboard",
		"ClientID":    clientID,
		"DeviceCount": len(devices),
		"TotalSold":   totalSold,
	}

	var buf bytes.Buffer
	err = h.templates.ExecuteTemplate(&buf, "dashboard", data)
	if err != nil {
		log.Printf("Error rendering dashboard template: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func (h *UIHandler) GetDevices(c *gin.Context) {
	if h.templates == nil {
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	clientID := c.Param("clientId")

	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid client ID")
		return
	}

	devices, err := h.deviceService.GetDevicesByClient(c.Request.Context(), clientUUID)
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		c.String(http.StatusInternalServerError, "Error fetching devices")
		return
	}

	if len(devices) == 0 {
		emptyHTML := `<div class="col-span-full text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"/>
            </svg>
            <p class="mt-2 text-gray-500">No devices found</p>
        </div>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(emptyHTML))
		return
	}

	var html bytes.Buffer
	for _, device := range devices {
		var cardBuf bytes.Buffer
		err := h.templates.ExecuteTemplate(&cardBuf, "device-card", device)
		if err != nil {
			log.Printf("Error rendering device card: %v", err)
			continue
		}
		html.Write(cardBuf.Bytes())
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", html.Bytes())
}

func (h *UIHandler) GetDeviceModal(c *gin.Context) {
	if h.templates == nil {
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	deviceID := c.Param("deviceId")

	deviceUUID, err := uuid.Parse(deviceID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid device ID")
		return
	}

	device, err := h.deviceService.GetDevice(c.Request.Context(), deviceUUID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching device")
		return
	}

	var buf bytes.Buffer
	err = h.templates.ExecuteTemplate(&buf, "device-modal", device)
	if err != nil {
		log.Printf("Error rendering modal: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering modal: %v", err)
		return
	}

	modalHTML := fmt.Sprintf(`<div x-data="{ modalOpen: true }">%s</div>`, buf.String())
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(modalHTML))
}

func (h *UIHandler) Logout(c *gin.Context) {
	// Clear the cookie
	c.SetCookie("client_id", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/ui/login")
}

func (h *UIHandler) HelloWorldPage(c *gin.Context) {
	if h.templates == nil {
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	data := gin.H{
		"Title": "Hello World",
	}

	var buf bytes.Buffer
	err := h.templates.ExecuteTemplate(&buf, "hello_world", data)
	if err != nil {
		log.Printf("Error rendering hello_world template: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func (h *UIHandler) GoToHelloWorldPage(c *gin.Context) {
	if h.templates == nil {
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	data := gin.H{
		"Title": "Go To Hello World",
	}

	var buf bytes.Buffer
	err := h.templates.ExecuteTemplate(&buf, "go_to_hello_world", data)
	if err != nil {
		log.Printf("Error rendering go_to_hello_world template: %v", err)
		c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
