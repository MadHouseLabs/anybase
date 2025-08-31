package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/madhouselabs/anybase/internal/collection"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MCPHandler handles MCP protocol requests
type MCPHandler struct {
	collectionService collection.Service
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(collectionService collection.Service) *MCPHandler {
	return &MCPHandler{
		collectionService: collectionService,
	}
}

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleMCPRequest processes MCP protocol requests
func (h *MCPHandler) HandleMCPRequest(c *gin.Context) {
	var req MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32700,
				Message: "Parse error",
				Data:    err.Error(),
			},
			ID: nil,
		})
		return
	}

	// Handle different MCP methods
	switch req.Method {
	case "initialize":
		h.handleInitialize(c, req)
	case "resources/list":
		h.handleResourcesList(c, req)
	case "resources/read":
		h.handleResourcesRead(c, req)
	case "tools/list":
		h.handleToolsList(c, req)
	case "tools/call":
		h.handleToolsCall(c, req)
	default:
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: req.ID,
		})
	}
}

// handleInitialize handles MCP initialization
func (h *MCPHandler) handleInitialize(c *gin.Context, req MCPRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"resources": map[string]bool{
				"list": true,
				"read": true,
			},
			"tools": map[string]bool{
				"call": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "anybase-mcp",
			"version": "1.0.0",
		},
	}

	c.JSON(http.StatusOK, MCPResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	})
}

// handleResourcesList lists available resources (collections and views)
func (h *MCPHandler) handleResourcesList(c *gin.Context, req MCPRequest) {
	// Check authentication type and permissions
	authType, _ := c.Get("auth_type")
	permissions, _ := c.Get("permissions")
	
	resources := []map[string]interface{}{}
	
	if authType == "access_key" {
		// For access keys, only show resources they have permission to access
		perms := permissions.([]string)
		
		// Parse permissions to find accessible collections and views
		collectionPermMap := make(map[string]bool)
		viewPermMap := make(map[string]bool)
		hasAllCollections := false
		hasAllViews := false
		
		for _, perm := range perms {
			parts := strings.Split(perm, ":")
			if len(parts) >= 3 {
				resource := parts[0]
				name := parts[1]
				action := parts[2]
				
				if resource == "collection" {
					if name == "*" {
						hasAllCollections = true
					} else if action == "read" || action == "*" {
						collectionPermMap[name] = true
					}
				} else if resource == "view" {
					if name == "*" {
						hasAllViews = true
					} else if action == "execute" || action == "*" {
						viewPermMap[name] = true
					}
				}
			}
		}
		
		// Only fetch collections the access key has permission for
		if hasAllCollections {
			// Has wildcard permission - can list all collections
			collections, err := h.collectionService.ListCollections(c.Request.Context())
			if err == nil {
				for _, col := range collections {
					// Include schema information in the resource
					schemaInfo := "No schema defined"
					if col.Schema != nil {
						schemaBytes, _ := json.Marshal(col.Schema)
						schemaInfo = string(schemaBytes)
					}
					
					resources = append(resources, map[string]interface{}{
						"uri":         "anybase://collection/" + col.Name,
						"name":        "Collection: " + col.Name,
						"description": fmt.Sprintf("%s\nSchema: %s", col.Description, schemaInfo),
						"mimeType":    "application/json",
					})
				}
			}
		} else if len(collectionPermMap) > 0 {
			// Only fetch specific collections
			for collName := range collectionPermMap {
				col, err := h.collectionService.GetCollection(c.Request.Context(), primitive.NilObjectID, collName)
				if err == nil {
					// Include schema information in the resource
					schemaInfo := "No schema defined"
					if col.Schema != nil {
						schemaBytes, _ := json.Marshal(col.Schema)
						schemaInfo = string(schemaBytes)
					}
					
					resources = append(resources, map[string]interface{}{
						"uri":         "anybase://collection/" + collName,
						"name":        "Collection: " + collName,
						"description": fmt.Sprintf("%s\nSchema: %s", col.Description, schemaInfo),
						"mimeType":    "application/json",
					})
				}
			}
		}
		
		// Only fetch views the access key has permission for
		if hasAllViews {
			// Has wildcard permission - can list all views
			views, err := h.collectionService.ListViews(c.Request.Context())
			if err == nil {
				for _, view := range views {
					resources = append(resources, map[string]interface{}{
						"uri":         "anybase://view/" + view.Name,
						"name":        "View: " + view.Name,
						"description": fmt.Sprintf("%s\nCollection: %s\nPipeline: %s", view.Description, view.Collection, view.Pipeline),
						"mimeType":    "application/json",
					})
				}
			}
		} else if len(viewPermMap) > 0 {
			// Only fetch specific views
			for viewName := range viewPermMap {
				view, err := h.collectionService.GetView(c.Request.Context(), viewName)
				if err == nil {
					resources = append(resources, map[string]interface{}{
						"uri":         "anybase://view/" + viewName,
						"name":        "View: " + viewName,
						"description": fmt.Sprintf("%s\nCollection: %s\nPipeline: %s", view.Description, view.Collection, view.Pipeline),
						"mimeType":    "application/json",
					})
				}
			}
		}
	} else {
		// JWT auth - use user permissions
		// getUserID(c) would be used here if we needed user-specific filtering
		
		// List all accessible collections
		collections, err := h.collectionService.ListCollections(c.Request.Context())
		if err == nil {
			for _, col := range collections {
				// Include schema information in the resource
				schemaInfo := "No schema defined"
				if col.Schema != nil {
					schemaBytes, _ := json.Marshal(col.Schema)
					schemaInfo = string(schemaBytes)
				}
				
				resources = append(resources, map[string]interface{}{
					"uri":         "anybase://collection/" + col.Name,
					"name":        "Collection: " + col.Name,
					"description": fmt.Sprintf("%s\nSchema: %s", col.Description, schemaInfo),
					"mimeType":    "application/json",
				})
			}
		}
		
		// List all accessible views
		views, err := h.collectionService.ListViews(c.Request.Context())
		if err == nil {
			for _, view := range views {
				resources = append(resources, map[string]interface{}{
					"uri":         "anybase://view/" + view.Name,
					"name":        "View: " + view.Name,
					"description": fmt.Sprintf("%s\nCollection: %s\nPipeline: %s", view.Description, view.Collection, view.Pipeline),
					"mimeType":    "application/json",
				})
			}
		}
	}

	c.JSON(http.StatusOK, MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"resources": resources,
		},
		ID: req.ID,
	})
}

