package repos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"scti/config"
	"scti/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mercadopago/sdk-go/pkg/order"
	"github.com/mercadopago/sdk-go/pkg/refund"
	"gorm.io/gorm"
)

type ProductRepo struct {
	DB *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{DB: db}
}

func (r *ProductRepo) CreateProduct(product *models.Product) error {
	return r.DB.Create(product).Error
}

func (r *ProductRepo) GetProductByID(id string) (*models.Product, error) {
	var product models.Product
	if err := r.DB.Preload("AccessTargets").Where("id = ?", id).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepo) GetProductsByIDs(ids []string) ([]models.Product, error) {
	var products []models.Product
	if err := r.DB.Preload("AccessTargets").Where("id IN ?", ids).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepo) GetProductsByEventID(eventID string) ([]models.Product, error) {
	var products []models.Product
	if err := r.DB.Preload("AccessTargets").Where("event_id = ?", eventID).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepo) UpdateProduct(product *models.Product) error {
	return r.DB.Save(product).Error
}

func (r *ProductRepo) RemoveAccessTargets(product *models.Product) error {
	return r.DB.Where("product_id = ?", product.ID).Delete(&models.AccessTarget{}).Error
}

func (r *ProductRepo) DeleteProduct(id string) error {
	return r.DB.Where("id = ?", id).Delete(&models.Product{}).Error
}

func (r *ProductRepo) CreatePurchase(purchase *models.Purchase) error {
	return r.DB.Create(purchase).Error
}

func (r *ProductRepo) GetUserPurchases(userID string) ([]models.Purchase, error) {
	var purchases []models.Purchase
	if err := r.DB.Where("user_id = ?", userID).Find(&purchases).Error; err != nil {
		return nil, err
	}
	return purchases, nil
}

func (r *ProductRepo) GetUserByID(userID string) (models.User, error) {
	var user models.User
	if err := r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *ProductRepo) GetUserByEmail(userEmail string) (models.User, error) {
	lemail := strings.TrimSpace(strings.ToLower(userEmail))
	var user models.User
	if err := r.DB.Where("email = ?", lemail).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *ProductRepo) GetActivityByID(activityID string) (*models.Activity, error) {
	var activity models.Activity
	if err := r.DB.Where("id = ?", activityID).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ProductRepo) GetEventByID(eventID string) (*models.Event, error) {
	var event models.Event
	if err := r.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *ProductRepo) GetEventBySlug(slug string) (*models.Event, error) {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *ProductRepo) GetAdminStatusForEvent(userID string, eventID string) (models.AdminStatus, error) {
	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_id = ?", userID, eventID).First(&adminStatus).Error; err != nil {
		return models.AdminStatus{}, err
	}
	return adminStatus, nil
}

func (r *ProductRepo) IsUserRegisteredToEvent(userID string, eventID string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.EventRegistration{}).
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ProductRepo) CreateUserProduct(userProduct *models.UserProduct) error {
	return r.DB.Create(userProduct).Error
}

func (r *ProductRepo) CreateUserToken(userToken *models.UserToken) error {
	return r.DB.Create(userToken).Error
}

func (r *ProductRepo) GetUserProductByUserIDAndProductID(userID string, productID string) ([]models.UserProduct, error) {
	var userProducts []models.UserProduct
	if err := r.DB.Where("user_id = ? AND product_id = ?", userID, productID).Find(&userProducts).Error; err != nil {
		return nil, err
	}
	return userProducts, nil
}

func (r *ProductRepo) GetUserProductsRelation(userID string) ([]models.UserProduct, error) {
	var userProducts []models.UserProduct
	if err := r.DB.Where("user_id = ?", userID).Find(&userProducts).Error; err != nil {
		return nil, err
	}
	return userProducts, nil
}

func (r *ProductRepo) GetUserTokens(userID string) ([]models.UserToken, error) {
	var userTokens []models.UserToken
	if err := r.DB.Where("user_id = ?", userID).Find(&userTokens).Error; err != nil {
		return nil, err
	}
	return userTokens, nil
}

