package handler

import (
	"net/http"

	"sql2api/internal/middleware"
	"sql2api/internal/model"
	"sql2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ResourceHandler 统一资源处理器
type ResourceHandler struct {
	userService service.UserService
	itemService service.ItemService
}

// NewResourceHandler 创建资源处理器
func NewResourceHandler(userService service.UserService, itemService service.ItemService) *ResourceHandler {
	return &ResourceHandler{
		userService: userService,
		itemService: itemService,
	}
}

// HandleResource 统一资源处理端点
// @Summary 统一资源操作
// @Description 通过 action 字段分发不同的资源操作
// @Tags 资源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.ResourceRequest true "资源操作请求"
// @Success 200 {object} model.SuccessResponse "操作成功"
// @Failure 400 {object} model.ErrorResponse "请求格式错误"
// @Failure 401 {object} model.ErrorResponse "未认证"
// @Failure 403 {object} model.ErrorResponse "权限不足"
// @Failure 404 {object} model.ErrorResponse "资源不存在"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /api/v1/resource [post]
func (h *ResourceHandler) HandleResource(c *gin.Context) {
	var req model.ResourceRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid request format",
			err.Error(),
		))
		return
	}

	// 验证必填字段
	if req.Action == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Action field is required",
		))
		return
	}

	if req.Resource == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Resource field is required",
		))
		return
	}

	// 根据资源类型分发
	switch req.Resource {
	case "user":
		h.handleUserResource(c, &req)
	case "item":
		h.handleItemResource(c, &req)
	default:
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Unsupported resource type",
			"Supported resources: user, item",
		))
	}
}

// handleUserResource 处理用户资源操作
func (h *ResourceHandler) handleUserResource(c *gin.Context, req *model.ResourceRequest) {
	switch req.Action {
	case string(model.ActionList):
		h.listUsers(c, req)
	case string(model.ActionGet):
		h.getUser(c, req)
	case string(model.ActionUpdate):
		h.updateUser(c, req)
	case string(model.ActionDelete):
		h.deleteUser(c, req)
	case string(model.ActionSetActive):
		h.setUserActive(c, req)
	default:
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Unsupported action for user resource",
			"Supported actions: list, get, update, delete, set_active",
		))
	}
}

// handleItemResource 处理项目资源操作
func (h *ResourceHandler) handleItemResource(c *gin.Context, req *model.ResourceRequest) {
	switch req.Action {
	case string(model.ActionCreate):
		h.createItem(c, req)
	case string(model.ActionList):
		h.listItems(c, req)
	case string(model.ActionGet):
		h.getItem(c, req)
	case string(model.ActionUpdate):
		h.updateItem(c, req)
	case string(model.ActionDelete):
		h.deleteItem(c, req)
	case string(model.ActionSearch):
		h.searchItems(c, req)
	case string(model.ActionSetActive):
		h.setItemActive(c, req)
	default:
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Unsupported action for item resource",
			"Supported actions: create, list, get, update, delete, search, set_active",
		))
	}
}

// listUsers 获取用户列表
func (h *ResourceHandler) listUsers(c *gin.Context, req *model.ResourceRequest) {
	// 解析分页参数
	page := 1
	pageSize := 10
	orderBy := "id"
	order := "asc"

	if req.Data != nil {
		if p, ok := req.Data["page"]; ok {
			if pageFloat, ok := p.(float64); ok {
				page = int(pageFloat)
			}
		}
		if ps, ok := req.Data["page_size"]; ok {
			if pageSizeFloat, ok := ps.(float64); ok {
				pageSize = int(pageSizeFloat)
			}
		}
		if ob, ok := req.Data["order_by"]; ok {
			if orderByStr, ok := ob.(string); ok {
				orderBy = orderByStr
			}
		}
		if o, ok := req.Data["order"]; ok {
			if orderStr, ok := o.(string); ok {
				order = orderStr
			}
		}
	}

	users, total, err := h.userService.ListUsers(page, pageSize, orderBy, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			http.StatusInternalServerError,
			"Failed to list users",
			err.Error(),
		))
		return
	}

	// 转换为响应格式
	userResponses := make([]model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response := map[string]interface{}{
		"users":       userResponses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response, "Users retrieved successfully"))
}

// getUser 获取用户详情
func (h *ResourceHandler) getUser(c *gin.Context, req *model.ResourceRequest) {
	// 解析用户ID
	var userID uint
	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				userID = uint(idFloat)
			}
		}
	}

	if userID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"User ID is required",
		))
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(
			http.StatusNotFound,
			"User not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user.ToResponse(), "User retrieved successfully"))
}

