package controllers

import (
	"database/sql"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/DavidDevGt/go-finance/database"
	"github.com/DavidDevGt/go-finance/models"
	"github.com/gin-gonic/gin"
)

// ---------- Helpers ----------

func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func respondWithSuccess(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

func parseWeekParam(c *gin.Context) (int, bool) {
	weekParam := c.Param("week")
	week, err := strconv.Atoi(weekParam)
	if err != nil || week < 1 || week > 53 {
		respondWithError(c, http.StatusBadRequest, "Número de semana inválido")
		return 0, false
	}
	return week, true
}

// ---------- Controllers ----------

// SetWeeklyBudget godoc
// @Summary Establecer o actualizar presupuesto semanal
// @Description Define el monto máximo que se puede gastar en una semana específica.
// @Tags Budget
// @Accept json
// @Produce json
// @Param budget body models.WeeklyBudget true "Datos del presupuesto"
// @Success 200 {object} models.WeeklyBudget
// @Failure 400 {object} map[string]string
// @Router /api/budget [post]
func SetWeeklyBudget(c *gin.Context) {
	var input models.WeeklyBudget
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.WeeklyBudget
	db := database.DB

	err := db.Where("week = ? AND year = ?", input.Week, input.Year).First(&existing).Error
	if err == nil {
		existing.Amount = input.Amount
		db.Save(&existing)
		c.JSON(http.StatusOK, existing)
		return
	}

	db.Create(&input)
	c.JSON(http.StatusCreated, input)
}

// GetWeeklySummary godoc
// @Summary Obtener resumen del presupuesto semanal
// @Description Devuelve cuánto se gastó en una semana y cuánto queda disponible.
// @Tags Budget
// @Produce json
// @Param year path int true "Año"
// @Param week path int true "Semana del año"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/budget/{year}/{week} [get]
func GetWeeklySummary(c *gin.Context) {
	weekParam := c.Param("week")
	yearParam := c.Param("year")
	week, _ := strconv.Atoi(weekParam)
	year, _ := strconv.Atoi(yearParam)

	var budget models.WeeklyBudget
	err := database.DB.Where("week = ? AND year = ?", week, year).First(&budget).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "No hay presupuesto registrado para esta semana",
			"no_budget":   true,
			"week":        week,
			"year":        year,
			"budget":      0,
			"spent":       0,
			"remaining":   0,
			"over_budget": false,
		})
		return
	}

	var result sql.NullFloat64
	database.DB.Model(&models.Expense{}).
		Select("sum(amount)").
		Where("week = ? AND strftime('%Y', date_raw) = ?", week, strconv.Itoa(year)).
		Scan(&result)

	totalSpent := 0.0
	if result.Valid {
		totalSpent = result.Float64
	}

	remaining := budget.Amount - totalSpent

	c.JSON(http.StatusOK, gin.H{
		"week":        week,
		"year":        year,
		"budget":      budget.Amount,
		"spent":       totalSpent,
		"remaining":   remaining,
		"over_budget": remaining < 0,
	})
}

// GetExpenses godoc
// @Summary Listar todos los gastos
// @Description Devuelve un listado de todos los gastos registrados.
// @Tags Expenses
// @Produce json
// @Success 200 {array} models.Expense
// @Failure 500 {object} map[string]string
// @Router /api/expenses [get]
func GetExpenses(c *gin.Context) {
	var expenses []models.Expense
	if err := database.DB.Order("date_raw desc").Find(&expenses).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	for i := range expenses {
		expenses[i].Date = models.FormattedDate(expenses[i].DateRaw)
	}
	respondWithSuccess(c, http.StatusOK, expenses)
}

func GetExpensesByWeek(c *gin.Context) {
	week, ok := parseWeekParam(c)
	if !ok {
		return
	}

	var expenses []models.Expense
	if err := database.DB.Where("week = ?", week).Order("date_raw desc").Find(&expenses).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, expenses)
}

