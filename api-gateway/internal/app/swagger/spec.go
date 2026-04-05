package swagger

type SwaggerDoc struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Servers    []Server            `json:"servers"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components"`
	Tags       []Tag               `json:"tags"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

type Operation struct {
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Tags        []string              `json:"tags"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name        string                 `json:"name"`
	In          string                 `json:"in"`
	Required    bool                   `json:"required"`
	Schema      map[string]interface{} `json:"schema"`
	Description string                 `json:"description,omitempty"`
}

type RequestBody struct {
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
	Description string               `json:"description,omitempty"`
}

type MediaType struct {
	Schema map[string]interface{} `json:"schema"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
	Schemas         map[string]interface{}    `json:"schemas"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetSwaggerSpec() SwaggerDoc {
	return SwaggerDoc{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       "Winx API Gateway",
			Description: "Complete API documentation for Winx dating platform",
			Version:     "2.0.0",
		},
		Servers: []Server{
			{URL: "http://localhost:3000", Description: "Development server"},
			{URL: "https://api.winx.com", Description: "Production server"},
		},
		Tags: []Tag{
			{Name: "Auth", Description: "Authentication and user management"},
			{Name: "Profile", Description: "User profile management"},
			{Name: "Notifications", Description: "User notifications"},
			{Name: "Password", Description: "Password management"},
			{Name: "System", Description: "System endpoints"},
		},
		Paths: map[string]PathItem{
			"/healthz": {
				Get: &Operation{
					Summary:     "Health check",
					Description: "Check if API gateway is running",
					Tags:        []string{"System"},
					Responses: map[string]Response{
						"200": {Description: "Gateway is healthy"},
					},
				},
			},
			"/api/v1/login": {
				Post: &Operation{
					Summary:     "User login",
					Description: "Authenticate user and get access token",
					Tags:        []string{"Auth"},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"email":    map[string]string{"type": "string", "format": "email"},
										"password": map[string]string{"type": "string"},
									},
									"required": []string{"email", "password"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"200": {
							Description: "Successful login",
							Content: map[string]MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]bool{"type": true},
											"message": map[string]string{"type": "string"},
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"access_token":  map[string]string{"type": "string"},
													"refresh_token": map[string]string{"type": "string"},
													"token_type":    map[string]string{"type": "string"},
													"expires_in":    map[string]string{"type": "integer"},
												},
											},
										},
									},
								},
							},
						},
						"400": {Description: "Invalid request body"},
						"401": {Description: "Invalid credentials"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/register": {
				Post: &Operation{
					Summary:     "User registration",
					Description: "Register a new user account",
					Tags:        []string{"Auth"},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"email":    map[string]string{"type": "string", "format": "email"},
										"password": map[string]string{"type": "string", "minLength": "8"},
									},
									"required": []string{"email", "password"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"201": {Description: "User created successfully"},
						"400": {Description: "Invalid request body"},
						"409": {Description: "Email already exists"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/refresh": {
				Post: &Operation{
					Summary:     "Refresh access token",
					Description: "Get new access token using refresh token",
					Tags:        []string{"Auth"},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"refresh_token": map[string]string{"type": "string"},
									},
									"required": []string{"refresh_token"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "Token refreshed successfully"},
						"401": {Description: "Invalid refresh token"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/check": {
				Post: &Operation{
					Summary:     "Check token validity",
					Description: "Validate access token and get user info",
					Tags:        []string{"Auth"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					Responses: map[string]Response{
						"200": {Description: "Token is valid"},
						"401": {Description: "Invalid or expired token"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/logout": {
				Post: &Operation{
					Summary:     "User logout",
					Description: "Invalidate current access token",
					Tags:        []string{"Auth"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					Responses: map[string]Response{
						"200": {Description: "Logged out successfully"},
						"401": {Description: "Invalid token"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/password/forgot": {
				Post: &Operation{
					Summary:     "Forgot password",
					Description: "Request password reset PIN sent to email",
					Tags:        []string{"Password"},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"email": map[string]string{"type": "string", "format": "email"},
									},
									"required": []string{"email"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "Reset PIN sent to email"},
						"404": {Description: "Email not found"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/password/reset": {
				Post: &Operation{
					Summary:     "Reset password",
					Description: "Reset password using PIN code",
					Tags:        []string{"Password"},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"email":                     map[string]string{"type": "string", "format": "email"},
										"token":                     map[string]string{"type": "string"},
										"new_password":              map[string]string{"type": "string"},
										"new_password_confirmation": map[string]string{"type": "string"},
									},
									"required": []string{"email", "token", "new_password", "new_password_confirmation"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "Password reset successfully"},
						"400": {Description: "Invalid request"},
						"401": {Description: "Invalid or expired token"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/password/change": {
				Post: &Operation{
					Summary:     "Change password",
					Description: "Change password for authenticated user",
					Tags:        []string{"Password"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"password":     map[string]string{"type": "string"},
										"new_password": map[string]string{"type": "string"},
									},
									"required": []string{"password", "new_password"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "Password changed successfully"},
						"400": {Description: "Invalid request"},
						"401": {Description: "Invalid current password"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Auth service unreachable"},
					},
				},
			},
			"/api/v1/profile/me": {
				Get: &Operation{
					Summary:     "Get my profile",
					Description: "Get current user's profile information",
					Tags:        []string{"Profile"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					Responses: map[string]Response{
						"200": {
							Description: "Profile retrieved successfully",
							Content: map[string]MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]bool{"type": true},
											"message": map[string]string{"type": "string"},
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"user_id":    map[string]string{"type": "integer"},
													"email":      map[string]string{"type": "string"},
													"first_name": map[string]string{"type": "string"},
													"birth_date": map[string]string{"type": "string"},
													"gender":     map[string]string{"type": "string"},
													"about_me":   map[string]string{"type": "string"},
													"interests":  map[string]interface{}{"type": "array"},
													"location":   map[string]interface{}{"type": "object"},
												},
											},
										},
									},
								},
							},
						},
						"401": {Description: "Unauthorized"},
						"404": {Description: "Profile not found"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Profile service unreachable"},
					},
				},
			},
			"/api/v1/profile/store": {
				Post: &Operation{
					Summary:     "Create/update profile",
					Description: "Create or update user profile information",
					Tags:        []string{"Profile"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"first_name":    map[string]string{"type": "string"},
										"birth_date":    map[string]string{"type": "string"},
										"gender":        map[string]string{"type": "string"},
										"about_me":      map[string]string{"type": "string"},
										"interest_ids":  map[string]interface{}{"type": "array", "items": map[string]string{"type": "integer"}},
										"interested_in": map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
										"looking_for":   map[string]string{"type": "string"},
										"location": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"city":    map[string]string{"type": "string"},
												"country": map[string]string{"type": "string"},
											},
										},
									},
								},
							},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "Profile updated"},
						"201": {Description: "Profile created"},
						"400": {Description: "Invalid request"},
						"401": {Description: "Unauthorized"},
						"422": {Description: "Validation error"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Profile service unreachable"},
					},
				},
			},
			"/api/v1/profile/photo": {
				Get: &Operation{
					Summary:     "Get profile photo",
					Description: "Get current user's profile photo URL",
					Tags:        []string{"Profile"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					Responses: map[string]Response{
						"200": {
							Description: "Photo retrieved successfully",
							Content: map[string]MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]bool{"type": true},
											"message": map[string]string{"type": "string"},
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"url":        map[string]string{"type": "string"},
													"mime_type":  map[string]string{"type": "string"},
													"size_bytes": map[string]string{"type": "integer"},
												},
											},
										},
									},
								},
							},
						},
						"401": {Description: "Unauthorized"},
						"404": {Description: "Photo not found"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Profile service unreachable"},
					},
				},
			},
			"/api/v1/profile/photo/store": {
				Post: &Operation{
					Summary:     "Upload profile photo",
					Description: "Upload or update profile photo",
					Tags:        []string{"Profile"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					RequestBody: &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"multipart/form-data": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"photo": map[string]interface{}{
											"type":   "string",
											"format": "binary",
										},
									},
									"required": []string{"photo"},
								},
							},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "Photo updated"},
						"201": {Description: "Photo uploaded"},
						"400": {Description: "Invalid file"},
						"401": {Description: "Unauthorized"},
						"415": {Description: "Unsupported media type"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Profile service unreachable"},
					},
				},
			},
			"/api/v1/profile/interests": {
				Get: &Operation{
					Summary:     "List interests",
					Description: "Get list of available interests for profile",
					Tags:        []string{"Profile"},
					Responses: map[string]Response{
						"200": {
							Description: "Interests retrieved successfully",
							Content: map[string]MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]bool{"type": true},
											"message": map[string]string{"type": "string"},
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"type": "object",
													"properties": map[string]interface{}{
														"id":       map[string]string{"type": "integer"},
														"name":     map[string]string{"type": "string"},
														"category": map[string]string{"type": "string"},
														"icon_url": map[string]string{"type": "string"},
													},
												},
											},
										},
									},
								},
							},
						},
						"500": {Description: "Internal server error"},
						"502": {Description: "Profile service unreachable"},
					},
				},
			},
			"/api/v1/profile/location/ip": {
				Get: &Operation{
					Summary:     "Get location by IP",
					Description: "Get geolocation information for an IP address",
					Tags:        []string{"Profile"},
					Parameters: []Parameter{
						{
							Name:        "ip",
							In:          "query",
							Required:    false,
							Description: "IP address to lookup (defaults to client IP)",
							Schema:      map[string]interface{}{"type": "string", "format": "ipv4"},
						},
					},
					Responses: map[string]Response{
						"200": {
							Description: "Location retrieved successfully",
							Content: map[string]MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]bool{"type": true},
											"message": map[string]string{"type": "string"},
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"ip":        map[string]string{"type": "string"},
													"city":      map[string]string{"type": "string"},
													"country":   map[string]string{"type": "string"},
													"latitude":  map[string]string{"type": "number"},
													"longitude": map[string]string{"type": "number"},
												},
											},
										},
									},
								},
							},
						},
						"400": {Description: "Invalid IP address"},
						"500": {Description: "Internal server error"},
						"502": {Description: "GeoIP service unreachable"},
					},
				},
			},
			"/api/v1/notifications": {
				Get: &Operation{
					Summary:     "Get notifications",
					Description: "Get current user's notifications",
					Tags:        []string{"Notifications"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					Responses: map[string]Response{
						"200": {
							Description: "Notifications retrieved successfully",
							Content: map[string]MediaType{
								"application/json": {
									Schema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]bool{"type": true},
											"message": map[string]string{"type": "string"},
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"type": "object",
													"properties": map[string]interface{}{
														"id":         map[string]string{"type": "integer"},
														"type":       map[string]string{"type": "string"},
														"subject":    map[string]string{"type": "string"},
														"body":       map[string]string{"type": "string"},
														"status":     map[string]string{"type": "string"},
														"created_at": map[string]string{"type": "string"},
													},
												},
											},
										},
									},
								},
							},
						},
						"401": {Description: "Unauthorized"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Notification service unreachable"},
					},
				},
			},
			"/api/v1/notifications/{id}": {
				Delete: &Operation{
					Summary:     "Delete notification",
					Description: "Delete a specific notification by ID",
					Tags:        []string{"Notifications"},
					Security:    []map[string][]string{{"bearerAuth": {}}},
					Parameters: []Parameter{
						{
							Name:        "id",
							In:          "path",
							Required:    true,
							Description: "Notification ID",
							Schema:      map[string]interface{}{"type": "integer"},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "Notification deleted"},
						"401": {Description: "Unauthorized"},
						"404": {Description: "Notification not found"},
						"500": {Description: "Internal server error"},
						"502": {Description: "Notification service unreachable"},
					},
				},
			},
		},
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			},
			Schemas: map[string]interface{}{
				"ErrorResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]bool{"type": false},
						"message": map[string]string{"type": "string"},
						"data":    map[string]interface{}{"type": "null"},
					},
				},
				"SuccessResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]bool{"type": true},
						"message": map[string]string{"type": "string"},
						"data":    map[string]interface{}{"type": "object"},
					},
				},
			},
		},
	}
}