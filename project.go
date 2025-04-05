package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jung-kurt/gofpdf"
	"log"
	"net/http"
)

var db *gorm.DB

func init() {
	var err error
	dsn := "user=turar password=2222 dbname=datebase sslmode=disable"
	db, err = gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Text{})
}

type Text struct {
	ID   uint   `json:"id"`
	Text string `json:"text"`
}

func main() {
	r := gin.Default()
	r.POST("/create-pdf", func(c *gin.Context) {
		var json struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		newText := Text{Text: json.Text}
		if err := db.Create(&newText).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить текст в базу данных"})
			return
		}
		outputFile := "output.pdf"
		err := generatePDF(json.Text, outputFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.File(outputFile)
	})
	r.GET("/texts", func(c *gin.Context) {
		var texts []Text
		if err := db.Find(&texts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить тексты"})
			return
		}
		c.JSON(http.StatusOK, texts)
	})
	r.GET("/texts/:id", func(c *gin.Context) {
		id := c.Param("id")
		var text Text
		if err := db.First(&text, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Текст не найден"})
			return
		}
		c.JSON(http.StatusOK, text)
	})
	r.PUT("/texts/:id", func(c *gin.Context) {
		id := c.Param("id")
		var text Text
		if err := db.First(&text, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Текст не найден"})
			return
		}
		var updatedText struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&updatedText); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		text.Text = updatedText.Text
		db.Save(&text)
		c.JSON(http.StatusOK, text)
	})

	r.DELETE("/texts/:id", func(c *gin.Context) {
		id := c.Param("id")
		var text Text
		if err := db.First(&text, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Текст не найден"})
			return
		}

		if err := db.Delete(&text).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить текст"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Текст успешно удалён"})
	})

	r.Run(":8080")
}

func generatePDF(text string, outputFile string) error {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, text, "", "L", false)

	err := pdf.OutputFileAndClose(outputFile)
	if err != nil {

		return fmt.Errorf("не удалось создать PDF: %w", err)
	}
	return nil
}
