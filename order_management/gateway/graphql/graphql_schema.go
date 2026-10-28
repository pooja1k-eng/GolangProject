package graphql

import (
	"errors"

	orderModel "order-management/order/model"
	orderService "order-management/order/service"
	productClient "order-management/product/client"

	"github.com/graphql-go/graphql"
)

func BuildGraphQLSchema(oSvc orderService.Service, pClient productClient.Service) (graphql.Schema, error) {
	// Silence compiler if pClient is kept for interface signatures
	_ = pClient

	orderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id":         &graphql.Field{Type: graphql.Int},
			"customerId": &graphql.Field{Type: graphql.Int},
			"productId":  &graphql.Field{Type: graphql.Int},
			"quantity":   &graphql.Field{Type: graphql.Int},
			"unitPrice":  &graphql.Field{Type: graphql.Float},
			"totalPrice": &graphql.Field{Type: graphql.Float},
			"status":     &graphql.Field{Type: graphql.String},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"orders": &graphql.Field{
				Type: graphql.NewList(orderType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return oSvc.ListOrders(p.Context, 100, 0)
				},
			},
			"order": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idVal, ok := p.Args["id"].(int)
					if !ok {
						return nil, errors.New("invalid order id")
					}
					return oSvc.GetOrder(p.Context, uint(idVal))
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createOrder": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"customerId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"productId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"quantity":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					req := &orderModel.CreateRequest{
						CustomerID: uint(p.Args["customerId"].(int)),
						ProductID:  uint(p.Args["productId"].(int)),
						Quantity:   p.Args["quantity"].(int),
					}
					return oSvc.CreateOrder(p.Context, req)
				},
			},
			"updateOrderStatus": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"id":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"status": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := uint(p.Args["id"].(int))
					statusStr := p.Args["status"].(string)

					// Cast raw string to custom orderModel.Status type
					if err := oSvc.UpdateOrderStatus(p.Context, id, orderModel.Status(statusStr)); err != nil {
						return nil, err
					}
					return oSvc.GetOrder(p.Context, id)
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
}