package location

import "winx-profile/pkg/geoip"

type PointResource struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Resource struct {
	IP              string        `json:"ip"`
	City            string        `json:"city"`
	Country         string        `json:"country"`
	CurrentLocation PointResource `json:"current_location"`
}

func NewResource(item geoip.Result) *Resource {
	return &Resource{
		IP:      item.IP,
		City:    item.City,
		Country: item.Country,
		CurrentLocation: PointResource{
			Latitude:  item.Latitude,
			Longitude: item.Longitude,
		},
	}
}