// updateUser 更新用户信息
func (h *ResourceHandler) updateUser(c *gin.Context, req *model.ResourceRequest) {
	// 获取当前用户ID
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 解析要更新的用户ID
	var targetUserID uint
	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				targetUserID = uint(idFloat)
			}
		}
	}

	if targetUserID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"User ID is required",
		))
		return
	}

	// 检查权限：用户只能更新自己的信息
	if currentUserID != targetUserID {
		c.JSON(http.StatusForbidden, model.NewErrorResponse(
			http.StatusForbidden,
			"You can only update your own profile",
		))
		return
	}

	// 构造更新请求
	updateReq := &model.UserUpdateRequest{}
	if req.Data != nil {
		if email, ok := req.Data["email"]; ok {
			if emailStr, ok := email.(string); ok {
				updateReq.Email = &emailStr
			}
		}
		if fullName, ok := req.Data["full_name"]; ok {
			if fullNameStr, ok := fullName.(string); ok {
				updateReq.FullName = &fullNameStr
			}
		}
	}

	user, err := h.userService.UpdateUser(targetUserID, updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Failed to update user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user.ToResponse(), "User updated successfully"))
}

// deleteUser 删除用户
func (h *ResourceHandler) deleteUser(c *gin.Context, req *model.ResourceRequest) {
	// 获取当前用户ID
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 解析要删除的用户ID
	var targetUserID uint
	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				targetUserID = uint(idFloat)
			}
		}
	}

	if targetUserID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"User ID is required",
		))
		return
	}

	// 检查权限：用户只能删除自己的账户
	if currentUserID != targetUserID {
		c.JSON(http.StatusForbidden, model.NewErrorResponse(
			http.StatusForbidden,
			"You can only delete your own account",
		))
		return
	}

	err := h.userService.DeleteUser(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			http.StatusInternalServerError,
			"Failed to delete user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil, "User deleted successfully"))
}

// setUserActive 设置用户激活状态
func (h *ResourceHandler) setUserActive(c *gin.Context, req *model.ResourceRequest) {
	// 获取当前用户ID
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 解析参数
	var targetUserID uint
	var active bool

	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				targetUserID = uint(idFloat)
			}
		}
		if a, ok := req.Data["active"]; ok {
			if activeBool, ok := a.(bool); ok {
				active = activeBool
			}
		}
	}

	if targetUserID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"User ID is required",
		))
		return
	}

	// 检查权限：用户只能设置自己的激活状态
	if currentUserID != targetUserID {
		c.JSON(http.StatusForbidden, model.NewErrorResponse(
			http.StatusForbidden,
			"You can only modify your own account status",
		))
		return
	}

	err := h.userService.SetUserActive(targetUserID, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			http.StatusInternalServerError,
			"Failed to set user active status",
			err.Error(),
		))
		return
	}

	message := "User activated successfully"
	if !active {
		message = "User deactivated successfully"
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil, message))
}

// createItem 创建项目
func (h *ResourceHandler) createItem(c *gin.Context, req *model.ResourceRequest) {
	// 获取当前用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 构造创建请求
	createReq := &model.ItemCreateRequest{}
	if req.Data != nil {
		if name, ok := req.Data["name"]; ok {
			if nameStr, ok := name.(string); ok {
				createReq.Name = nameStr
			}
		}
		if value, ok := req.Data["value"]; ok {
			if valueFloat, ok := value.(float64); ok {
				createReq.Value = int64(valueFloat)
			}
		}
		if description, ok := req.Data["description"]; ok {
			if descStr, ok := description.(string); ok {
				createReq.Description = descStr
			}
		}
		if category, ok := req.Data["category"]; ok {
			if catStr, ok := category.(string); ok {
				createReq.Category = catStr
			}
		}
		if tags, ok := req.Data["tags"]; ok {
			if tagsStr, ok := tags.(string); ok {
				createReq.Tags = tagsStr
			}
		}
	}

	item, err := h.itemService.CreateItem(createReq, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Failed to create item",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(item.ToResponse(), "Item created successfully"))
}

// listItems 获取项目列表
func (h *ResourceHandler) listItems(c *gin.Context, req *model.ResourceRequest) {
	// 构造查询请求
	query := &model.ItemQueryRequest{
		Page:     1,
		PageSize: 10,
		OrderBy:  "id",
		Order:    "asc",
	}

	includeCreator := false

	if req.Data != nil {
		if p, ok := req.Data["page"]; ok {
			if pageFloat, ok := p.(float64); ok {
				query.Page = int(pageFloat)
			}
		}
		if ps, ok := req.Data["page_size"]; ok {
			if pageSizeFloat, ok := ps.(float64); ok {
				query.PageSize = int(pageSizeFloat)
			}
		}
		if name, ok := req.Data["name"]; ok {
			if nameStr, ok := name.(string); ok {
				query.Name = nameStr
			}
		}
		if category, ok := req.Data["category"]; ok {
			if catStr, ok := category.(string); ok {
				query.Category = catStr
			}
		}
		if active, ok := req.Data["is_active"]; ok {
			if activeBool, ok := active.(bool); ok {
				query.IsActive = &activeBool
			}
		}
		if ob, ok := req.Data["order_by"]; ok {
			if orderByStr, ok := ob.(string); ok {
				query.OrderBy = orderByStr
			}
		}
		if o, ok := req.Data["order"]; ok {
			if orderStr, ok := o.(string); ok {
				query.Order = orderStr
			}
		}
		if ic, ok := req.Data["include_creator"]; ok {
			if icBool, ok := ic.(bool); ok {
				includeCreator = icBool
			}
		}
	}

	response, err := h.itemService.ListItems(query, includeCreator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			http.StatusInternalServerError,
			"Failed to list items",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response, "Items retrieved successfully"))
}

