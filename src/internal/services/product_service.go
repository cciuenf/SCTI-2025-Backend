package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"scti/config"
	"scti/internal/models"
	repos "scti/internal/repositories"
	"time"

	"github.com/google/uuid"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

type ProductService struct {
	ProductRepo *repos.ProductRepo
}

func NewProductService(repo *repos.ProductRepo) *ProductService {
	return &ProductService{
		ProductRepo: repo,
	}
}

// TODO: Integrate bundled products
// TODO: Event access target should give access to all activities in the event
func (s *ProductService) CreateEventProduct(user models.User, eventSlug string, req models.ProductRequest) (*models.Product, error) {
	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	// Check if user has permission to create products
	// master admins can create products for the event
	if !user.IsSuperUser && event.CreatedBy != user.ID {
		// Get admin status of the user
		adminStatus, err := s.ProductRepo.GetAdminStatusForEvent(user.ID, event.ID)
		if err != nil {
			return nil, errors.New("failed to get admin status: " + err.Error())
		}

		if adminStatus.AdminType != models.AdminTypeMaster {
			return nil, errors.New("unauthorized to create products for this event")
		}
	}

	if req.IsActivityToken && req.TokenQuantity <= 0 {
		return nil, errors.New("token quantity must be greater than 0")
	}

	if req.IsEventAccess || req.IsActivityAccess {
		req.IsTicketType = true
	} else {
		req.IsTicketType = false
	}

	productID := uuid.New().String()

	accessTargets := make([]models.AccessTarget, len(req.AccessTargets))
	for i, target := range req.AccessTargets {
		if target.IsEvent && req.IsEventAccess {
			if target.TargetID != event.ID {
				return nil, errors.New("invalid access target: targeting wrong event")
			}
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: productID,
				TargetID:  target.TargetID,
				EventID:   &event.ID,
				IsEvent:   target.IsEvent,
			}
		} else if !target.IsEvent && req.IsActivityAccess {
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: productID,
				TargetID:  target.TargetID,
				EventID:   &event.ID,
				IsEvent:   target.IsEvent,
			}
		}
	}

	for _, target := range accessTargets {
		if req.IsActivityAccess {
			activity, err := s.ProductRepo.GetActivityByID(target.TargetID)
			if err != nil {
				return nil, errors.New("invalid access target, couldn't find activity")
			}
			if activity.EventID != nil && target.EventID != nil && *activity.EventID != *target.EventID {
				return nil, errors.New("invalid access target: targeting inexistent activity")
			}
		}
	}

	if req.ExpiresAt.IsZero() {
		req.ExpiresAt = event.EndDate
	}

	expiresAfterEvent := req.ExpiresAt.After(event.EndDate)
	if expiresAfterEvent {
		return nil, errors.New("product needs to expire before the event end date")
	}

	product := models.Product{
		ID:                   productID,
		EventID:              event.ID,
		Name:                 req.Name,
		Description:          req.Description,
		PriceInt:             req.PriceInt,
		MaxOwnableQuantity:   req.MaxOwnableQuantity,
		IsEventAccess:        req.IsEventAccess,
		IsActivityAccess:     req.IsActivityAccess,
		IsActivityToken:      req.IsActivityToken,
		IsPhysicalItem:       req.IsPhysicalItem,
		IsTicketType:         req.IsTicketType,
		IsPublic:             req.IsPublic,
		IsHidden:             req.IsHidden,
		IsBlocked:            req.IsBlocked,
		TokenQuantity:        req.TokenQuantity,
		HasUnlimitedQuantity: req.HasUnlimitedQuantity,
		Quantity:             req.Quantity,
		ExpiresAt:            req.ExpiresAt,
		AccessTargets:        accessTargets,
	}

	err = s.ProductRepo.CreateProduct(&product)
	if err != nil {
		return nil, errors.New("failed to create product: " + err.Error())
	}

	return &product, nil
}

