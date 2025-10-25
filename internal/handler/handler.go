package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AleGaliev/gofermart/internal/middleware"
	models "github.com/AleGaliev/gofermart/internal/model"
	"github.com/AleGaliev/gofermart/internal/service/hash"
	"github.com/AleGaliev/gofermart/internal/service/jwtmanager"
	"github.com/AleGaliev/gofermart/internal/service/order"
	"github.com/go-chi/chi/v5"
)

type JWTManager interface {
	IssueJWT(models.User) (string, error)
	GetLoginFromToken(tokenString string) (string, error)
}

type Storage interface {
	Register(user models.User) error
	Login(user models.User) error
	UploadOrder(orderUser, orderNumber string) error
	GetOrder(OrderNumber string) (string, error)
	GetOrdersUser(user string) ([]models.Order, error)
	GetUserBalance(user string) (models.UserBalance, error)
	GetUserWithdrawals(user string) ([]models.Withdraw, error)
	UploadOrderWithdraw(user string, withdraw models.Withdraw) error
}
type MyHandler struct {
	storage    Storage
	jwtManager JWTManager
}

func CreateMyHandler(storage Storage, logger middleware.Logger, hashKey string) http.Handler {
	jwtManagerHandler := jwtmanager.NewJWTManager(hashKey)
	h := &MyHandler{
		storage:    storage,
		jwtManager: jwtManagerHandler,
	}

	mux := chi.NewRouter()

	mux.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.ServeHTTPRegister)
		r.Post("/login", h.ServeHTTPLogin)
		r.With(middleware.AuthMiddleware(h.jwtManager)).Post("/orders", h.ServeHTTPOrdersNumders)
		r.With(middleware.AuthMiddleware(h.jwtManager)).Post("/balance/withdraw", h.ServeHTTPBalanceWithDraw)
		r.With(middleware.AuthMiddleware(h.jwtManager)).Get("/orders", h.ServeHTTPOrdersInfo)
		r.With(middleware.AuthMiddleware(h.jwtManager)).Get("/balance", h.ServeHTTPBalance)
		r.With(middleware.AuthMiddleware(h.jwtManager)).Get("/withdrawals", h.ServeHTTPWithdrawals)
	})

	muxGzip := middleware.GzipMiddlewareHandler(mux)
	muxMiddlewareLogger := middleware.MiddlewareHandlerLogger(muxGzip, logger)

	return muxMiddlewareLogger
}

func (h *MyHandler) ServeHTTPRegister(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		failedResponse(w, "Invalid request format", "", http.StatusBadRequest)
		return
	}

	if err := h.storage.Register(user); err != nil {
		failedResponse(w, err.Error(), "", http.StatusConflict)
		fmt.Println(err)
		return
	}

	token, err := h.jwtManager.IssueJWT(user)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		MaxAge:   3600,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	successResponse(w, "", http.StatusOK)

}

func (h *MyHandler) ServeHTTPLogin(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		failedResponse(w, "Invalid request format", "", http.StatusBadRequest)
		return
	}

	if err := h.storage.Login(user); err != nil {
		failedResponse(w, "incorrect login password", "", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtManager.IssueJWT(user)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		MaxAge:   3600,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	successResponse(w, "", http.StatusOK)

}

func (h *MyHandler) ServeHTTPBalance(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("X-User-Login")

	balance, err := h.storage.GetUserBalance(login)

	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(balance)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json ")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *MyHandler) ServeHTTPOrdersNumders(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("X-User-Login")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		failedResponse(w, "Invalid request format", "", http.StatusBadRequest)
		return
	}

	if err = order.UploadOrder(login, string(body), h.storage); err != nil {
		switch err.Error() {
		case "not valid order number":
			failedResponse(w, "Invalid order number", "", http.StatusUnprocessableEntity)
		case "invalid order user":
			failedResponse(w, "invalid order user", "", http.StatusConflict)
		case "the order number was already taken by this user":
			successResponse(w, "", http.StatusOK)
		default:
			failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		}
		return
	}

	successResponse(w, "", http.StatusAccepted)
}

func (h *MyHandler) ServeHTTPOrdersInfo(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("X-User-Login")

	orders, err := h.storage.GetOrdersUser(login)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
	}
	if len(orders) == 0 {
		successResponse(w, "", http.StatusNoContent)
	}

	jsonBytes, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "application/json ")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}

func (h *MyHandler) ServeHTTPWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (h *MyHandler) ServeHTTPBalanceWithDraw(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("X-User-Login")

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var withdraw models.Withdraw

	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		failedResponse(w, "Invalid request format", "", http.StatusBadRequest)
		return
	}
	if err := order.UploadOrderWithdraw(login, withdraw, h.storage); err != nil {
		switch err.Error() {
		case "not valid order number":
			failedResponse(w, err.Error(), "", http.StatusUnprocessableEntity)
		case "insufficient funds":
			failedResponse(w, err.Error(), "", http.StatusPaymentRequired)
		case "the order number exists":
			successResponse(w, "", http.StatusAccepted)
		default:
			failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		}
		return
	}
	successResponse(w, "", http.StatusOK)
}

func (h *MyHandler) ServeHTTPWithdrawals(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("X-User-Login")

	withdrawals, err := h.storage.GetUserWithdrawals(login)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		successResponse(w, "", http.StatusNoContent)
		return
	}
	jsonBytes, err := json.Marshal(withdrawals)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "application/json ")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func successResponse(res http.ResponseWriter, hashKey string, statusCode int) {
	res.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  `success`,
		"message": "Запрос обработан",
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	if hashKey != "" {
		hashSHA256 := hash.CreateHash(hashKey, jsonBytes)
		res.Header().Set("HashSHA256", hashSHA256)
	}
	res.WriteHeader(statusCode)
	_, err = res.Write(jsonBytes)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func failedResponse(res http.ResponseWriter, message, hashKey string, statusCode int) {
	res.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  `failed`,
		"message": message,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
	if hashKey != "" {
		hashSHA256 := hash.CreateHash(hashKey, jsonBytes)
		res.Header().Set("HashSHA256", hashSHA256)
	}
	res.WriteHeader(statusCode)
	_, err = res.Write(jsonBytes)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}
