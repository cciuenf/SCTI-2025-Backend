package services

import (
	"errors"
	"fmt"
	"net/http"
	"scti/internal/models"
	repos "scti/internal/repositories"

	"github.com/google/uuid"
)

type ProductService struct {
	ProductRepo *repos.ProductRepo
}

func NewProductService(repo *repos.ProductRepo) *ProductService {
	return &ProductService{
		ProductRepo: repo,
	}
}

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

	accessTargets := make([]models.AccessTarget, len(req.AccessTargets))
	for i, target := range req.AccessTargets {
		if target.IsEvent && req.IsEventAccess {
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: target.ProductID,
				TargetID:  target.TargetID,
				IsEvent:   target.IsEvent,
			}
		} else if !target.IsEvent && req.IsActivityAccess {
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: target.ProductID,
				TargetID:  target.TargetID,
				IsEvent:   target.IsEvent,
			}
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

	product := models.Product{
		ID:                   uuid.New().String(),
		EventID:              event.ID,
		Name:                 req.Name,
		Description:          req.Description,
		Price:                req.Price,
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
		AccessTargets:        accessTargets,
	}

	err = s.ProductRepo.CreateProduct(&product)
	if err != nil {
		return nil, errors.New("failed to create product: " + err.Error())
	}

	return &product, nil
}

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

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
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
	product.AccessTargets = req.AccessTargets

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

// TODO: Think very carefully about this but for now, just create a purchase record
// TODO: Implement a try buy so the frontend can show the user if anything will go wrong before they purchase
func (s *ProductService) PurchaseProducts(user models.User, eventSlug string, req models.PurchaseRequest, w http.ResponseWriter) (*models.PurchaseResponse, error) {
	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	product, err := s.ProductRepo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found: " + err.Error())
	}

	if product.EventID != event.ID {
		return nil, errors.New("product does not belong to this event")
	}

	if product.IsBlocked {
		return nil, errors.New("product is blocked from purchases")
	}

	if !product.HasUnlimitedQuantity {
		if product.Quantity < req.Quantity {
			return nil, errors.New("not enough quantity available")
		}
	}

	if req.Quantity > product.MaxOwnableQuantity {
		return nil, errors.New("quantity exceeds max ownable quantity")
	}

	// Query for existing user product
	ownedUserProducts, err := s.ProductRepo.GetUserProductByUserIDAndProductID(user.ID, product.ID)
	if err != nil {
		return nil, errors.New("failed to get user product: " + err.Error())
	}

	// check if the user product is already purchased and how many they have
	// if trying to buy more than allowed, I.E has 3 max allowed 4 but trying to buy 2
	// send a response to the server to let the user know they can't buy more than 1
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

	purchaseID := uuid.New().String()
	purchase := &models.Purchase{
		ID:         purchaseID,
		UserID:     user.ID,
		ProductID:  product.ID,
		Quantity:   req.Quantity,
		IsGift:     req.IsGift,
		GiftedToID: req.GiftedToID,
	}

	err = s.ProductRepo.CreatePurchase(purchase)
	if err != nil {
		return nil, errors.New("failed to create purchase: " + err.Error())
	}

	if !product.HasUnlimitedQuantity {
		product.Quantity -= req.Quantity
		err = s.ProductRepo.UpdateProduct(product)
		if err != nil {
			return nil, errors.New("failed to update product quantity: " + err.Error())
		}
	}

	// Create a user product record
	userProduct := &models.UserProduct{
		ID:         uuid.New().String(),
		PurchaseID: purchaseID,
		ProductID:  product.ID,
		Quantity:   req.Quantity,
	}

	if req.IsGift {
		userProduct.ReceivedAsGift = true
		userProduct.GiftedFromID = &user.ID
		userProduct.UserID = *req.GiftedToID
	} else {
		userProduct.UserID = user.ID
	}

	err = s.ProductRepo.CreateUserProduct(userProduct)
	if err != nil {
		return nil, errors.New("failed to create user product: " + err.Error())
	}

	// create tokens if any
	userTokens := make([]models.UserToken, product.TokenQuantity)
	if product.IsActivityToken {
		for i := 0; i < product.TokenQuantity; i++ {
			token := &models.UserToken{
				ID:            uuid.New().String(),
				UserID:        user.ID,
				UserProductID: userProduct.ID,
				ProductID:     product.ID,
				IsUsed:        false,
				UsedAt:        nil,
				UsedForID:     nil,
			}

			err = s.ProductRepo.CreateUserToken(token)
			if err != nil {
				return nil, errors.New("failed to create user token: " + err.Error())
			}
			userTokens[i] = *token
		}
	}

	return &models.PurchaseResponse{
		Purchase:    *purchase,
		UserProduct: *userProduct,
		UserTokens:  userTokens,
	}, nil
}