// TODO: Integrate bundled products
func (s *ProductService) UpdateEventProduct(user models.User, eventSlug string, productID string, req models.ProductRequest) (*models.Product, error) {
	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	product, err := s.ProductRepo.GetProductByID(productID)
	if err != nil {
		return nil, errors.New("product not found: " + err.Error())
	}

	if product.EventID != event.ID {
		return nil, errors.New("product does not belong to this event")
	}

	// Check if user has permission to update products
	if !user.IsSuperUser && event.CreatedBy != user.ID {
		// Get admin status of the user
		adminStatus, err := s.ProductRepo.GetAdminStatusForEvent(user.ID, event.ID)
		if err != nil {
			return nil, errors.New("failed to get admin status: " + err.Error())
		}

		if adminStatus.AdminType != models.AdminTypeMaster {
			return nil, errors.New("unauthorized to update products for this event")
		}
	}

	if req.ExpiresAt.IsZero() {
		req.ExpiresAt = event.EndDate
	}

	if req.ExpiresAt.After(event.EndDate) {
		return nil, errors.New("product can't expire after event end date")
	}

	product.Name = req.Name
	product.Description = req.Description
	product.PriceInt = req.PriceInt
	product.MaxOwnableQuantity = req.MaxOwnableQuantity
	product.IsEventAccess = req.IsEventAccess
	product.IsActivityAccess = req.IsActivityAccess
	product.IsActivityToken = req.IsActivityToken
	product.IsPhysicalItem = req.IsPhysicalItem
	product.IsTicketType = req.IsTicketType
	product.IsPublic = req.IsPublic
	product.IsHidden = req.IsHidden
	product.IsBlocked = req.IsBlocked
	product.TokenQuantity = req.TokenQuantity
	product.HasUnlimitedQuantity = req.HasUnlimitedQuantity
	product.Quantity = req.Quantity
	product.ExpiresAt = req.ExpiresAt

	accessTargets := make([]models.AccessTarget, len(req.AccessTargets))
	for i, target := range req.AccessTargets {
		if target.IsEvent && req.IsEventAccess {
			if target.TargetID != event.ID {
				return nil, errors.New("invalid access target: targeting wrong event")
			}
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: productID,
				TargetID:  target.TargetID,
				EventID:   &event.ID,
				IsEvent:   target.IsEvent,
			}
		} else if !target.IsEvent && req.IsActivityAccess {
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: productID,
				TargetID:  target.TargetID,
				EventID:   &event.ID,
				IsEvent:   target.IsEvent,
			}
		}
	}

	for _, target := range accessTargets {
		if req.IsActivityAccess {
			activity, err := s.ProductRepo.GetActivityByID(target.TargetID)
			if err != nil {
				return nil, errors.New("invalid access target, couldn't find activity")
			}
			if activity.EventID != nil && target.EventID != nil && *activity.EventID != *target.EventID {
				return nil, errors.New("invalid access target: targeting inexistent activity")
			}
		}
	}

	err = s.ProductRepo.RemoveAccessTargets(product)
	if err != nil {
		return nil, errors.New("failed to clear access targets for updating: " + err.Error())
	}

	product.AccessTargets = accessTargets

	err = s.ProductRepo.UpdateProduct(product)
	if err != nil {
		return nil, errors.New("failed to update product: " + err.Error())
	}

	return product, nil
}

// TODO: Can't delete a product if it has been purchased
func (s *ProductService) DeleteEventProduct(user models.User, eventSlug string, productID string) error {
	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	product, err := s.ProductRepo.GetProductByID(productID)
	if err != nil {
		return errors.New("product not found: " + err.Error())
	}

	if product.EventID != event.ID {
		return errors.New("product does not belong to this event")
	}

	// Check if user has permission to delete products
	if !user.IsSuperUser && event.CreatedBy != user.ID {
		// Get admin status of the user
		adminStatus, err := s.ProductRepo.GetAdminStatusForEvent(user.ID, event.ID)
		if err != nil {
			return errors.New("failed to get admin status: " + err.Error())
		}

		if adminStatus.AdminType != models.AdminTypeMaster {
			return errors.New("unauthorized to delete products for this event")
		}
	}

	// TODO: Check if product has been purchased before deletion

	err = s.ProductRepo.DeleteProduct(productID)
	if err != nil {
		return errors.New("failed to delete product: " + err.Error())
	}

	return nil
}