// handleResourcesRead reads a specific resource
func (h *MCPHandler) handleResourcesRead(c *gin.Context, req MCPRequest) {
	var params struct {
		URI string `json:"uri"`
	}
	
	if err := json.Unmarshal(req.Params, &params); err != nil {
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
			ID: req.ID,
		})
		return
	}

	// Parse URI to determine resource type
	var resourceType, resourceName string
	if strings.HasPrefix(params.URI, "anybase://collection/") {
		resourceType = "collection"
		resourceName = strings.TrimPrefix(params.URI, "anybase://collection/")
	} else if strings.HasPrefix(params.URI, "anybase://view/") {
		resourceType = "view"
		resourceName = strings.TrimPrefix(params.URI, "anybase://view/")
	} else {
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid resource URI",
			},
			ID: req.ID,
		})
		return
	}

	userID := getUserID(c)
	if authType, _ := c.Get("auth_type"); authType == "access_key" {
		userID = primitive.NilObjectID
	}

	var content interface{}
	
	if resourceType == "collection" {
		col, err := h.collectionService.GetCollection(c.Request.Context(), userID, resourceName)
		if err != nil {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				Error: &MCPError{
					Code:    -32603,
					Message: "Failed to read collection",
					Data:    err.Error(),
				},
				ID: req.ID,
			})
			return
		}
		
		// Get sample data
		query := &models.DataQuery{
			Collection: resourceName,
			Limit:      5,
			UserID:     userID,
		}
		docs, _ := h.collectionService.QueryDocuments(c.Request.Context(), query)
		
		content = map[string]interface{}{
			"name":        col.Name,
			"description": col.Description,
			"schema":      col.Schema,
			"indexes":     col.Indexes,
			"sampleData":  docs,
		}
	} else {
		view, err := h.collectionService.GetView(c.Request.Context(), resourceName)
		if err != nil {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				Error: &MCPError{
					Code:    -32603,
					Message: "Failed to read view",
					Data:    err.Error(),
				},
				ID: req.ID,
			})
			return
		}
		
		// Get view results
		results, _ := h.collectionService.QueryView(c.Request.Context(), userID, resourceName, collection.QueryOptions{
			Limit: 5,
		})
		
		content = map[string]interface{}{
			"name":        view.Name,
			"description": view.Description,
			"collection":  view.Collection,
			"pipeline":    view.Pipeline,
			"sampleData":  results,
		}
	}

	c.JSON(http.StatusOK, MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      params.URI,
					"mimeType": "application/json",
					"text":     content,
				},
			},
		},
		ID: req.ID,
	})
}