func (s *ProductService) TryPurchaseProducts(user models.User, eventSlug string, req models.PurchaseRequest) (*models.PurchaseResponse, error) {
	event, err := s.ProductRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	product, err := s.ProductRepo.GetProductByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found: " + err.Error())
	}

	if product.EventID != event.ID {
		return nil, errors.New("product does not belong to this event")
	}

	if product.IsBlocked {
		return nil, errors.New("product is blocked from purchases")
	}

	if !product.HasUnlimitedQuantity {
		if product.Quantity < req.Quantity {
			return nil, errors.New("not enough quantity available")
		}
	}

	if req.Quantity > product.MaxOwnableQuantity {
		return nil, errors.New("quantity exceeds max ownable quantity")
	}

	// Query for existing user product
	ownedUserProducts, err := s.ProductRepo.GetUserProductByUserIDAndProductID(user.ID, product.ID)
	if err != nil {
		return nil, errors.New("failed to get user product: " + err.Error())
	}

	// check if the user product is already purchased and how many they have
	// if trying to buy more than allowed, I.E has 3 max allowed 4 but trying to buy 2
	// send a response to the server to let the user know they can't buy more than 1
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

	purchaseID := uuid.New().String()
	purchase := &models.Purchase{
		ID:         purchaseID,
		UserID:     user.ID,
		ProductID:  product.ID,
		Quantity:   req.Quantity,
		IsGift:     req.IsGift,
		GiftedToID: req.GiftedToID,
	}

	// Simulate a purchase record with a special db function that doesn't actually create a record
	err = s.ProductRepo.SimulatePurchase(purchase)
	if err != nil {
		return nil, errors.New("failed to simulate purchase: " + err.Error())
	}

	// simulate updating product quantity
	product.Quantity -= req.Quantity
	err = s.ProductRepo.SimulateUpdateProduct(product)
	if err != nil {
		return nil, errors.New("failed to simulate update product quantity: " + err.Error())
	}

	// Create a user product record
	userProduct := &models.UserProduct{
		ID:         uuid.New().String(),
		PurchaseID: purchaseID,
		ProductID:  product.ID,
		Quantity:   req.Quantity,
	}

	if req.IsGift {
		userProduct.ReceivedAsGift = true
		userProduct.GiftedFromID = &user.ID
		userProduct.UserID = *req.GiftedToID
	} else {
		userProduct.UserID = user.ID
	}

	// simulate creating a user product record
	err = s.ProductRepo.SimulateUserProduct(userProduct)
	if err != nil {
		return nil, errors.New("failed to simulate user product: " + err.Error())
	}

	// create tokens if any
	userTokens := make([]models.UserToken, product.TokenQuantity)
	if product.IsActivityToken {
		for i := 0; i < product.TokenQuantity; i++ {
			token := &models.UserToken{
				ID:            uuid.New().String(),
				UserID:        user.ID,
				UserProductID: userProduct.ID,
				ProductID:     product.ID,
				IsUsed:        false,
				UsedAt:        nil,
				UsedForID:     nil,
			}

			// simulate creating a user token record
			err = s.ProductRepo.SimulateUserToken(token)
			if err != nil {
				return nil, errors.New("failed to simulate user token: " + err.Error())
			}
			userTokens[i] = *token
		}
	}

	return &models.PurchaseResponse{
		Purchase:    *purchase,
		UserProduct: *userProduct,
		UserTokens:  userTokens,
	}, nil
}
