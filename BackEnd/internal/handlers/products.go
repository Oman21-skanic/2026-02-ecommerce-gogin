package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/middleware"
	"github.com/example/ecommerce-api/internal/models"
	"github.com/example/ecommerce-api/internal/store"
)

type ProductsHandler struct {
	store store.Store
}

func NewProductsHandler(st store.Store) *ProductsHandler {
	return &ProductsHandler{store: st}
}

// Admin create product

type createProductReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	Price       json.Number `json:"price"` // accept both price and price_cents
	PriceCents  json.Number `json:"price_cents"`
	SKU         string      `json:"sku"`
	Stock       int         `json:"stock"`
	Thumbnail   string      `json:"thumbnail"`
}

func (h *ProductsHandler) Create(c *gin.Context) {
	var req createProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	// Validasi name dan stock
	if req.Name == "" || req.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan stock harus valid"})
		return
	}

	// Gunakan price jika price_cents tidak ada
	priceCents := int64(0)
	if req.PriceCents != "" {
		parsed, err := parsePriceNumber(req.PriceCents)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price_cents invalid"})
			return
		}
		priceCents = parsed
	} else if req.Price != "" {
		parsed, err := parsePriceNumber(req.Price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price invalid"})
			return
		}
		priceCents = parsed
	}

	if priceCents <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price atau price_cents harus lebih dari 0"})
		return
	}

	// Generate SKU otomatis jika tidak ada
	sku := req.SKU
	if sku == "" {
		sku = generateSKU(req.Name)
	}
	p := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		PriceCents:  priceCents,
		SKU:         sku,
		Stock:       req.Stock,
		Thumbnail:   req.Thumbnail,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	res, err := h.store.CreateProduct(p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Admin update product

type updateProductReq struct {
	Name        *string      `json:"name"`
	Description *string      `json:"description"`
	Category    *string      `json:"category"`
	PriceCents  *json.Number `json:"price_cents"`
	SKU         *string      `json:"sku"`
	Stock       *int         `json:"stock"`
	Thumbnail   *string      `json:"thumbnail"`
}

func (h *ProductsHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req updateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}
	res, err := h.store.UpdateProduct(id, func(p *models.Product) error {
		if req.Name != nil {
			p.Name = *req.Name
		}
		if req.Description != nil {
			p.Description = *req.Description
		}
		if req.Category != nil {
			p.Category = *req.Category
		}
		if req.PriceCents != nil {
			parsed, err := parsePriceNumber(*req.PriceCents)
			if err != nil {
				return err
			}
			p.PriceCents = parsed
		}
		if req.SKU != nil {
			p.SKU = *req.SKU
		}
		if req.Stock != nil {
			p.Stock = *req.Stock
		}
		if req.Thumbnail != nil {
			p.Thumbnail = *req.Thumbnail
		}
		p.UpdatedAt = time.Now()
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ProductsHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteProduct(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductsHandler) Get(c *gin.Context) {
	id := c.Param("id")
	p, err := h.store.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProductsHandler) List(c *gin.Context) {
	q := c.Query("q")
	ps, err := h.store.ListProducts(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ps)
}

func (h *ProductsHandler) AdminList(c *gin.Context) {
	q := c.Query("q")
	sortParam := strings.ToLower(strings.TrimSpace(c.Query("sort")))

	ps, err := h.store.ListProducts(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch sortParam {
	case "bestseller":
		orders, err := h.store.ListOrders()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		sales := make(map[string]int)
		for _, o := range orders {
			status := strings.ToLower(o.Status)
			if status != "paid" && status != "done" && status != "completed" {
				continue
			}
			for _, item := range o.Items {
				sales[item.ProductID] += item.Quantity
			}
		}

		sort.SliceStable(ps, func(i, j int) bool {
			left := sales[ps[i].ID]
			right := sales[ps[j].ID]
			if left == right {
				return ps[i].CreatedAt.After(ps[j].CreatedAt)
			}
			return left > right
		})
	default:
		sort.SliceStable(ps, func(i, j int) bool {
			return ps[i].CreatedAt.After(ps[j].CreatedAt)
		})
	}

	c.JSON(http.StatusOK, ps)
}

// Admin only middleware helper (not a handler)

func AdminOnly() gin.HandlerFunc { return middleware.RequireAdmin() }

// Helper function to generate SKU from product name
func generateSKU(name string) string {
	// Simple SKU generation: uppercase first letters + timestamp
	// Example: "Test Product" -> "TP-1234567890"
	var initials string
	words := strings.Fields(name)
	for _, word := range words {
		if len(word) > 0 {
			initials += strings.ToUpper(string(word[0]))
		}
	}
	if initials == "" {
		initials = "PROD"
	}
	return initials + "-" + fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}

func parsePriceNumber(value json.Number) (int64, error) {
	if value == "" {
		return 0, fmt.Errorf("empty")
	}
	parsed, err := strconv.ParseInt(value.String(), 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
