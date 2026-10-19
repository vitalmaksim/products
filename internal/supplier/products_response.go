package supplier

type ProductsResponse struct {
	Page       int          `json:"page"`
	TotalPages int          `json:"total_pages"`
	Items      []ProductDTO `json:"items"`
}