// ExportExpensesByWeekToCSV godoc
// @Summary Exportar gastos a CSV por semana
// @Description Genera y descarga un archivo CSV con todos los gastos de una semana específica.
// @Tags Expenses
// @Produce text/csv
// @Param week path int true "Semana del año"
// @Success 200
// @Failure 404 {object} map[string]string
// @Router /api/expenses/week/{week}/export [get]
func ExportExpensesByWeekToCSV(c *gin.Context) {
	week, ok := parseWeekParam(c)
	if !ok {
		return
	}

	var expenses []models.Expense
	if err := database.DB.Where("week = ?", week).Order("date_raw desc").Find(&expenses).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range expenses {
		expenses[i].Date = models.FormattedDate(expenses[i].DateRaw)
	}
	if len(expenses) == 0 {
		respondWithError(c, http.StatusNotFound, "No hay gastos para exportar en esta semana")
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=expenses_week_"+strconv.Itoa(week)+".csv")
	c.Header("Cache-Control", "no-cache")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	writer.Write([]string{"ID", "Titulo", "Descripcion", "Monto", "Categoría", "Fecha", "Semana"})

	for _, expense := range expenses {
		record := []string{
			strconv.Itoa(int(expense.ID)),
			expense.Title,
			expense.Description,
			strconv.FormatFloat(expense.Amount, 'f', 2, 64),
			expense.Category,
			time.Time(expense.Date).Format("02-01-2006"),
			strconv.Itoa(expense.Week),
		}
		writer.Write(record)
	}
}

// GetExpenseByID godoc
// @Summary Obtener gasto por ID
// @Description Devuelve el detalle de un gasto específico.
// @Tags Expenses
// @Produce json
// @Param id path int true "ID del gasto"
// @Success 200 {object} models.Expense
// @Failure 404 {object} map[string]string
// @Router /api/expenses/{id} [get]
func GetExpenseByID(c *gin.Context) {
	id := c.Param("id")
	var expense models.Expense

	if err := database.DB.First(&expense, id).Error; err != nil {
		respondWithError(c, http.StatusNotFound, "Gasto no encontrado")
		return
	}

	expense.Date = models.FormattedDate(expense.DateRaw)
	respondWithSuccess(c, http.StatusOK, expense)
}

// CreateExpense godoc
// @Summary Crear un nuevo gasto
// @Description Crea un gasto con título, monto, categoría y fecha.
// @Tags Expenses
// @Accept json
// @Produce json
// @Param expense body map[string]interface{} true "Datos del gasto"
// @Success 201 {object} models.Expense
// @Failure 400 {object} map[string]string
// @Router /api/expenses [post]
func CreateExpense(c *gin.Context) {
	var input struct {
		Title       string  `json:"title" binding:"required"`
		Description string  `json:"description"`
		Amount      float64 `json:"amount" binding:"required"`
		Category    string  `json:"category"`
		Date        string  `json:"date" binding:"required"` // se espera en formato YYYY-MM-DD
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Formato de fecha inválido, usa YYYY-MM-DD")
		return
	}

	expense := models.Expense{
		Title:       input.Title,
		Description: input.Description,
		Amount:      input.Amount,
		Category:    input.Category,
		DateRaw:     parsedDate,
		Date:        models.FormattedDate(parsedDate),
		Week:        models.CalculateWeek(parsedDate),
	}

	if err := database.DB.Create(&expense).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusCreated, expense)
}

// UpdateExpense godoc
// @Summary Actualizar un gasto
// @Description Actualiza los datos de un gasto existente.
// @Tags Expenses
// @Accept json
// @Produce json
// @Param id path int true "ID del gasto"
// @Param expense body map[string]interface{} true "Datos a actualizar"
// @Success 200 {object} models.Expense
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/expenses/{id} [put]
func UpdateExpense(c *gin.Context) {
	id := c.Param("id")
	var expense models.Expense

	if err := database.DB.First(&expense, id).Error; err != nil {
		respondWithError(c, http.StatusNotFound, "Gasto no encontrado")
		return
	}

	var input struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Amount      float64 `json:"amount"`
		Category    string  `json:"category"`
		Date        string  `json:"date"` // se espera en formato YYYY-MM-DD
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if input.Title != "" {
		expense.Title = input.Title
	}
	if input.Description != "" {
		expense.Description = input.Description
	}
	if input.Amount != 0 {
		expense.Amount = input.Amount
	}
	if input.Category != "" {
		expense.Category = input.Category
	}
	if input.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", input.Date)
		if err != nil {
			respondWithError(c, http.StatusBadRequest, "Formato de fecha inválido, usa YYYY-MM-DD")
			return
		}
		expense.DateRaw = parsedDate
		expense.Week = models.CalculateWeek(parsedDate)
	}

	if err := database.DB.Save(&expense).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	expense.Date = models.FormattedDate(expense.DateRaw)
	respondWithSuccess(c, http.StatusOK, expense)
}

// DeleteExpense godoc
// @Summary Eliminar un gasto
// @Description Elimina un gasto existente por ID.
// @Tags Expenses
// @Param id path int true "ID del gasto"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/expenses/{id} [delete]
func DeleteExpense(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := database.DB.Delete(&models.Expense{}, parsedID).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Gasto eliminado"})
}