func (s *ProductService) GetAllProductsFromEvent(eventSlug string) ([]models.Product, error) {
	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	products, err := s.ProductRepo.GetProductsByEventID(event.ID)
	if err != nil {
		return nil, errors.New("failed to get products: " + err.Error())
	}

	return products, nil
}

func (s *ProductService) GetUserProductsRelation(user models.User) ([]models.UserProduct, error) {
	products, err := s.ProductRepo.GetUserProductsRelation(user.ID)
	if err != nil {
		return nil, errors.New("failed to get products: " + err.Error())
	}

	return products, nil
}

func (s *ProductService) GetUserProducts(user models.User) ([]models.Product, error) {
	userProducts, err := s.ProductRepo.GetUserProductsRelation(user.ID)
	if err != nil {
		return nil, errors.New("failed to get products: " + err.Error())
	}

	productIDs := make([]string, len(userProducts))
	for i, product := range userProducts {
		productIDs[i] = product.ProductID
	}

	products, err := s.ProductRepo.GetProductsByIDs(productIDs)
	if err != nil {
		return nil, errors.New("failed to get products: " + err.Error())
	}

	return products, nil
}

func (s *ProductService) GetUserTokens(user models.User) ([]models.UserToken, error) {
	return s.ProductRepo.GetUserTokens(user.ID)
}

func (s *ProductService) GetUserPurchases(user models.User) ([]models.Purchase, error) {
	return s.ProductRepo.GetUserPurchases(user.ID)
}

// TODO: Integrate bundled products
func (s *ProductService) PurchaseProducts(user models.User, eventSlug string, req models.PurchaseRequest, w http.ResponseWriter) (*models.PurchaseResponse, error) {
	if req.IsGift {
		if req.GiftedToEmail == nil {
			return nil, errors.New("gifted_to_email is required when gifting")
		}
		if *req.GiftedToEmail == user.Email {
			return nil, errors.New("invalid operation: cannot gift to yourself")
		}
	}

	if req.PaymentMethodID == "" {
		return nil, errors.New("payment method ID is required")
	}
	if req.PaymentMethodID == "pix" {
		return nil, errors.New("use the create-pix-purchase endpoint")
	}
	if req.PaymentMethodToken == "" {
		return nil, errors.New("payment method token is required")
	}
	if req.PaymentMethodInstallments < 1 {
		return nil, errors.New("installments must be at least 1")
	}

	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	isUserRegistered, err := s.ProductRepo.IsUserRegisteredToEvent(user.ID, event.ID)
	if err != nil {
		return nil, errors.New("error checking user registration: " + err.Error())
	}

	if !isUserRegistered {
		return nil, errors.New("user is not registered to this event")
	}

	product, err := s.ProductRepo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found: " + err.Error())
	}

	if product.IsBlocked {
		return nil, errors.New("product is blocked from purchases")
	}

	if product.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("product has expired")
	}

	if product.EventID != event.ID {
		return nil, errors.New("product does not belong to this event")
	}

	if req.Quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}

	if !product.HasUnlimitedQuantity {
		if product.Quantity < req.Quantity {
			return nil, fmt.Errorf("not enough quantity available, want %v have %v", req.Quantity, product.Quantity)
		}
	}

	if req.Quantity > product.MaxOwnableQuantity {
		return nil, fmt.Errorf("requested quantity exceeds max ownable quantity by: %d", req.Quantity-product.MaxOwnableQuantity)
	}

	ownedUserProducts, err := s.ProductRepo.GetUserProductByUserIDAndProductID(user.ID, product.ID)
	if err != nil {
		return nil, errors.New("failed to get user product: " + err.Error())
	}

	var ownedQuantity int
	if len(ownedUserProducts) > 0 {
		for _, userProduct := range ownedUserProducts {
			ownedQuantity += userProduct.Quantity
		}
	}

	if ownedQuantity+req.Quantity > product.MaxOwnableQuantity {
		text := fmt.Sprintf("user with %d of this product is trying to buy %d, max ownable quantity is %d, this exceeds it by %d", ownedQuantity, req.Quantity, product.MaxOwnableQuantity, ownedQuantity+req.Quantity-product.MaxOwnableQuantity)
		return nil, errors.New(text)
	}

	return s.ProductRepo.PurchaseProduct(user, event, product, req, w)
}