// getItem 获取项目详情
func (h *ResourceHandler) getItem(c *gin.Context, req *model.ResourceRequest) {
	// 解析项目ID
	var itemID uint
	includeCreator := false

	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				itemID = uint(idFloat)
			}
		}
		if ic, ok := req.Data["include_creator"]; ok {
			if icBool, ok := ic.(bool); ok {
				includeCreator = icBool
			}
		}
	}

	if itemID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Item ID is required",
		))
		return
	}

	item, err := h.itemService.GetItemByID(itemID, includeCreator)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(
			http.StatusNotFound,
			"Item not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(item.ToResponse(), "Item retrieved successfully"))
}

// updateItem 更新项目
func (h *ResourceHandler) updateItem(c *gin.Context, req *model.ResourceRequest) {
	// 获取当前用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 解析项目ID
	var itemID uint
	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				itemID = uint(idFloat)
			}
		}
	}

	if itemID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Item ID is required",
		))
		return
	}

	// 构造更新请求
	updateReq := &model.ItemUpdateRequest{}
	if req.Data != nil {
		if name, ok := req.Data["name"]; ok {
			if nameStr, ok := name.(string); ok {
				updateReq.Name = &nameStr
			}
		}
		if value, ok := req.Data["value"]; ok {
			if valueFloat, ok := value.(float64); ok {
				valueInt := int64(valueFloat)
				updateReq.Value = &valueInt
			}
		}
		if description, ok := req.Data["description"]; ok {
			if descStr, ok := description.(string); ok {
				updateReq.Description = &descStr
			}
		}
		if category, ok := req.Data["category"]; ok {
			if catStr, ok := category.(string); ok {
				updateReq.Category = &catStr
			}
		}
		if tags, ok := req.Data["tags"]; ok {
			if tagsStr, ok := tags.(string); ok {
				updateReq.Tags = &tagsStr
			}
		}
		if active, ok := req.Data["is_active"]; ok {
			if activeBool, ok := active.(bool); ok {
				updateReq.IsActive = &activeBool
			}
		}
	}

	item, err := h.itemService.UpdateItem(itemID, updateReq, userID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "access denied: you can only modify your own items" {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, model.NewErrorResponse(
			statusCode,
			"Failed to update item",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(item.ToResponse(), "Item updated successfully"))
}

// deleteItem 删除项目
func (h *ResourceHandler) deleteItem(c *gin.Context, req *model.ResourceRequest) {
	// 获取当前用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 解析项目ID
	var itemID uint
	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				itemID = uint(idFloat)
			}
		}
	}

	if itemID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Item ID is required",
		))
		return
	}

	err := h.itemService.DeleteItem(itemID, userID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "access denied: you can only modify your own items" {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, model.NewErrorResponse(
			statusCode,
			"Failed to delete item",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil, "Item deleted successfully"))
}

// searchItems 搜索项目
func (h *ResourceHandler) searchItems(c *gin.Context, req *model.ResourceRequest) {
	// 解析搜索参数
	var keyword string
	page := 1
	pageSize := 10

	if req.Data != nil {
		if k, ok := req.Data["keyword"]; ok {
			if keywordStr, ok := k.(string); ok {
				keyword = keywordStr
			}
		}
		if p, ok := req.Data["page"]; ok {
			if pageFloat, ok := p.(float64); ok {
				page = int(pageFloat)
			}
		}
		if ps, ok := req.Data["page_size"]; ok {
			if pageSizeFloat, ok := ps.(float64); ok {
				pageSize = int(pageSizeFloat)
			}
		}
	}

	if keyword == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Search keyword is required",
		))
		return
	}

	response, err := h.itemService.SearchItems(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			http.StatusInternalServerError,
			"Failed to search items",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response, "Items searched successfully"))
}

// setItemActive 设置项目激活状态
func (h *ResourceHandler) setItemActive(c *gin.Context, req *model.ResourceRequest) {
	// 获取当前用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 解析参数
	var itemID uint
	var active bool

	if req.Data != nil {
		if id, ok := req.Data["id"]; ok {
			if idFloat, ok := id.(float64); ok {
				itemID = uint(idFloat)
			}
		}
		if a, ok := req.Data["active"]; ok {
			if activeBool, ok := a.(bool); ok {
				active = activeBool
			}
		}
	}

	if itemID == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Item ID is required",
		))
		return
	}

	err := h.itemService.SetItemActive(itemID, active, userID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "access denied: you can only modify your own items" {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, model.NewErrorResponse(
			statusCode,
			"Failed to set item active status",
			err.Error(),
		))
		return
	}

	message := "Item activated successfully"
	if !active {
		message = "Item deactivated successfully"
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil, message))
}
