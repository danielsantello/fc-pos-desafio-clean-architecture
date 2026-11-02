package usecase

import "github.com/danielsantello/fc-pos-desafio-clean-architecture/internal/entity"

type ListOrdersUseCase struct {
	OrderRepository entity.OrderRepositoryInterface
}

func NewListOrdersUseCase(repository entity.OrderRepositoryInterface) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		OrderRepository: repository,
	}
}

func (u *ListOrdersUseCase) Execute() ([]OrderOutputDTO, error) {
	orders, err := u.OrderRepository.List()
	if err != nil {
		return nil, err
	}

	output := make([]OrderOutputDTO, 0, len(orders))

	for _, order := range orders {
		output = append(output, OrderOutputDTO{
			ID:         order.ID,
			Price:      order.Price,
			Tax:        order.Tax,
			FinalPrice: order.FinalPrice,
		})
	}

	return output, nil
}
