module github.com/tvitcom/go-workreport

replace github.com/tvitcom/go-workreport/mstime => ./mstime

replace github.com/tvitcom/go-workreport/models => ./models

go 1.16

require (
	github.com/gin-gonic/gin v1.7.7
	github.com/tealeg/xlsx v1.0.5
	github.com/tvitcom/go-workreport/mstime v0.0.0-00010101000000-000000000000
	github.com/xuri/excelize/v2 v2.5.0
)
