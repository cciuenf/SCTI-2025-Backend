package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"scti/config"
	"scti/internal/models"
	"scti/internal/services"
	"strconv"
	"strings"
)

type ProductHandler struct {
	ProductService *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{ProductService: service}
}

// CreateEventProduct godoc
// @Summary      Create a product for an event
// @Description  Creates a new product for the specified event
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ProductRequest true "Product creation info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Product}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Failure      403  {object}  ProductStandardErrorResponse
// @Router       /events/{slug}/product [post]
func (h *ProductHandler) CreateEventProduct(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	var reqBody models.ProductRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "product")
		return
	}

	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	product, err := h.ProductService.CreateEventProduct(user, slug, reqBody)
	if err != nil {
		HandleErrMsg("error creating product", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, product, "", http.StatusOK)
}

// UpdateEventProduct godoc
// @Summary      Update a product
// @Description  Updates an existing product for the specified event
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ProductUpdateRequest true "Product update info with ID"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Product}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Failure      403  {object}  ProductStandardErrorResponse
// @Router       /events/{slug}/product [patch]
func (h *ProductHandler) UpdateEventProduct(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	var reqBody models.ProductUpdateRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "product")
		return
	}

	if reqBody.ProductID == "" {
		BadRequestError(w, errors.New("product ID is required"), "product")
		return
	}

	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	product, err := h.ProductService.UpdateEventProduct(user, slug, reqBody.ProductID, reqBody.Product)
	if err != nil {
		HandleErrMsg("error updating product", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, product, "", http.StatusOK)
}

// DeleteEventProduct godoc
// @Summary      Delete a product
// @Description  Deletes an existing product from the specified event
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ProductDeleteRequest true "Product deletion info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Failure      403  {object}  ProductStandardErrorResponse
// @Router       /events/{slug}/product [delete]
func (h *ProductHandler) DeleteEventProduct(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	var reqBody models.ProductDeleteRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "product")
		return
	}

	if reqBody.ProductID == "" {
		BadRequestError(w, errors.New("product ID is required"), "product")
		return
	}

	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	if err := h.ProductService.DeleteEventProduct(user, slug, reqBody.ProductID); err != nil {
		HandleErrMsg("error deleting product", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, nil, "deleted product", http.StatusOK)
}

// GetAllProductsFromEvent godoc
// @Summary      Get all products from an event
// @Description  Returns a list of all products for the specified event
// @Tags         products
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Product}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /events/{slug}/products [get]
func (h *ProductHandler) GetAllProductsFromEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	products, err := h.ProductService.GetAllProductsFromEvent(slug)
	if err != nil {
		HandleErrMsg("error getting products", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, products, "", http.StatusOK)
}

// PurchaseProducts godoc
// @Summary      Purchase products
// @Description  Processes a purchase of products for the authenticated user
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.PurchaseRequest true "Purchase info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Purchase}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /events/{slug}/purchase [post]
func (h *ProductHandler) PurchaseProducts(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	var reqBody models.PurchaseRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "product")
		return
	}

	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	purchase_info, err := h.ProductService.PurchaseProducts(user, slug, reqBody, w)
	if err != nil {
		HandleErrMsg("error processing purchase", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, purchase_info, "", http.StatusOK)
}

// ForcedPix godoc
// @Summary      Start a PIX purchase
// @Description  Creates a pending PIX purchase so the user can pay for the product
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.PixPurchaseRequest true "Purchase info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Purchase}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /events/{slug}/forced-pix [post]
func (h *ProductHandler) ForcedPix(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	var reqBody models.PurchaseRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "product")
		return
	}

	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	purchase_info, err := h.ProductService.ForcedPix(user, slug, reqBody)
	if err != nil {
		HandleErrMsg("error starting pix purchase", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, purchase_info, "", http.StatusOK)
}

type Approved struct {
	Status string `json:"status"`
}

