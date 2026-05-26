package api

import (
	"errors"
	"net/http"
	"time"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	token "github.com/Jingqi0327/eleclog/token"
	util "github.com/Jingqi0327/eleclog/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"` //alphanum表示用户名只能包含字母和数字
	Password string `json:"password" binding:"required,min=6"`    //min=6表示密码至少要有6个字符
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"role"`       
	Email    string `json:"email" binding:"required,email"` //email表示邮箱格式必须正确
}

// 自定义不包含hashPassword的响应体
type UserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	Role              string    `json:"role"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// 将数据库中的User转换为API响应体UserResponse,不包含敏感信息
func newUserResponse(user db.User) UserResponse {
	return UserResponse{
		Username:  user.Username,
		FullName:  user.FullName,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	currentUserRole := payload.Role
	requestedRole := req.Role
	if currentUserRole == util.ManagerRole && requestedRole != util.UserRole {
		ctx.JSON(http.StatusForbidden, errorResponse(errors.New("managers can only create users")))
		return
	}

	hashPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPassword,
		FullName:       req.FullName,
		Email:          req.Email,
		Role:           req.Role,
	}

	user, err := server.store.CreateUser(ctx, arg)

	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserResponse(user)

	ctx.JSON(http.StatusOK, rsp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken          string       `json:"access_token"`
	AccessTokenExpiresAt time.Time    `json:"access_token_expires_at"`
	User                 UserResponse `json:"user"` //返回用户信息，方便前端展示
}

// 登录用户，验证用户名和密码是否正确，如果正确则生成一个访问令牌返回给客户端
func (server *Server) loginUser(ctx *gin.Context) {
	// 1. 解析请求体，获取用户名和密码
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 2. 从数据库中获取用户信息，验证用户名是否存在
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 3. 验证密码是否正确
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// 4. 生成访问令牌，返回给客户端
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		user.Role,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 6. 返回响应体，包含访问令牌和刷新令牌，以及用户信息
	rsp := loginUserResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiresAt.Time,
		User:                 newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)
}

type UpdateUserRequest struct {
	Username string  `json:"username" binding:"required,alphanum"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=6"` //omitempty表示如果密码字段为空，则不进行验证
	FullName *string `json:"full_name,omitempty"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	Role     *string `json:"role,omitempty" binding:"omitempty,role"`
}

type UpdateUserResponse struct {
	User UserResponse `json:"user"`
}

func (server *Server) UpdateUser(ctx *gin.Context) {
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if payload.Username != req.Username && payload.Role == util.UserRole {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("cannot update other user's info")))
		return
	}

	arg := db.UpdateUserParams{
		Username: req.Username,
	}

	if req.FullName != nil {
		arg.FullName = pgtype.Text{
			String: *req.FullName,
			Valid:  true,
		}
	}

	if req.Email != nil {
		arg.Email = pgtype.Text{
			String: *req.Email,
			Valid:  true,
		}
	}

	if req.Role != nil {
		if payload.Role != util.AdminRole {
			ctx.JSON(http.StatusForbidden, errorResponse(errors.New("cannot update other user's role")))
			return
		}
		arg.Role = pgtype.Text{
			String: *req.Role,
			Valid:  true,
		}
	}

	if req.Password != nil {
		hashPassword, err := util.HashPassword(*req.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		arg.HashedPassword = pgtype.Text{
			String: hashPassword,
			Valid:  true,
		}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := UpdateUserResponse{
		User: newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)

}

type ListUsersRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=20"`
}

type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
}

func (server *Server) ListUsers(ctx *gin.Context) {
	var req ListUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Offset: (req.PageID - 1) * req.PageSize,
		Limit:  req.PageSize,
	}

	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	total, err := server.store.CountUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	usersrsp := make([]UserResponse, len(users))
	for i, user := range users {
		usersrsp[i] = newUserResponse(user)
	}

	rsp := ListUsersResponse{
		Users: usersrsp,
		Total: total,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type DeleteUserRequest struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

func (server *Server) deleteUser(ctx *gin.Context) {
	var req DeleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	deleteUser, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if (deleteUser.Role == util.ManagerRole || deleteUser.Role == util.AdminRole) && payload.Role != util.AdminRole {
		ctx.JSON(http.StatusForbidden, errorResponse(errors.New("you don't have permission to delete this user")))
		return
	} else if deleteUser.Role == util.AdminRole {
		ctx.JSON(http.StatusForbidden, errorResponse(errors.New("you can't delete admin")))
		return
	}

	err = server.store.DeleteUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
