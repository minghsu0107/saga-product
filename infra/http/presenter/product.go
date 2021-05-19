package presenter

// Product payload
type Product struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	BrandName   string `json:"brand_name" binding:"required"`
	Price       int64  `json:"price" binding:"required"`
	Inventory   int64  `json:"inventory" binding:"required"`
}

// ProductCatalogs response payload
type ProductCatalogs struct {
	Catalogs []Productcatalog `json:"catalogs"`
}

// Product catalog payload
type Productcatalog struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Inventory int64  `json:"inventory"`
	Price     int64  `json:"price"`
}

// Pagination payload
type Pagination struct {
	Offset int `form:"offset" binding:"numeric,min=0"`
	Size   int `form:"size" binding:"required,numeric,min=1,max=500"`
}
