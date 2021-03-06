package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	retrospectiveServices "github.com/iReflect/reflect-app/apps/retrospective/services"
	"github.com/iReflect/reflect-app/constants"
)

// SprintController ...
type SprintController struct {
	SprintService     retrospectiveServices.SprintService
	PermissionService retrospectiveServices.PermissionService
	TrailService      retrospectiveServices.TrailService
}

// Routes for Sprints
func (ctrl SprintController) Routes(r *gin.RouterGroup) {
	r.GET("/", ctrl.List)
	r.POST("/", ctrl.Create)
	r.GET("/:sprintID/", ctrl.Get)
	r.DELETE("/:sprintID/", ctrl.Delete)
	r.PUT("/:sprintID/", ctrl.Update)

	r.POST("/:sprintID/activate/", ctrl.ActivateSprint)
	r.POST("/:sprintID/freeze/", ctrl.FreezeSprint)
	r.POST("/:sprintID/process/", ctrl.Process)

	r.GET("/:sprintID/member-summary/", ctrl.GetSprintMemberSummary)

	r.GET("/:sprintID/process_history/", ctrl.GetTrails)
}

// List the sprints accessible to the user
func (ctrl SprintController) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")
	after, _ := c.GetQuery("after")
	perPage, _ := c.GetQuery("count")

	perPageInt, err := strconv.Atoi(perPage)
	if err != nil || perPageInt < 0 {
		perPageInt = 20
	}

	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprints, status, err := ctrl.SprintService.GetSprintsList(retroID, userID.(uint), perPageInt, after)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, sprints)
}

// Create a new draft sprint for the retro
func (ctrl SprintController) Create(c *gin.Context) {
	userID, _ := c.Get("userID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessRetro(retroID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprintData := retroSerializers.CreateSprintSerializer{CreatedByID: userID.(uint)}
	err := c.BindJSON(&sprintData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	sprint, status, err := ctrl.SprintService.Create(retroID, userID.(uint), sprintData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.CreatedSprint,
		constants.Sprint,
		strconv.Itoa(int(sprint.ID)),
		userID.(uint))

	c.JSON(http.StatusCreated, sprint)
}

// Get Sprint Data
func (ctrl SprintController) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprint, status, err := ctrl.SprintService.Get(sprintID, userID.(uint), true)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, sprint)
}

// Delete Sprint
func (ctrl SprintController) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	status, err := ctrl.SprintService.DeleteSprint(sprintID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.DeletedSprint,
		constants.Sprint,
		sprintID,
		userID.(uint))

	c.JSON(status, nil)
}

// Update a sprint
func (ctrl SprintController) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	var sprintData retroSerializers.UpdateSprintSerializer
	if err := c.BindJSON(&sprintData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	response, status, err := ctrl.SprintService.UpdateSprint(sprintID, userID.(uint), sprintData)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.UpdatedSprint,
		constants.Sprint,
		sprintID,
		userID.(uint))

	c.JSON(status, response)
}

// ActivateSprint activates the given sprint
func (ctrl SprintController) ActivateSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	status, err := ctrl.SprintService.ActivateSprint(sprintID, retroID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.ActivatedSprint,
		constants.Sprint,
		sprintID,
		userID.(uint))

	c.JSON(status, nil)
}

// FreezeSprint freezes the given sprint
func (ctrl SprintController) FreezeSprint(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanEditSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	status, err := ctrl.SprintService.FreezeSprint(sprintID, retroID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	ctrl.TrailService.Add(
		constants.FreezeSprint,
		constants.Sprint,
		sprintID,
		userID.(uint))

	c.JSON(status, nil)
}

// Process Sprint
func (ctrl SprintController) Process(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	sprintIDInt, err := strconv.Atoi(sprintID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid sprint"})
	}

	sprint, _, err := ctrl.SprintService.Get(sprintID, userID.(uint), false)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "sprint not found"})
	}

	ctrl.SprintService.QueueSprint(uint(sprintIDInt), sprint.Status == retroModels.ActiveSprint)

	ctrl.TrailService.Add(
		constants.TriggeredSprintRefresh,
		constants.Sprint,
		sprintID,
		userID.(uint))

	c.JSON(http.StatusNoContent, nil)
}

// GetSprintMemberSummary returns the sprint member summary list
func (ctrl SprintController) GetSprintMemberSummary(c *gin.Context) {
	userID, _ := c.Get("userID")
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")

	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	response, status, err := ctrl.SprintService.GetSprintMembersSummary(sprintID)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(status, response)
}

// GetTrails is method to get the all trails related to a particular sprint
func (ctrl SprintController) GetTrails(c *gin.Context) {
	sprintID := c.Param("sprintID")
	retroID := c.Param("retroID")
	userID, _ := c.Get("userID")
	sprintIDInt, errConversion := strconv.Atoi(sprintID)

	if errConversion != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	if !ctrl.PermissionService.UserCanAccessSprint(retroID, sprintID, userID.(uint)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}

	trails, status, err := ctrl.TrailService.GetTrails(uint(sprintIDInt))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{})
		return
	}
	c.JSON(status, trails)
}
