package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
)

type UserHandler struct {
	service     model.UserService
	roleService model.RoleService
	httpsConfig bool
	lineClient  model.LineClient
}

func NewUserHandler(service model.UserService, roleService model.RoleService, httpsConfig bool, lineClient model.LineClient) *UserHandler {
	return &UserHandler{service: service, roleService: roleService, lineClient: lineClient}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var res Response
	var creds model.LoginCredentials
	var err error
	if err = json.NewDecoder(r.Body).Decode(&creds); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	issueTime := time.Now()
	expTime := time.Now().Add(24 * time.Hour)
	clientInfo := auth.GetClientInfo(r)
	user, tokenString, err := h.service.LoginUserJwt(r.Context(), &creds, issueTime, expTime, clientInfo)
	if err != nil {
		var code int
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, model.ErrInvalidLogin) {
			res.Message = model.ErrInvalidLogin.Error()
			code = http.StatusBadRequest
		} else if errors.Is(err, model.ErrNotActive) {
			res.Message = model.ErrNotActive.Error()
			code = http.StatusBadRequest
		} else {
			res.Message = "Error"
			code = http.StatusInternalServerError
			slog.ErrorContext(r.Context(), res.Message,
				slog.String("track", err.Error()),
			)
		}
		respondJson(w, code, &res)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    tokenString,
		Expires:  expTime,
		HttpOnly: true,                    // CRITICAL: Prevents JavaScript/XSS from reading the cookie
		Secure:   h.httpsConfig,           // CRITICAL: Ensures cookie is only sent over HTTPS (set to false ONLY if testing on localhost HTTP)
		SameSite: http.SameSiteStrictMode, // Protects against Cross-Site Request Forgery (CSRF)
		Path:     "/",
	})

	res.Data = map[string]any{
		"userId":    user.UserId,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
	}
	respondJson(w, http.StatusOK, &res)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Expires:  time.Unix(0, 0), // Set date to the past to delete it
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Permission(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error

	userId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}
	permissions, err := h.service.GetPermissionMapByUserId(r.Context(), userId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = permissions
	respondJson(w, http.StatusOK, &res)
}

func (h *UserHandler) GetAllDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	users, err := h.service.GetAllDetail(r.Context(), true)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
	}

	res.Data = users
	respondJson(w, http.StatusOK, &res)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var res Response
	var createUserReq model.CreateUser
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "User",
		ActionName: "Create",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&createUserReq); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if createUserReq.Username == "" || createUserReq.Password == "" {
		res.Message = "Username and password are required"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if createUserReq.FirstName == "" || createUserReq.LastName == "" {
		res.Message = "FirstName and LastName are required"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	_, err = h.service.CreateUser(r.Context(), &createUserReq, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			res.Message = model.ErrDuplicate.Error()
			respondJson(w, http.StatusBadRequest, &res)
		} else {
			res.Message = "Error"
			slog.ErrorContext(r.Context(), res.Message,
				slog.String("track", err.Error()),
			)
			respondJson(w, http.StatusInternalServerError, &res)
		}
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var res Response
	var updateUserReq model.UpdateUser
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "User",
		ActionName: "Update",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&updateUserReq); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if updateUserReq.FirstName == "" || updateUserReq.LastName == "" {
		res.Message = "FirstName and LastName are required"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	_, err = h.service.UpdateUser(r.Context(), &updateUserReq, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			res.Message = model.ErrDuplicate.Error()
			respondJson(w, http.StatusBadRequest, &res)
		} else {
			res.Message = "Error"
			slog.ErrorContext(r.Context(), res.Message,
				slog.String("track", err.Error()),
			)
			respondJson(w, http.StatusInternalServerError, &res)
		}
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "User",
		ActionName: "Delete",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	idStr := r.PathValue("id")
	deleteUserId, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.DeleteUser(r.Context(), deleteUserId, authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *UserHandler) LinkLine(w http.ResponseWriter, r *http.Request) {
	var res Response
	var linkLine model.UserLinkLine
	var err error

	if err = json.NewDecoder(r.Body).Decode(&linkLine); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	userSub, err := h.lineClient.VerifyIDToken(linkLine.IdToken)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}
	linkLine.LineUserToken = userSub

	err = h.service.UserLinkLineUserToken(r.Context(), linkLine, 0)
	if err != nil {
		code := http.StatusInternalServerError
		switch {
		case errors.Is(err, model.ErrInvalidLogin):
			code = http.StatusUnauthorized
			res.Message = model.ErrInvalidLogin.Error()

		case errors.Is(err, model.ErrNotActive):
			code = http.StatusForbidden
			res.Message = model.ErrNotActive.Error()

		case errors.Is(err, model.ErrDuplicate):
			code = http.StatusConflict
			res.Message = model.ErrDuplicate.Error()

		default:
			slog.ErrorContext(r.Context(), "Internal server error",
				slog.String("track", err.Error()),
			)
		}
		respondJson(w, code, &res)
		return
	}

	respondJson(w, http.StatusOK, &res)
}