// handleToolsList lists available tools based on permissions
func (h *MCPHandler) handleToolsList(c *gin.Context, req MCPRequest) {
	tools := []map[string]interface{}{}
	
	// Check authentication type and permissions
	authType, _ := c.Get("auth_type")
	permissions, _ := c.Get("permissions")
	
	if authType == "access_key" {
		// For access keys, only show tools they have permission to use
		perms := permissions.([]string)
		
		// Parse permissions to find accessible collections and views
		collectionPerms := make(map[string][]string) // collection name -> actions
		viewPerms := make(map[string][]string)       // view name -> actions
		
		for _, perm := range perms {
			parts := strings.Split(perm, ":")
			if len(parts) >= 3 {
				resource := parts[0]
				name := parts[1]
				action := parts[2]
				
				if resource == "collection" {
					if name == "*" {
						// Has wildcard permission - fetch all collections
						collections, err := h.collectionService.ListCollections(c.Request.Context())
						if err == nil {
							for _, col := range collections {
								if collectionPerms[col.Name] == nil {
									collectionPerms[col.Name] = []string{}
								}
								collectionPerms[col.Name] = append(collectionPerms[col.Name], action)
							}
						}
					} else {
						if collectionPerms[name] == nil {
							collectionPerms[name] = []string{}
						}
						collectionPerms[name] = append(collectionPerms[name], action)
					}
				} else if resource == "view" {
					if name == "*" {
						// Has wildcard permission - fetch all views
						views, err := h.collectionService.ListViews(c.Request.Context())
						if err == nil {
							for _, view := range views {
								if viewPerms[view.Name] == nil {
									viewPerms[view.Name] = []string{}
								}
								viewPerms[view.Name] = append(viewPerms[view.Name], action)
							}
						}
					} else {
						if viewPerms[name] == nil {
							viewPerms[name] = []string{}
						}
						viewPerms[name] = append(viewPerms[name], action)
					}
				}
			}
		}
		
		// Create specific tools for each collection
		for collName, actions := range collectionPerms {
			// Get collection details for schema
			col, err := h.collectionService.GetCollection(c.Request.Context(), primitive.NilObjectID, collName)
			if err != nil {
				continue
			}
			
			// Build schema properties from collection schema
			schemaProps := make(map[string]interface{})
			requiredFields := []string{}
			
			if col.Schema != nil {
				for propName, prop := range col.Schema.Properties {
					schemaProps[propName] = map[string]interface{}{
						"type": prop.Type,
						"description": prop.Description,
					}
				}
				requiredFields = col.Schema.Required
			}
			
			// Check permissions and add appropriate tools
			hasRead := false
			hasWrite := false
			hasDelete := false
			
			for _, action := range actions {
				if action == "read" || action == "*" {
					hasRead = true
				}
				if action == "write" || action == "create" || action == "*" {
					hasWrite = true
				}
				if action == "delete" || action == "*" {
					hasDelete = true
				}
			}
			
			if hasRead {
				// Build schema description for the query tool
				schemaDesc := ""
				if col.Schema != nil && len(col.Schema.Properties) > 0 {
					schemaDesc = "\n\nSchema fields:\n"
					for propName, prop := range col.Schema.Properties {
						propType := fmt.Sprintf("%v", prop.Type)
						required := ""
						for _, req := range requiredFields {
							if req == propName {
								required = " (required)"
								break
							}
						}
						desc := prop.Description
						if desc == "" {
							desc = "No description"
						}
						schemaDesc += fmt.Sprintf("- %s: %s%s - %s\n", propName, propType, required, desc)
					}
				}
				
				tools = append(tools, map[string]interface{}{
					"name":        "query_" + collName,
					"description": fmt.Sprintf("Query documents from %s collection. %s%s", collName, col.Description, schemaDesc),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"filter": map[string]interface{}{
								"type":        "object",
								"description": "MongoDB filter query. Example: {\"price\": {\"$gt\": 100}, \"category\": \"electronics\"}",
							},
							"limit": map[string]interface{}{
								"type":        "number",
								"description": "Maximum number of documents to return",
								"default":     10,
							},
							"skip": map[string]interface{}{
								"type":        "number",
								"description": "Number of documents to skip",
								"default":     0,
							},
							"sort": map[string]interface{}{
								"type":        "object",
								"description": "Sort order (e.g., {\"field\": 1} for ascending, {\"field\": -1} for descending)",
							},
						},
					},
				})
			}
			
			if hasWrite {
				// Build schema description for the insert tool  
				schemaDesc := ""
				if col.Schema != nil && len(col.Schema.Properties) > 0 {
					schemaDesc = "\n\nSchema fields:\n"
					for propName, prop := range col.Schema.Properties {
						propType := fmt.Sprintf("%v", prop.Type)
						required := ""
						for _, req := range requiredFields {
							if req == propName {
								required = " (required)"
								break
							}
						}
						desc := prop.Description
						if desc == "" {
							desc = "No description"
						}
						schemaDesc += fmt.Sprintf("- %s: %s%s - %s\n", propName, propType, required, desc)
					}
				}
				
				tools = append(tools, map[string]interface{}{
					"name":        "insert_" + collName,
					"description": fmt.Sprintf("Insert a new document into %s collection. %s%s", collName, col.Description, schemaDesc),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"document": map[string]interface{}{
								"type":        "object",
								"description": "Document to insert with the following schema",
								"properties":  schemaProps,
								"required":    requiredFields,
							},
						},
						"required": []string{"document"},
					},
				})
				
				tools = append(tools, map[string]interface{}{
					"name":        "update_" + collName,
					"description": fmt.Sprintf("Update documents in %s collection", collName),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"filter": map[string]interface{}{
								"type":        "object",
								"description": "MongoDB filter to select documents to update",
							},
							"update": map[string]interface{}{
								"type":        "object",
								"description": "Update operations ($set, $inc, etc.)",
							},
						},
						"required": []string{"filter", "update"},
					},
				})
			}
			
			if hasDelete {
				tools = append(tools, map[string]interface{}{
					"name":        "delete_" + collName,
					"description": fmt.Sprintf("Delete documents from %s collection", collName),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"filter": map[string]interface{}{
								"type":        "object",
								"description": "MongoDB filter to select documents to delete",
							},
						},
						"required": []string{"filter"},
					},
				})
			}
		}
		
		// Create specific tools for each view
		for viewName, actions := range viewPerms {
			// Get view details
			view, err := h.collectionService.GetView(c.Request.Context(), viewName)
			if err != nil {
				continue
			}
			
			for _, action := range actions {
				if action == "execute" || action == "*" {
					// Build more descriptive view information
					viewDesc := fmt.Sprintf("Execute %s view. %s", viewName, view.Description)
					if view.Collection != "" {
						viewDesc += fmt.Sprintf("\nBase collection: %s", view.Collection)
					}
					if view.Filter != nil && len(view.Filter) > 0 {
						filterJSON, _ := json.Marshal(view.Filter)
						viewDesc += fmt.Sprintf("\nBase filter: %s", string(filterJSON))
					}
					if len(view.Fields) > 0 {
						viewDesc += fmt.Sprintf("\nProjected fields: %v", view.Fields)
					}
					if len(view.Pipeline) > 0 {
						pipelineJSON, _ := json.Marshal(view.Pipeline)
						viewDesc += fmt.Sprintf("\nAggregation pipeline: %s", string(pipelineJSON))
					}
					
					tools = append(tools, map[string]interface{}{
						"name":        "execute_view_" + viewName,
						"description": viewDesc,
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"limit": map[string]interface{}{
									"type":        "number",
									"description": "Maximum number of results to return",
									"default":     10,
								},
								"filter": map[string]interface{}{
									"type":        "object",
									"description": "Additional filter to apply on view results",
								},
							},
						},
					})
					break
				}
			}
		}
	} else {
		// JWT auth - show all available collections and views as specific tools
		collections, err := h.collectionService.ListCollections(c.Request.Context())
		if err == nil {
			for _, col := range collections {
				// Build schema properties from collection schema
				schemaProps := make(map[string]interface{})
				requiredFields := []string{}
				
				if col.Schema != nil {
					for propName, prop := range col.Schema.Properties {
						schemaProps[propName] = map[string]interface{}{
							"type": prop.Type,
							"description": prop.Description,
						}
					}
					requiredFields = col.Schema.Required
				}
				
				// Add query tool
				tools = append(tools, map[string]interface{}{
					"name":        "query_" + col.Name,
					"description": fmt.Sprintf("Query documents from %s collection. %s", col.Name, col.Description),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"filter": map[string]interface{}{
								"type":        "object",
								"description": "MongoDB filter query",
							},
							"limit": map[string]interface{}{
								"type":        "number",
								"description": "Maximum number of documents to return",
								"default":     10,
							},
							"skip": map[string]interface{}{
								"type":        "number",
								"description": "Number of documents to skip",
								"default":     0,
							},
							"sort": map[string]interface{}{
								"type":        "object",
								"description": "Sort order",
							},
						},
					},
				})
				
				// Add insert tool
				tools = append(tools, map[string]interface{}{
					"name":        "insert_" + col.Name,
					"description": fmt.Sprintf("Insert a new document into %s collection. %s", col.Name, col.Description),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"document": map[string]interface{}{
								"type":        "object",
								"description": "Document to insert",
								"properties":  schemaProps,
								"required":    requiredFields,
							},
						},
						"required": []string{"document"},
					},
				})
				
				// Add update tool
				tools = append(tools, map[string]interface{}{
					"name":        "update_" + col.Name,
					"description": fmt.Sprintf("Update documents in %s collection", col.Name),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"filter": map[string]interface{}{
								"type":        "object",
								"description": "MongoDB filter to select documents",
							},
							"update": map[string]interface{}{
								"type":        "object",
								"description": "Update operations",
							},
						},
						"required": []string{"filter", "update"},
					},
				})
				
				// Add delete tool
				tools = append(tools, map[string]interface{}{
					"name":        "delete_" + col.Name,
					"description": fmt.Sprintf("Delete documents from %s collection", col.Name),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"filter": map[string]interface{}{
								"type":        "object",
								"description": "MongoDB filter to select documents",
							},
						},
						"required": []string{"filter"},
					},
				})
			}
		}
		
		// Add view tools
		views, err := h.collectionService.ListViews(c.Request.Context())
		if err == nil {
			for _, view := range views {
				tools = append(tools, map[string]interface{}{
					"name":        "execute_view_" + view.Name,
					"description": fmt.Sprintf("Execute %s view. %s", view.Name, view.Description),
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"limit": map[string]interface{}{
								"type":        "number",
								"description": "Maximum number of results",
								"default":     10,
							},
							"filter": map[string]interface{}{
								"type":        "object",
								"description": "Additional filter",
							},
						},
					},
				})
			}
		}
	}

	c.JSON(http.StatusOK, MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"tools": tools,
		},
		ID: req.ID,
	})
}

