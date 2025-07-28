package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents any purchasable item that grants access or includes physical items
type Product struct {
	ID          string `gorm:"type:varchar(36);primaryKey"`
	EventID     string `gorm:"type:varchar(36);index" json:"event_id"` // event the product is associated with
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `json:"description"`
	PriceInt    int    `gorm:"not null" json:"price_int"`

	MaxOwnableQuantity int `json:"max_ownable_quantity"`

	// Product type flags - a product can be multiple types
	IsEventAccess    bool `gorm:"default:false" json:"is_event_access"`    // Grants event access
	IsActivityAccess bool `gorm:"default:false" json:"is_activity_access"` // Grants activity access
	IsActivityToken  bool `gorm:"default:false" json:"is_activity_token"`  // Can be used as tokens for fee-based activities
	IsPhysicalItem   bool `gorm:"default:false" json:"is_physical_item"`   // Is a physical merchandise item
	IsTicketType     bool `gorm:"default:false" json:"is_ticket_type"`     // Is a ticket type (user can only have one)

	// Visibility and blocking
	IsPublic  bool `gorm:"default:false" json:"is_public"`  // Whether the product is public and can be purchased by anyone
	IsHidden  bool `gorm:"default:false" json:"is_hidden"`  // Whether the product is hidden from search/listings
	IsBlocked bool `gorm:"default:false" json:"is_blocked"` // Whether the product is blocked from purchases

	// Token properties
	TokenQuantity int `gorm:"default:0" json:"token_quantity"` // Number of activity tokens included

	// Bundling
	BundledProducts []Product `gorm:"many2many:product_bundles;constraint:OnDelete:CASCADE" json:"bundled_products"`

	// Stock management (for physical items)
	HasUnlimitedQuantity bool `gorm:"default:false" json:"has_unlimited_quantity"` // If true, ignore Quantity
	Quantity             int  `gorm:"default:0" json:"quantity"`                   // Available quantity

	ExpiresAt time.Time `json:"expires_at"`

	// Relationships - combined into single table with type flag
	AccessTargets []AccessTarget `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"access_targets"`

	Users []User `gorm:"many2many:user_products;foreignKey:ID;joinForeignKey:ProductID;References:ID;joinReferences:UserID" json:"users,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

type ProductRequest struct {
	Name        string `json:"name"`
	EventID     string `json:"event_id"`
	Description string `json:"description"`
	PriceInt    int    `json:"price_int"`

	MaxOwnableQuantity int `json:"max_ownable_quantity"`

	// Product type flags
	IsEventAccess    bool `json:"is_event_access"`
	IsActivityAccess bool `json:"is_activity_access"`
	IsActivityToken  bool `json:"is_activity_token"`
	IsPhysicalItem   bool `json:"is_physical_item"`
	IsTicketType     bool `json:"is_ticket_type"`

	// Visibility and blocking
	IsPublic  bool `json:"is_public"`
	IsHidden  bool `json:"is_hidden"`
	IsBlocked bool `json:"is_blocked"`

	// Token properties
	TokenQuantity int `json:"token_quantity"`

	// Bundling
	// BundledProducts []string `json:"bundled_products"`

	// Stock management
	HasUnlimitedQuantity bool `json:"has_unlimited_quantity"`
	Quantity             int  `json:"quantity"`

	ExpiresAt time.Time `json:"expires_at"`

	// Access targets
	AccessTargets []AccessTargetRequest `json:"access_targets"`
}

type ProductUpdateRequest struct {
	ProductID string         `json:"product_id"`
	Product   ProductRequest `json:"product"`
}

type ProductDeleteRequest struct {
	ProductID string `json:"product_id"`
}

// Purchase represents a transaction record
type Purchase struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    string `gorm:"type:varchar(36);index" json:"user_id"` // User who made the purchase
	ProductID string `gorm:"type:varchar(36);index" json:"product_id"`

	PurchasedAt time.Time `gorm:"autoCreateTime" json:"purchased_at"`
	Quantity    int       `gorm:"default:1" json:"quantity"` // How many of this product

	// For gifting functionality
	IsGift        bool    `gorm:"default:false" json:"is_gift"` // Whether this purchase was a gift
	GiftedToEmail *string `json:"gifted_to_email"`              // User ID of gift recipient

	// For physical items
	IsDelivered bool       `gorm:"default:false" json:"is_delivered"` // If physical item has been delivered
	DeliveredAt *time.Time `json:"delivered_at"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Purchase) TableName() string {
	return "purchases"
}

type PurchaseRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`

	// For gifting functionality
	IsGift        bool    `json:"is_gift"`         // Whether this purchase was a gift
	GiftedToEmail *string `json:"gifted_to_email"` // User email of gift recipient
}

type PurchaseResponse struct {
	Purchase    Purchase    `json:"purchase"`
	UserProduct UserProduct `json:"user_product"`
	UserTokens  []UserToken `json:"user_tokens"`
}

// ProductBundle represents products bundled within other products
type ProductBundle struct {
	ID              string `gorm:"type:varchar(36);primaryKey" json:"id"`
	ParentProductID string `gorm:"type:varchar(36);index" json:"parent_product_id"`
	ChildProductID  string `gorm:"type:varchar(36);index" json:"child_product_id"`
	Quantity        int    `gorm:"default:1" json:"quantity"` // How many of the child product included

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (ProductBundle) TableName() string {
	return "product_bundles"
}

// UserProduct represents products owned by users
type UserProduct struct {
	ID         string `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     string `gorm:"type:varchar(36);index" json:"user_id"`
	ProductID  string `gorm:"type:varchar(36);index" json:"product_id"`
	PurchaseID string `gorm:"type:varchar(36);index" json:"purchase_id"` // Reference to original purchase

	Quantity int `gorm:"default:1" json:"quantity"` // How many of this product the user has

	// For gift tracking
	ReceivedAsGift bool    `gorm:"default:false" json:"received_as_gift"`  // Whether received as a gift
	GiftedFromID   *string `gorm:"type:varchar(36)" json:"gifted_from_id"` // User ID who gifted this product

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (UserProduct) TableName() string {
	return "user_products"
}

type AccessTargetRequest struct {
	ProductID string `json:"product_id"`
	TargetID  string `json:"target_id"`
	IsEvent   bool   `json:"is_event"`
}

// TODO: Simplify access targets not use target ID, but use EventID and ActivityID and IsEvent flag
// AccessTarget represents what a product grants access to (event or activity)
type AccessTarget struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	ProductID string `gorm:"type:varchar(36);index" json:"product_id"`
	TargetID  string `gorm:"type:varchar(36);index" json:"target_id"` // Event ID or Activity ID
	IsEvent   bool   `json:"is_event"`                                // True if target is an event, false if activity

	EventID *string `gorm:"type:varchar(36)" json:"event_id"` // For searching purposes

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AccessTarget) TableName() string {
	return "access_targets"
}

// UserToken represents tokens a user has for accessing fee-based activities
type UserToken struct {
	ID            string `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID        string `gorm:"type:varchar(36);index" json:"user_id"`
	EventID       string `gorm:"type:varchar(36);index" json:"event_id"`
	UserProductID string `gorm:"type:varchar(36);index" json:"user_product_id"` // Which user product this token belongs to
	ProductID     string `gorm:"type:varchar(36);index" json:"product_id"`      // Original product that created this token

	IsUsed    bool       `gorm:"default:false" json:"is_used"`
	UsedAt    *time.Time `json:"used_at"`
	UsedForID *string    `gorm:"type:varchar(36)" json:"used_for_id"` // Activity ID if used

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (UserToken) TableName() string {
	return "user_tokens"
}
