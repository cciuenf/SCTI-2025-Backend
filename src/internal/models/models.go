package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents any purchasable item that grants access or includes physical items
type Product struct {
	ID          string  `gorm:"type:varchar(36);primaryKey"`
	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Description string  `json:"description"`
	Price       float64 `gorm:"not null" json:"price"`

	// Product type flags - a product can be multiple types
	IsEventAccess    bool `gorm:"default:false" json:"is_event_access"`    // Grants event access
	IsActivityAccess bool `gorm:"default:false" json:"is_activity_access"` // Grants activity access
	IsActivityToken  bool `gorm:"default:false" json:"is_activity_token"`  // Can be used as tokens for fee-based activities
	IsPhysicalItem   bool `gorm:"default:false" json:"is_physical_item"`   // Is a physical merchandise item
	IsTicketType     bool `gorm:"default:false" json:"is_ticket_type"`     // Is a ticket type (user can only have one)

	// Visibility and blocking
	IsHidden  bool `gorm:"default:false" json:"is_hidden"`  // Whether the product is hidden from search/listings
	IsBlocked bool `gorm:"default:false" json:"is_blocked"` // Whether the product is blocked from purchases

	// Token properties
	TokenQuantity int `gorm:"default:0" json:"token_quantity"` // Number of activity tokens included

	// Bundling
	BundledProducts []Product `gorm:"many2many:product_bundles;constraint:OnDelete:CASCADE" json:"bundled_products"`

	// Stock management (for physical items)
	HasUnlimitedQuantity bool `gorm:"default:false" json:"has_unlimited_quantity"` // If true, ignore Quantity
	Quantity             int  `gorm:"default:0" json:"quantity"`                   // Available quantity

	// Relationships - combined into single table with type flag
	AccessTargets []AccessTarget `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"access_targets"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// ActivityRegistration represents a user's registration for an activity
type ActivityRegistration struct {
	ActivityID string `gorm:"type:varchar(36);primaryKey" json:"activity_id"`
	UserID     string `gorm:"type:varchar(36);primaryKey" json:"user_id"`

	// Is this from an event or standalone
	IsStandaloneRegistration bool       `gorm:"default:false" json:"is_standalone_registration"`
	RegisteredAt             time.Time  `gorm:"autoCreateTime" json:"registered_at"`
	AttendedAt               *time.Time `json:"attended_at"` // Time of attendance, null if not attended yet

	// Access method tracking
	AccessMethod string  `gorm:"type:varchar(20)" json:"access_method"` // "event", "product", "token", or "direct"
	TokenID      *string `gorm:"type:varchar(36)" json:"token_id"`      // Which token was used (if applicable)

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Purchase represents a transaction record
type Purchase struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    string `gorm:"type:varchar(36);index" json:"user_id"` // User who made the purchase
	ProductID string `gorm:"type:varchar(36);index" json:"product_id"`

	PurchasedAt time.Time `gorm:"autoCreateTime" json:"purchased_at"`
	Quantity    int       `gorm:"default:1" json:"quantity"` // How many of this product

	// For gifting functionality
	IsGift     bool    `gorm:"default:false" json:"is_gift"`         // Whether this purchase was a gift
	GiftedToID *string `gorm:"type:varchar(36)" json:"gifted_to_id"` // User ID of gift recipient

	// For physical items
	IsDelivered bool       `gorm:"default:false" json:"is_delivered"` // If physical item has been delivered
	DeliveredAt *time.Time `json:"delivered_at"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
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

// AccessTarget represents what a product grants access to (event or activity)
type AccessTarget struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	ProductID string `gorm:"type:varchar(36);index" json:"product_id"`
	TargetID  string `gorm:"type:varchar(36);index" json:"target_id"` // Event ID or Activity ID
	IsEvent   bool   `gorm:"default:true" json:"is_event"`            // True if target is an event, false if activity

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// UserToken represents tokens a user has for accessing fee-based activities
type UserToken struct {
	ID            string `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID        string `gorm:"type:varchar(36);index" json:"user_id"`
	UserProductID string `gorm:"type:varchar(36);index" json:"user_product_id"` // Which product this token belongs to
	ProductID     string `gorm:"type:varchar(36);index" json:"product_id"`      // Original product that created this token

	IsUsed    bool       `gorm:"default:false" json:"is_used"`
	UsedAt    *time.Time `json:"used_at"`
	UsedForID *string    `gorm:"type:varchar(36)" json:"used_for_id"` // Activity ID if used

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// QRCode model (preserved from original code)
type QRCode struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Code      string    `gorm:"type:varchar(100);unique;not null" json:"code"`
	UserID    string    `gorm:"type:varchar(36);index" json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Re-implementing admin system
type AdminType string

const (
	AdminTypeMaster AdminType = "master_admin"
	AdminTypeNormal AdminType = "admin"
)

// AdminStatus represents user admin status for events
type AdminStatus struct {
	gorm.Model
	UserID    string    `gorm:"type:varchar(36)"`
	EventID   string    `gorm:"type:varchar(36)"`
	AdminType AdminType `gorm:"type:varchar(20)"`
}

func (Activity) TableName() string {
	return "activities"
}

func (Product) TableName() string {
	return "products"
}

func (ActivityRegistration) TableName() string {
	return "activity_registrations"
}

func (Purchase) TableName() string {
	return "purchases"
}

func (UserProduct) TableName() string {
	return "user_products"
}

func (ProductBundle) TableName() string {
	return "product_bundles"
}

func (AccessTarget) TableName() string {
	return "access_targets"
}

func (QRCode) TableName() string {
	return "qr_codes"
}

func (UserToken) TableName() string {
	return "user_tokens"
}

func (AdminStatus) TableName() string {
	return "admin_statuses"
}