// GetUserPurchases godoc
// @Summary      Get user purchases
// @Description  Returns a list of all purchases for the authenticated user
// @Tags         products
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Purchase}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /webhook/mp [post]
func (h *ProductHandler) MPWebhook(w http.ResponseWriter, r *http.Request) {
	var reqBody models.MP_WebhookRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "product")
		return
	}

	xSignature := r.Header.Get("x-signature")
	xRequestId := r.Header.Get("x-request-id")

	queryParams := r.URL.Query()
	dataID := queryParams.Get("data.id")
	parts := strings.Split(xSignature, ",")

	var ts, hash string
	for _, part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) == 2 {
			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])
			if key == "ts" {
				ts = value
			} else if key == "v1" {
				hash = value
			}
		}
	}

	secret := config.GetWebhookSignature()
	manifest := fmt.Sprintf("id:%v;request-id:%v;ts:%v;", dataID, xRequestId, ts)

	hmac := hmac.New(sha256.New, []byte(secret))
	hmac.Write([]byte(manifest))

	sha := hex.EncodeToString(hmac.Sum(nil))
	if sha == hash {
		handleSuccess(w, nil, "", http.StatusOK)
	} else {
		BadRequestError(w, errors.New("hmac verification failed"), "product")
		return
	}

	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%v", reqBody.Data.Id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	AuthHeader := fmt.Sprintf("Bearer %v", config.GetMercadoPagoAccessToken())
	req.Header.Add("Authorization", AuthHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	var bodyContent Approved
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&bodyContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// URGENT: CREATE FILE LOG
	PurchaseID, err := strconv.Atoi(reqBody.Data.Id)
	if err != nil {
		log.Println("ATTENTION COULD NOT FINISH PURCHASE")
		return
	}

	if bodyContent.Status == "approved" {
		purchase, err := h.ProductService.ProductRepo.GetPixPurchase(PurchaseID)
		if err != nil {
			log.Println("ATTENTION COULD NOT FINISH PURCHASE")
			return
		}

		err = h.ProductService.ProductRepo.FinalizePixPurchase(*purchase)
		if err != nil {
			log.Println("WTF: " + err.Error())
			return
		}

		err = h.ProductService.ProductRepo.DeletePixPurchase(PurchaseID)
		if err != nil {
			log.Println("Error deleting pix purchase")
		}
	}
}

// GetUserProducts godoc
// @Summary      Get user products
// @Description  Returns a list of all products for the authenticated user
// @Tags         products
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Product}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /user-products-relation [get]
func (h *ProductHandler) GetUserProductsRelation(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	products, err := h.ProductService.GetUserProductsRelation(user)
	if err != nil {
		HandleErrMsg("error getting products", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, products, "", http.StatusOK)
}

// GetUserProducts godoc
// @Summary      Get all bought user products
// @Description  Returns a list of all products for the authenticated user
// @Tags         products
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Product}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /user-products-relation [get]
func (h *ProductHandler) GetAllUserProductsRelation(w http.ResponseWriter, r *http.Request) {
	products, err := h.ProductService.GetAllUserProductsRelation()
	if err != nil {
		HandleErrMsg("error getting products", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, products, "", http.StatusOK)
}

// GetUserProducts godoc
// @Summary      Get user products
// @Description  Returns a list of all products for the authenticated user
// @Tags         products
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Product}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /user-products [get]
func (h *ProductHandler) GetUserProducts(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	products, err := h.ProductService.GetUserProducts(user)
	if err != nil {
		HandleErrMsg("error getting products", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, products, "", http.StatusOK)
}

// GetUserTokens godoc
// @Summary      Get user tokens
// @Description  Returns a list of all tokens for the authenticated user
// @Tags         products
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.UserToken}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /user-tokens [get]
func (h *ProductHandler) GetUserTokens(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	tokens, err := h.ProductService.GetUserTokens(user)
	if err != nil {
		HandleErrMsg("error getting tokens", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, tokens, "", http.StatusOK)
}

// GetUserPurchases godoc
// @Summary      Get user purchases
// @Description  Returns a list of all purchases for the authenticated user
// @Tags         products
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Purchase}
// @Failure      400  {object}  ProductStandardErrorResponse
// @Failure      401  {object}  ProductStandardErrorResponse
// @Router       /user-purchases [get]
func (h *ProductHandler) GetUserPurchases(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.ProductService.ProductRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "product")
		return
	}

	purchases, err := h.ProductService.GetUserPurchases(user)
	if err != nil {
		HandleErrMsg("error getting purchases", err, w).Stack("product").BadRequest()
		return
	}

	handleSuccess(w, purchases, "", http.StatusOK)
}
