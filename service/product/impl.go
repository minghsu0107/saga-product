package product

import "github.com/minghsu0107/saga-product/repo"

// ProductServiceImpl implementation
type ProductServiceImpl struct {
	productRepo repo.ProductRepository
}