func (s *ProductService) PreferenceRequest(user models.User, eventSlug string, req models.PixPurchaseRequest) (*preference.Response, error) {
	if req.IsGift {
		if req.GiftedToEmail == nil {
			return nil, errors.New("gifted_to_email is required when gifting")
		}
		if *req.GiftedToEmail == user.Email {
			return nil, errors.New("invalid operation: cannot gift to yourself")
		}
	}

	if req.PaymentMethodID == "" {
		return nil, errors.New("payment method ID is required")
	}
	if req.PaymentMethodID != "pix" {
		if req.PaymentMethodToken == "" {
			return nil, errors.New("payment method token is required")
		}
		if req.PaymentMethodInstallments < 1 {
			return nil, errors.New("installments must be at least 1")
		}
	}

	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	isUserRegistered, err := s.ProductRepo.IsUserRegisteredToEvent(user.ID, event.ID)
	if err != nil {
		return nil, errors.New("error checking user registration: " + err.Error())
	}

	if !isUserRegistered {
		return nil, errors.New("user is not registered to this event")
	}

	product, err := s.ProductRepo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found: " + err.Error())
	}

	if product.IsBlocked {
		return nil, errors.New("product is blocked from purchases")
	}

	if product.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("product has expired")
	}

	if product.EventID != event.ID {
		return nil, errors.New("product does not belong to this event")
	}

	if req.Quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}

	if !product.HasUnlimitedQuantity {
		if product.Quantity < req.Quantity {
			return nil, fmt.Errorf("not enough quantity available, want %v have %v", req.Quantity, product.Quantity)
		}
	}

	if req.Quantity > product.MaxOwnableQuantity {
		return nil, fmt.Errorf("requested quantity exceeds max ownable quantity by: %d", req.Quantity-product.MaxOwnableQuantity)
	}

	ownedUserProducts, err := s.ProductRepo.GetUserProductByUserIDAndProductID(user.ID, product.ID)
	if err != nil {
		return nil, errors.New("failed to get user product: " + err.Error())
	}

	var ownedQuantity int
	if len(ownedUserProducts) > 0 {
		for _, userProduct := range ownedUserProducts {
			ownedQuantity += userProduct.Quantity
		}
	}

	if ownedQuantity+req.Quantity > product.MaxOwnableQuantity {
		text := fmt.Sprintf("user with %d of this product is trying to buy %d, max ownable quantity is %d, this exceeds it by %d", ownedQuantity, req.Quantity, product.MaxOwnableQuantity, ownedQuantity+req.Quantity-product.MaxOwnableQuantity)
		return nil, errors.New(text)
	}

	// ---------------------------------------------------------- //
	// ----------------COMEÇO DA PREFERENCIA -------------------- //
	// ---------------------------------------------------------- //
	mercadoPagoConfig := config.GetMercadoPagoConfig()

	request := preference.Request{
		BackURLs: &preference.BackURLsRequest{
			Success: "http://localhost:3000/",
			Pending: "",
			Failure: "",
		},
		Items: []preference.ItemRequest{
			{
				ID:          product.ID,
				Title:       product.Name,
				UnitPrice:   float64(product.PriceInt) / 100,
				Quantity:    req.Quantity,
				Description: product.Description,
				CurrencyID:  "BRL",
				CategoryID:  "event-product",
			},
		},
		NotificationURL: "https://webhook.site/fdfdb700-b508-45f6-bd90-ebab4e9dc81b",
	}

	if req.PaymentMethodID != "pix" {
		request.PaymentMethods = &preference.PaymentMethodsRequest{
			DefaultPaymentMethodID: req.PaymentMethodID,
			Installments:           req.PaymentMethodInstallments,
		}
	}

	client := preference.NewClient(mercadoPagoConfig)
	resource, err := client.Create(context.Background(), request)
	if err != nil {
		log.Printf("Mercado Pago Preference error: %v", err)
		return nil, errors.New("failed to create mercado pago preference: " + err.Error())
	}

	return resource, nil
	// ---------------------------------------------------- //
	// ---------------- FIM DA PREFERENCIA ---------------- //
	// ---------------------------------------------------- //
}

