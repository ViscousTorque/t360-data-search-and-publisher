package models

type VehicleData struct {
	VRM               string `json:"vrm"`
	ContraventionDate string `json:"contravention_date"`
	IsHirerVehicle    bool   `json:"is_hirer_vehicle"`
	LeaseCompany      any    `json:"lease_company"`
}
