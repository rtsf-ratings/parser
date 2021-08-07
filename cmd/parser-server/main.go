package main

import (
    "mime/multipart"
    "net/http"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/rtsf-ratings/parser"
    "github.com/rtsf-ratings/parser/fast"
)

type Form struct {
    Type string                `form:"type" binding:"required"`
    File *multipart.FileHeader `form:"file" binding:"required"`
}

func main() {
    router := gin.Default()
    router.Use(cors.Default())

    router.POST("/upload", func(c *gin.Context) {
        var form Form

        if err := c.ShouldBind(&form); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var fileParser parser.Parser

        switch form.Type {
        case "fast":
            fileParser = &fast.FastParser{}
            break

        default:
            c.JSON(http.StatusBadRequest, gin.H{"error": "unknown file type"})
            return
        }

        openedFile, err := form.File.Open()
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        defer openedFile.Close()

        result, err := fileParser.Parse(openedFile, form.File.Size)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.IndentedJSON(http.StatusOK, result)
    })

    router.Run("127.0.0.1:8082")
}
