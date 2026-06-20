package controllers

import "meow.net/models"

func InitRoomSlices(rooms []models.Room) {
	for i := range rooms {
		if rooms[i].Roles == nil {
			rooms[i].Roles = []models.RoomRoleEntry{}
		}
		if rooms[i].Tags == nil {
			rooms[i].Tags = []models.RoomTag{}
		}
		if rooms[i].SubRooms == nil {
			rooms[i].SubRooms = []models.SubRoom{}
		}
		if rooms[i].LoadScreens == nil {
			rooms[i].LoadScreens = []interface{}{}
		}
		if rooms[i].PromoImages == nil {
			rooms[i].PromoImages = []interface{}{}
		}
		if rooms[i].PromoExternalContent == nil {
			rooms[i].PromoExternalContent = []interface{}{}
		}
	}
}
