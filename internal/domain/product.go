package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the system
type Product struct {
	ID        uuid.UUID
	Name      string
	Price     decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProduct creates a new product
func NewProduct(name string, price decimal.Decimal) *Product {
	return &Product{
		ID:        uuid.New(),
		Name:      name,
		Price:     price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewProductFromString creates a new product with string price
func NewProductFromString(name, priceStr string) (*Product, error) {
	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		return nil, NewValidationError("invalid price format")
	}

	return NewProduct(name, price), nil
}

// UpdateDetails updates product name and price
func (p *Product) UpdateDetails(name string, price decimal.Decimal) {
	p.Name = name
	p.Price = price
	p.UpdatedAt = time.Now()
}

// UpdateDetailsFromString updates product with string price
func (p *Product) UpdateDetailsFromString(name, priceStr string) error {
	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		return NewValidationError("invalid price format")
	}

	p.UpdateDetails(name, price)
	return nil
}

// UpdatePrice updates only the price
func (p *Product) UpdatePrice(price decimal.Decimal) {
	p.Price = price
	p.UpdatedAt = time.Now()
}

// UpdatePriceFromString updates price from string
func (p *Product) UpdatePriceFromString(priceStr string) error {
	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		return NewValidationError("invalid price format")
	}

	p.UpdatePrice(price)
	return nil
}

// GetPriceString returns price as string
func (p *Product) GetPriceString() string {
	return p.Price.String()
}

