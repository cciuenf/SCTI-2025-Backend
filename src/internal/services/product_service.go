package services

import (
	"errors"
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
				IsEvent:   true,
			}
		} else if !target.IsEvent && req.IsActivityAccess {
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: productID,
				TargetID:  target.TargetID,
				EventID:   &event.ID,
				IsEvent:   false,
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
				IsEvent:   false,
			}
		} else if !target.IsEvent && req.IsActivityAccess {
			accessTargets[i] = models.AccessTarget{
				ID:        uuid.New().String(),
				ProductID: productID,
				TargetID:  target.TargetID,
				EventID:   &event.ID,
				IsEvent:   false,
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

// TODO: Integrate bundled products
func (s *ProductService) PurchaseProducts(user models.User, eventSlug string, req models.PurchaseRequest, w http.ResponseWriter) (*models.PurchaseResponse, error) {
	if req.IsGift && *req.GiftedToEmail == user.Email {
		return nil, errors.New("invalid operation: cannot gift to yourself")
	}
	return s.ProductRepo.PurchaseProduct(user, eventSlug, req)
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