// handleToolsCall executes a tool
func (h *MCPHandler) handleToolsCall(c *gin.Context, req MCPRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	
	if err := json.Unmarshal(req.Params, &params); err != nil {
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
			ID: req.ID,
		})
		return
	}

	userID := getUserID(c)
	if authType, _ := c.Get("auth_type"); authType == "access_key" {
		userID = primitive.NilObjectID
	}

	var result interface{}
	
	// Parse tool name to determine operation and collection/view
	if strings.HasPrefix(params.Name, "query_") {
		collectionName := strings.TrimPrefix(params.Name, "query_")
		filter, _ := params.Arguments["filter"].(map[string]interface{})
		limit, _ := params.Arguments["limit"].(float64)
		skip, _ := params.Arguments["skip"].(float64)
		// sort, _ := params.Arguments["sort"].(map[string]interface{}) // TODO: implement sort
		
		if limit == 0 {
			limit = 10
		}
		
		query := &models.DataQuery{
			Collection: collectionName,
			Filter:     filter,
			Limit:      int(limit),
			Skip:       int(skip),
			UserID:     userID,
		}
		
		docs, err := h.collectionService.QueryDocuments(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				Error: &MCPError{
					Code:    -32603,
					Message: "Failed to query collection",
					Data:    err.Error(),
				},
				ID: req.ID,
			})
			return
		}
		result = map[string]interface{}{
			"documents": docs,
			"count":     len(docs),
		}
		
	} else if strings.HasPrefix(params.Name, "insert_") {
		collectionName := strings.TrimPrefix(params.Name, "insert_")
		document, _ := params.Arguments["document"].(map[string]interface{})
		
		mutation := &models.DataMutation{
			Collection: collectionName,
			Operation:  "insert",
			Data:       document,
			UserID:     userID,
		}
		
		doc, err := h.collectionService.InsertDocument(c.Request.Context(), mutation)
		if err != nil {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				Error: &MCPError{
					Code:    -32603,
					Message: "Failed to insert document",
					Data:    err.Error(),
				},
				ID: req.ID,
			})
			return
		}
		result = map[string]interface{}{
			"id":      doc.ID.Hex(),
			"success": true,
		}
		
	} else if strings.HasPrefix(params.Name, "update_") {
		collectionName := strings.TrimPrefix(params.Name, "update_")
		filter, _ := params.Arguments["filter"].(map[string]interface{})
		update, _ := params.Arguments["update"].(map[string]interface{})
		
		mutation := &models.DataMutation{
			Collection: collectionName,
			Operation:  "update",
			Filter:     filter,
			Data:       update,
			UserID:     userID,
		}
		
		err := h.collectionService.UpdateDocument(c.Request.Context(), mutation)
		if err != nil {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				Error: &MCPError{
					Code:    -32603,
					Message: "Failed to update documents",
					Data:    err.Error(),
				},
				ID: req.ID,
			})
			return
		}
		result = map[string]interface{}{
			"success":  true,
		}
		
	} else if strings.HasPrefix(params.Name, "delete_") {
		collectionName := strings.TrimPrefix(params.Name, "delete_")
		filter, _ := params.Arguments["filter"].(map[string]interface{})
		
		mutation := &models.DataMutation{
			Collection: collectionName,
			Operation:  "delete",
			Filter:     filter,
			UserID:     userID,
		}
		
		err := h.collectionService.DeleteDocument(c.Request.Context(), mutation)
		if err != nil {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				Error: &MCPError{
					Code:    -32603,
					Message: "Failed to delete documents",
					Data:    err.Error(),
				},
				ID: req.ID,
			})
			return
		}
		result = map[string]interface{}{
			"success": true,
		}
		
	} else if strings.HasPrefix(params.Name, "execute_view_") {
		viewName := strings.TrimPrefix(params.Name, "execute_view_")
		limit, _ := params.Arguments["limit"].(float64)
		// filter, _ := params.Arguments["filter"].(map[string]interface{}) // TODO: implement additional filter
		
		if limit == 0 {
			limit = 10
		}
		
		opts := collection.QueryOptions{
			Limit:  int(limit),
		}
		
		docs, err := h.collectionService.QueryView(c.Request.Context(), userID, viewName, opts)
		if err != nil {
			c.JSON(http.StatusOK, MCPResponse{
				JSONRPC: "2.0",
				Error: &MCPError{
					Code:    -32603,
					Message: "Failed to query view",
					Data:    err.Error(),
				},
				ID: req.ID,
			})
			return
		}
		result = map[string]interface{}{
			"documents": docs,
			"count":     len(docs),
		}
		
	} else {
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32601,
				Message: "Unknown tool: " + params.Name,
			},
			ID: req.ID,
		})
		return
	}

	// Convert result to JSON string for text content
	resultJSON, err := json.Marshal(result)
	if err != nil {
		c.JSON(http.StatusOK, MCPResponse{
			JSONRPC: "2.0",
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to serialize result",
				Data:    err.Error(),
			},
			ID: req.ID,
		})
		return
	}
	
	c.JSON(http.StatusOK, MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(resultJSON),
				},
			},
		},
		ID: req.ID,
	})
}