func (s *ProductService) ForcedPayment(user models.User, eventSlug string, req models.PixPurchaseRequest) (*payment.Response, error) {
	if req.IsGift {
		if req.GiftedToEmail == nil {
			return nil, errors.New("gifted_to_email is required when gifting")
		}
		if *req.GiftedToEmail == user.Email {
			return nil, errors.New("invalid operation: cannot gift to yourself")
		}
	}

	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	isUserRegistered, err := s.ProductRepo.IsUserRegisteredToEvent(user.ID, event.ID)
	if err != nil {
		return nil, errors.New("error checking user registration: " + err.Error())
	}

	if !isUserRegistered {
		return nil, errors.New("user is not registered to this event")
	}

	product, err := s.ProductRepo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found: " + err.Error())
	}

	if product.IsBlocked {
		return nil, errors.New("product is blocked from purchases")
	}

	if product.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("product has expired")
	}

	if product.EventID != event.ID {
		return nil, errors.New("product does not belong to this event")
	}

	if req.Quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}

	if !product.HasUnlimitedQuantity {
		if product.Quantity < req.Quantity {
			return nil, fmt.Errorf("not enough quantity available, want %v have %v", req.Quantity, product.Quantity)
		}
	}

	if req.Quantity > product.MaxOwnableQuantity {
		return nil, fmt.Errorf("requested quantity exceeds max ownable quantity by: %d", req.Quantity-product.MaxOwnableQuantity)
	}

	ownedUserProducts, err := s.ProductRepo.GetUserProductByUserIDAndProductID(user.ID, product.ID)
	if err != nil {
		return nil, errors.New("failed to get user product: " + err.Error())
	}

	var ownedQuantity int
	if len(ownedUserProducts) > 0 {
		for _, userProduct := range ownedUserProducts {
			ownedQuantity += userProduct.Quantity
		}
	}

	if ownedQuantity+req.Quantity > product.MaxOwnableQuantity {
		text := fmt.Sprintf("user with %d of this product is trying to buy %d, max ownable quantity is %d, this exceeds it by %d", ownedQuantity, req.Quantity, product.MaxOwnableQuantity, ownedQuantity+req.Quantity-product.MaxOwnableQuantity)
		return nil, errors.New(text)
	}

	// ----------------------------------------------------- //
	// ----------------COMEÇO DO PAGAMENTO ----------------- //
	// ----------------------------------------------------- //

	mercadoPagoConfig := config.GetMercadoPagoConfig()
	paymentClient := payment.NewClient(mercadoPagoConfig)
	request := payment.Request{
		TransactionAmount: (float64(product.PriceInt) / 100) * float64(req.Quantity),
		PaymentMethodID:   "pix",
		Payer: &payment.PayerRequest{
			Email: user.Email,
		},
	}
	resource, err := paymentClient.Create(context.Background(), request)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to create mercado pago payment")
	}

	// -------------------------------------------------- //
	// ---------------- FIM DO PAGAMENTO ---------------- //
	// -------------------------------------------------- //

	return resource, nil
}
