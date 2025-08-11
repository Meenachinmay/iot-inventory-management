package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
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

	// Parse all templates
	templates, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/**/*.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		// Try alternative paths
		templates, err = template.New("").Funcs(funcMap).ParseGlob("web/templates/*/*.html")
		if err != nil {
			log.Fatalf("Failed to parse templates: %v", err)
		}
	}

	// Log loaded templates for debugging
	log.Println("Loaded templates:")
	for _, tmpl := range templates.Templates() {
		log.Printf("  - %s", tmpl.Name())
	}

	return &UIHandler{
		deviceService: deviceService,
		templates:     templates,
	}
}

func (h *UIHandler) LoginPage(c *gin.Context) {
	data := gin.H{
		"Title": "Login",
	}

	var buf bytes.Buffer
	err := h.templates.ExecuteTemplate(&buf, "login", data)
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

	// Check if client exists
	clientUUID, _ := uuid.Parse(clientID)
	devices, err := h.deviceService.GetDevicesByClient(c.Request.Context(), clientUUID)
	if err != nil {
		log.Printf("Error fetching devices for client %s: %v", clientID, err)
	}

	if len(devices) == 0 {
		log.Printf("No devices found for client %s, but allowing login", clientID)
	}

	// Set session/cookie
	cookie := &http.Cookie{
		Name:     "client_id",
		Value:    clientID,
		MaxAge:   3600,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)

	// For successful login, use HX-Redirect header
	c.Header("HX-Redirect", "/ui/dashboard")
	c.Status(http.StatusOK)
}

func (h *UIHandler) Dashboard(c *gin.Context) {
	clientID, err := c.Cookie("client_id")
	if err != nil {
		c.Redirect(http.StatusFound, "/ui/login")
		return
	}

	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		c.Redirect(http.StatusFound, "/ui/login")
		return
	}

	devices, err := h.deviceService.GetDevicesByClient(c.Request.Context(), clientUUID)
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		devices = []*domain.Device{} // Empty array with correct type
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
	// Use http.Cookie for SameSite attribute consistency
	cookie := &http.Cookie{
		Name:     "client_id",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/ui/login")
}
