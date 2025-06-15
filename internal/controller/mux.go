package controller

import (
	"encoding/json"
	"github.com/folivorra/studyDir/tree/develop/goRedis/internal/logger"
	"github.com/folivorra/studyDir/tree/develop/goRedis/internal/model"
	"github.com/folivorra/studyDir/tree/develop/goRedis/internal/storage"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type ItemController struct {
	Store storage.Storager
}

func NewItemController(store storage.Storager) *ItemController {
	return &ItemController{Store: store}
}

func (i *ItemController) AddItem(w http.ResponseWriter, r *http.Request) {
	var item model.Item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.ErrorLogger.Println("AddItem:", err)
		return
	}

	if item.Name == "" || item.ID <= 0 || item.Price < 0 {
		http.Error(w, "Invalid item data", http.StatusBadRequest)
		logger.ErrorLogger.Println("AddItem: invalid item data")
		return
	}

	if err := i.Store.AddItem(item); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		logger.ErrorLogger.Println("AddItem:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorLogger.Println("AddItem:", err)
		return
	}
	logger.InfoLogger.Println("AddItem successfully")
}

func (i *ItemController) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		logger.ErrorLogger.Println("DeleteItem: invalid ID")
		return
	}

	if err = i.Store.DeleteItem(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		logger.ErrorLogger.Println("DeleteItem:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	logger.InfoLogger.Println("DeleteItem successfully")
}

func (i *ItemController) GetAllItems(w http.ResponseWriter, r *http.Request) {
	items, err := i.Store.GetAllItems()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorLogger.Println("GetAllItems:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorLogger.Println("GetAllItems:", err)
		return
	}
	logger.InfoLogger.Println("GetAllItems successfully")
}

func (i *ItemController) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		logger.ErrorLogger.Println("GetItem: invalid ID")
		return
	}

	item, err := i.Store.GetItem(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		logger.ErrorLogger.Println("GetItem:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorLogger.Println("GetItem:", err)
		return
	}
	logger.InfoLogger.Println("GetItem successfully")
}

func (i *ItemController) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var item model.Item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.ErrorLogger.Println("UpdateItem:", err)
		return
	}

	if item.Name == "" || item.Price < 0 {
		http.Error(w, "Invalid item data", http.StatusBadRequest)
		logger.ErrorLogger.Println("UpdateItem: invalid item data")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID in URL", http.StatusBadRequest)
		logger.ErrorLogger.Println("UpdateItem: invalid ID in URL")
		return
	}

	item.ID = id

	// если id в json и в url не совпадают - приоритет отдается url-значению

	if err = i.Store.UpdateItem(item); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		logger.ErrorLogger.Println("UpdateItem:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorLogger.Println("UpdateItem:", err)
		return
	}
	logger.InfoLogger.Println("UpdateItem successfully")
}

func (i *ItemController) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/items", i.AddItem).Methods("POST")
	r.HandleFunc("/items", i.GetAllItems).Methods("GET")
	r.HandleFunc("/items/{id}", i.GetItem).Methods("GET")
	r.HandleFunc("/items/{id}", i.UpdateItem).Methods("PUT")
	r.HandleFunc("/items/{id}", i.DeleteItem).Methods("DELETE")
}