// TODO: Think very carefully about this but for now, just do the basic proccess as I think it should be done
func (r *ProductRepo) PurchaseProduct(user models.User, event *models.Event, product *models.Product, req models.PurchaseRequest, w http.ResponseWriter) (*models.PurchaseResponse, error) {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to begin transaction: " + tx.Error.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Query for existing user product
	purchaseID := uuid.New().String()
	purchase := &models.Purchase{
		ID:            purchaseID,
		UserID:        user.ID,
		ProductID:     product.ID,
		Quantity:      req.Quantity,
		IsGift:        req.IsGift,
		GiftedToEmail: req.GiftedToEmail,
	}

	err := tx.Create(purchase).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create purchase: " + err.Error())
	}

	if !product.HasUnlimitedQuantity {
		product.Quantity -= req.Quantity
		err = tx.Save(product).Error
		if err != nil {
			tx.Rollback()
			return nil, errors.New("failed to update product quantity: " + err.Error())
		}
	}

	userProduct := &models.UserProduct{
		ID:         uuid.New().String(),
		PurchaseID: purchaseID,
		ProductID:  product.ID,
		Quantity:   req.Quantity,
	}

	if req.IsGift {
		if req.GiftedToEmail == nil {
			tx.Rollback()
			return nil, errors.New("can't gift to nil email")
		}
		giftedUser, err := r.GetUserByEmail(*req.GiftedToEmail)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("failed to retrieve user for gifting")
		}
		userProduct.ReceivedAsGift = true
		userProduct.GiftedFromID = &user.ID
		userProduct.UserID = giftedUser.ID
	} else {
		userProduct.ReceivedAsGift = false
		userProduct.GiftedFromID = nil
		userProduct.UserID = user.ID
	}

	err = tx.Create(userProduct).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create user product: " + err.Error())
	}

	userTokens := make([]models.UserToken, product.TokenQuantity)
	if product.IsActivityToken {
		for i := 0; i < product.TokenQuantity; i++ {
			token := &models.UserToken{
				ID:            uuid.New().String(),
				EventID:       event.ID,
				UserID:        userProduct.UserID,
				UserProductID: userProduct.ID,
				ProductID:     product.ID,
				IsUsed:        false,
				UsedAt:        nil,
				UsedForID:     nil,
			}

			err = tx.Create(token).Error
			if err != nil {
				tx.Rollback()
				return nil, errors.New("failed to create user token: " + err.Error())
			}
			userTokens[i] = *token
		}
	}

	for _, access := range product.AccessTargets {
		registration := &models.ActivityRegistration{
			ActivityID:   access.TargetID,
			ProductID:    &product.ID,
			AccessMethod: string(models.AccessMethodProduct),
			UserID:       userProduct.UserID,
		}
		var count int64
		err = tx.Model(&models.ActivityRegistration{}).
			Where("activity_id = ? AND user_id = ?", registration.ActivityID, registration.UserID).
			Count(&count).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return nil, errors.New("failed to get activity registration: " + err.Error())
		}

		if count > 0 {
			continue
		}

		err = tx.Create(registration).Error
		if err != nil {
			tx.Rollback()
			return nil, errors.New("failed to create activity registration: " + err.Error())
		}
	}

	// ----------------------------------------------------- //
	// ----------------COMEÇO DO PAGAMENTO ----------------- //
	// ----------------------------------------------------- //

	mercadoPagoConfig := config.GetMercadoPagoConfig()

	client := order.NewClient(mercadoPagoConfig)
	request := order.Request{
		Type:              "online",
		TotalAmount:       fmt.Sprintf("%.2f", (float64(product.PriceInt)*float64(req.Quantity))/100),
		ExternalReference: fmt.Sprintf("%s_%s", event.Slug, user.ID),
		Transactions: &order.TransactionRequest{
			Payments: []order.PaymentRequest{
				{
					Amount: fmt.Sprintf("%.2f", (float64(product.PriceInt)*float64(req.Quantity))/100),
					PaymentMethod: &order.PaymentMethodRequest{
						ID:           req.PaymentMethodID,
						Token:        req.PaymentMethodToken,
						Type:         req.PaymentMethodType,
						Installments: req.PaymentMethodInstallments,
					},
				},
			},
		},
		Payer: &order.PayerRequest{
			Email: user.Email,
		},
	}

	resource, err := client.Create(context.Background(), request)
	if err != nil {
		tx.Rollback()
		log.Printf("Mercado Pago API error: %v", err)
		return nil, errors.New("failed to create mercado pago order: " + err.Error())
	}

	// -------------------------------------------------- //
	// ---------------- FIM DO PAGAMENTO ---------------- //
	// -------------------------------------------------- //

	// CRITICAL SECTION: Commit with refund fallback
	if err := tx.Commit().Error; err != nil {
		// Payment succeeded but database commit failed - MUST refund
		log.Printf("CRITICAL: Database commit failed after successful payment %s. Attempting refund...", resource.ID)

		refundErr := r.attemptRefund(resource)
		if refundErr != nil {
			// This is the worst case scenario - log extensively and alert admins
			log.Printf("CRITICAL FAILURE: Could not refund payment %s after failed commit. Manual intervention required. Original error: %v, Refund error: %v",
				resource.ID, err, refundErr)

			// Store for manual processing
			r.storeFailedTransaction(resource, user, purchase, err.Error(), refundErr.Error())
		}

		return nil, errors.New("failed to commit transaction: " + err.Error())
	}

	return &models.PurchaseResponse{
		Purchase:         *purchase,
		UserProduct:      *userProduct,
		UserTokens:       userTokens,
		PurchaseResource: resource,
	}, nil
}

// Helper to attempt refund
func (r *ProductRepo) attemptRefund(resource *order.Response) error {
	if resource == nil || resource.ID == "" {
		return errors.New("invalid payment resource")
	}

	paymentID, err := strconv.Atoi(resource.ID)
	if err != nil {
		return errors.New("invalid payment ID format: " + err.Error())
	}

	amount, err := strconv.ParseFloat(resource.TotalAmount, 64)
	if err != nil {
		return errors.New("invalid amount format: " + err.Error())
	}

	mercadoPagoConfig := config.GetMercadoPagoConfig()
	refundClient := refund.NewClient(mercadoPagoConfig)

	_, err = refundClient.Create(context.Background(), paymentID)

	if err != nil {
		log.Printf("Failed to refund payment %d: %v", paymentID, err)
		return err
	}

	log.Printf("Successfully refunded payment %d for amount %.2f", paymentID, amount)
	return nil
}

// Store failed transactions for manual processing, still need to implement on DB
func (r *ProductRepo) storeFailedTransaction(resource *order.Response, user models.User, purchase *models.Purchase, dbError, refundError string) {
	// Create a record in a separate table/system for manual intervention
	failedTx := map[string]interface{}{
		"payment_id":    resource.ID,
		"user_id":       user.ID,
		"amount":        resource.TotalAmount,
		"purchase_data": purchase,
		"db_error":      dbError,
		"refund_error":  refundError,
		"created_at":    time.Now(),
		"status":        "manual_intervention_required",
	}

	// Log to a monitoring system, database table, or external service
	log.Printf("FAILED_TRANSACTION: %+v", failedTx)

	// Send alerts to administrators
}
