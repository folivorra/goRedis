package grpc

import (
	"context"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/model"
	"github.com/folivorra/goRedis/internal/storage"
	goredis_v1 "github.com/folivorra/goRedis/pkg/proto/goredis/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ItemController struct {
	goredis_v1.UnimplementedGoRedisServiceServer
	Store storage.Storager
}

func NewItemController(store storage.Storager) *ItemController {
	return &ItemController{Store: store}
}

func (i *ItemController) GetItem(_ context.Context, r *goredis_v1.GetItemRequest) (*goredis_v1.GetItemResponse, error) {
	if r.Id <= 0 {
		logger.ErrorLogger.Println("GetItem: invalid ID")
		return nil, status.Error(codes.InvalidArgument, "Invalid ID")
	}

	item, err := i.Store.GetItem(r.Id)
	if err != nil {
		logger.ErrorLogger.Printf("GetItem: %v", err)
		return nil, err
	}

	logger.InfoLogger.Println("GetItem successfully")

	return &goredis_v1.GetItemResponse{
		Item: &goredis_v1.Item{
			Id:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		},
	}, nil
}

func (i *ItemController) CreateItem(_ context.Context, r *goredis_v1.CreateItemRequest) (*goredis_v1.CreateItemResponse, error) {
	if r.Item.Name == "" || r.Item.Id <= 0 || r.Item.Price < 0 {
		logger.ErrorLogger.Println("CreateItem: invalid item data")
		return nil, status.Error(codes.InvalidArgument, "Invalid item data")
	}

	item := model.Item{
		ID:    r.Item.Id,
		Name:  r.Item.Name,
		Price: r.Item.Price,
	}

	if err := i.Store.CreateItem(item); err != nil {
		logger.ErrorLogger.Printf("CreateItem: %v", err)
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	logger.InfoLogger.Println("CreateItem successfully")
	return &goredis_v1.CreateItemResponse{
		Item: &goredis_v1.Item{
			Id:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		},
	}, nil
}

func (i *ItemController) UpdateItem(_ context.Context, r *goredis_v1.UpdateItemRequest) (*goredis_v1.UpdateItemResponse, error) {
	if r.Item.Name == "" || r.Item.Id <= 0 || r.Item.Price < 0 {
		logger.ErrorLogger.Println("UpdateItem: invalid item data")
		return nil, status.Error(codes.InvalidArgument, "Invalid item data")
	}

	item := model.Item{
		ID:    r.Item.Id,
		Name:  r.Item.Name,
		Price: r.Item.Price,
	}

	if err := i.Store.UpdateItem(item); err != nil {
		logger.ErrorLogger.Printf("UpdateItem: %v", err)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	logger.InfoLogger.Println("UpdateItem successfully")
	return &goredis_v1.UpdateItemResponse{
		Item: &goredis_v1.Item{
			Id:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		},
	}, nil
}

func (i *ItemController) DeleteItem(_ context.Context, r *goredis_v1.DeleteItemRequest) (*goredis_v1.DeleteItemResponse, error) {
	if r.Id <= 0 {
		logger.ErrorLogger.Println("DeleteItem: invalid ID")
		return nil, status.Error(codes.InvalidArgument, "Invalid ID")
	}

	if err := i.Store.DeleteItem(r.Id); err != nil {
		logger.ErrorLogger.Printf("DeleteItem: %v", err)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	logger.InfoLogger.Println("DeleteItem successfully")

	return &goredis_v1.DeleteItemResponse{
		Empty: &emptypb.Empty{},
	}, nil
}

func (i *ItemController) GetAllItems(_ context.Context, r *goredis_v1.GetAllItemsRequest) (*goredis_v1.GetAllItemsResponse, error) {
	items, err := i.Store.GetAllItems()
	if err != nil {
		logger.ErrorLogger.Printf("GetAllItems: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	var responseItems []*goredis_v1.Item
	for _, item := range items {
		responseItems = append(responseItems, &goredis_v1.Item{
			Id:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		})
	}

	logger.InfoLogger.Println("GetAllItems successfully")

	return &goredis_v1.GetAllItemsResponse{
		Items: responseItems,
	}, nil
}
